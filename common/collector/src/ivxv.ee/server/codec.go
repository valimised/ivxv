package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"strconv"
	"strings"
	"time"

	"ivxv.ee/log"
)

// serverCodec is a rpc.ServerCodec that passes everything to the jsonrpc
// codec, but also sets reading and writing deadlines, injects the given
// context into the header of the request, and passes the header through
// provided filters.
type serverCodec struct {
	rpc.ServerCodec
	header  *Header
	conn    net.Conn
	timeout time.Duration
	filters headerFilters
}

func newCodec(ctx context.Context, conn net.Conn, timeout time.Duration,
	filters headerFilters) *serverCodec {

	return &serverCodec{
		jsonrpc.NewServerCodec(conn),
		&Header{Ctx: ctx},
		conn,
		timeout,
		filters,
	}
}

var errIgnored = errors.New("ignored")

// ReadRequestHeader sets a read timeout for the underlying connection and then
// calls ReadRequestHeader of the underlying codec.
//
// Note: err will be ignored by the codecFilter, so any error information must
// be logged by ReadRequestHeader. err will not be sent to the client. However,
// a non-nil error must still be returned to stop processing the request any
// further.
func (s *serverCodec) ReadRequestHeader(req *rpc.Request) error {
	if err := s.conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
		log.Error(s.header.Ctx, SetReadDeadlineError{Err: log.Alert(err)})
		return errIgnored
	}
	if err := s.ServerCodec.ReadRequestHeader(req); err != nil {
		log.Error(s.header.Ctx, ReadJSONRequestError{Err: err})
		return errIgnored
	}
	return nil
}

// ReadRequestBody calls ReadRequestBody of the underlying codec, checks the
// size of all of the request's fields, injects the connection context into the
// embedded server.Header field, and passes the header through filters.
// ReadRequestBody will panic if x does not satisfy the header interface: this
// is by design to avoid developers forgetting to embed the server header in
// request types.
//
// No timeout is required for ReadRequestBody since with jsonrpc the whole
// request was actually read in ReadRequestHeader.
//
// Note: err will be ignored by the codecFilter and sent to the client by the
// rpc package, so it should be as generic as possible. Any error information
// should be logged by ReadRequestBody.
func (s *serverCodec) ReadRequestBody(x interface{}) error {
	if err := s.ServerCodec.ReadRequestBody(x); err != nil {
		log.Error(s.header.Ctx, UnmarshalRequestParamsError{
			RequestType: reflect.TypeOf(x),
			Err:         err,
		})
		return ErrBadRequest
	}

	// ReadRequestBody may be called with a nil argument to discard the
	// body if a correct header was read, but bad method was requested.
	if x == nil {
		return nil
	}

	if err := checkSize(reflect.ValueOf(x)); err != nil {
		log.Error(s.header.Ctx, RequestSizeError{
			RequestType: reflect.TypeOf(x),
			Err:         err,
		})
		return ErrBadRequest
	}

	// Inject the context into x's header, pass the latter through filters,
	// and store it for the response. The filters already log any relevant
	// error information and return a generic error, so just pass it
	// through without doing anything.
	h := x.(header).header()
	h.Ctx = s.header.Ctx
	err := s.filters.next(h)
	s.header = h
	return err
}

// checkSize walks over v looking for struct fields of type string or []byte
// and with the tag `size:"<int>"`. For each matching field, the field value's
// length is checked to be lesser than or equal to <int>.
func checkSize(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			t := v.Type().Field(i)
			f := v.Field(i)

		field:
			var n int
			switch t.Type.Kind() {
			case reflect.String:
				n = f.Len()
			case reflect.Slice:
				if t.Type.Elem().Kind() != reflect.Uint8 {
					continue
				}
				n = f.Len()
			case reflect.Struct:
				if err := checkSize(f); err != nil {
					return NestedFieldSizeError{Field: t.Name, Err: err}
				}
				continue
			case reflect.Ptr, reflect.Interface:
				if !f.IsNil() {
					t.Type = t.Type.Elem()
					f = f.Elem()
					goto field
				}
				continue
			default:
				continue
			}
			if tag := t.Tag.Get("size"); len(tag) > 0 {
				max, err := strconv.ParseUint(tag, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("invalid %s field size: %v", t.Name, err))
				}
				if uint64(n) > max {
					return FieldSizeError{
						Field: t.Name,
						Size:  n,
						Max:   max,
					}
				}
			}
		}
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			return checkSize(v.Elem())
		}
	}
	return nil
}

// WriteResponse injects the request's server header into the response, sets a
// write timeout for the underlying connection, and then calls WriteResponse of
// the underlying codec. WriteResponse will panic if x does not satisfy the
// header interface: this is by design to avoid developers forgetting to embed
// the server header in response types.
//
// Note: err will be ignored by the codecFilter and will not be sent to the
// client (as writing failed): it is meant for debugging the rpc package. Any
// errors from reading or processing the request were already logged by other
// components. Therefore WriteResponse is left with logging any errors related
// to writing or generated by the rpc package.
func (s *serverCodec) WriteResponse(resp *rpc.Response, x interface{}) error {
	// Log any rpc package errors and replace with generic ones.
	if strings.HasPrefix(resp.Error, "rpc: ") {
		log.Error(s.header.Ctx, RPCMethodError{ErrString: resp.Error})
		resp.Error = ErrBadRequest.Error()
	}

	// Set the header of the response unless we are returning an error.
	if len(resp.Error) == 0 {
		*x.(header).header() = *s.header
	}

	if err := s.conn.SetWriteDeadline(time.Now().Add(s.timeout)); err != nil {
		log.Error(s.header.Ctx, SetWriteDeadlineError{Err: log.Alert(err)})
		return errIgnored
	}
	if err := s.ServerCodec.WriteResponse(resp, x); err != nil {
		log.Error(s.header.Ctx, WriteResponseError{Err: err})
		return errIgnored
	}
	return nil
}
