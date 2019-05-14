package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"ivxv.ee/log"
	"ivxv.ee/safereader"
)

// wrappedConn wraps net.Conn so its Read method can be replaced with another
// io.Reader.
type wrappedConn struct {
	net.Conn
	r io.Reader
}

func (w wrappedConn) Read(p []byte) (n int, err error) {
	return w.r.Read(p)
}

// serverCodec is a rpc.ServerCodec which wraps a jsonrpc codec with additional
// behavior:
//
//     - sets read and write deadlines,
//     - sets a maximum request size limit (if configured),
//     - logs requests to the request log (if configured),
//     - checks the size of all request fields tagged with "size",
//     - injects a context into the header of the request,
//     - applies filters to the header of the request, and
//     - sets the filtered header of the request as the header of the response.
//
type serverCodec struct {
	rpc.ServerCodec // Inner jsonrpc server codec.

	conn    wrappedConn // Wrapped connection served by this codec.
	timeout time.Duration
	logreq  bool

	header  *Header // Server header of the request and response.
	filters headerFilters
}

func newCodec(ctx context.Context, conf *CodecConf, conn net.Conn,
	filters headerFilters) *serverCodec {

	codec := &serverCodec{
		conn:    wrappedConn{Conn: conn, r: conn},
		timeout: time.Duration(conf.RWTimeout) * time.Second,
		logreq:  conf.LogRequests,
		header:  &Header{Ctx: ctx},
		filters: filters,
	}
	if conf.RequestSize > 0 {
		codec.conn.r = safereader.New(conn, conf.RequestSize)
	}
	// Must be reference so that changes to codec.conn affect inner codec.
	codec.ServerCodec = jsonrpc.NewServerCodec(&codec.conn)
	return codec
}

var errIgnored = errors.New("ignored")

var buffers = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}

// ReadRequestHeader sets a read deadline for the connection and then calls
// ReadRequestHeader of the underlying codec, logging the request if
// configured.
//
// Note: err will be ignored by the codecFilter, so any error information must
// be logged by ReadRequestHeader. err will not be sent to the client. However,
// a non-nil error must still be returned to stop processing the request any
// further.
func (s *serverCodec) ReadRequestHeader(req *rpc.Request) error {
	if s.logreq {
		tee := buffers.Get().(*bytes.Buffer)
		defer buffers.Put(tee)
		defer tee.Reset()
		s.conn.r = io.TeeReader(s.conn.r, tee)
		defer func() {
			if err := log.Request(s.header.Ctx, tee.Bytes()); err != nil {
				log.Error(s.header.Ctx, LogRequestError{Err: log.Alert(err)})
				// Do not block handling of request.
			}
		}()
	}

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
// write deadline for the connection, and then calls WriteResponse of the
// underlying codec. WriteResponse will panic if x does not satisfy the header
// interface: this is by design to avoid developers forgetting to embed the
// server header in response types.
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
