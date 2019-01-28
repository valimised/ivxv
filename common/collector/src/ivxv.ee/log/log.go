/*
Package log is the logging framework used by collector services.
*/
package log // import "ivxv.ee/log"

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/syslog"
	"net/url"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"ivxv.ee/cryptoutil"
	"ivxv.ee/errors"
)

// facility is the the syslog facility to log with.
const facility = syslog.LOG_LOCAL0

// maxlen is the maximum byte-length of a formatted log entries field. If a
// field exceeds this, then it will be truncated.
//
// maxlen is a variable and not a constant so that it can be overrriden
// by tests.
var maxlen = 4096

// key is the key type of values stored in context by this package.
type key int

const (
	loggerKey key = iota // Context key for logger instance.
	connIDKey            // Context key for connection ID.
	sessIDKey            // Context key for session ID.
)

type logger struct {
	debug uint32           // Is debug logging enabled?
	w     writer           // Destination to log formatted entries to.
	now   func() time.Time // Function used to get current time.
}

// writer provides an interface in front of syslog.Writer, so that it can be
// switched out for testing.
type writer interface {
	Debug(m string) error
	Info(m string) error
	Err(m string) error
	Alert(m string) error
	Close() error
}

type noopWriter struct{}

func (w noopWriter) Debug(m string) error { return nil }
func (w noopWriter) Info(m string) error  { return nil }
func (w noopWriter) Err(m string) error   { return nil }
func (w noopWriter) Alert(m string) error { return nil }
func (w noopWriter) Close() error         { return nil }

// NewContext adds a new logger into the context. The returned context can then
// be later passed to one of the logging functions, which use the logger from
// it.
func NewContext(ctx context.Context, tag string) (context.Context, error) {
	// No need to set default severity level, as we will not be using it.
	w, err := syslog.New(facility, tag)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, loggerKey, &logger{0, w, time.Now}), nil
}

// TestContext adds a logger to the context, which does not log anything. This
// is meant to be used in tests, which call logging functions, but where we do
// not want to log anything.
func TestContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, &logger{0, noopWriter{}, time.Now})
}

// SetDebug sets debugging in the logger in the context.
func SetDebug(ctx context.Context, debug bool) {
	var v uint32
	if debug {
		v = 1
	}
	atomic.StoreUint32(&fromctx(ctx).debug, v)
}

// Close closes the logger in the context.
func Close(ctx context.Context) error {
	return fromctx(ctx).w.Close()
}

// WithConnectionID sets the connection ID used when logging with the returned
// context.
func WithConnectionID(ctx context.Context, cid string) context.Context {
	return context.WithValue(ctx, connIDKey, cid)
}

// WithSessionID sets the session ID used when logging with the returned
// context.
func WithSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, sessIDKey, sid)
}

func fromctx(ctx context.Context) *logger {
	// fromctx will return nil if used on a context which was not acquired
	// through NewContext, but this is a programmer error which will panic
	// on the first logging function call and will always be detected, so
	// we are okay with it.
	return ctx.Value(loggerKey).(*logger)
}

// Entry declares the method used to identify a log entry.
type Entry interface {
	EntryID() string
}

// ErrorEntry is an Entry which is also an error.
type ErrorEntry interface {
	Entry
	error
}

// Sensitive is used for logging sensitive byte slices. The value of Sensitive
// is never logged directly, only its length and hash.
//
// NB! Logging sensitive data from a limited message space will not protect the
// data: only use Sensitive for data whose hashes cannot be brute-forced or
// precomputed.
type Sensitive []byte

// Debug logs a debug entry, which will not be sent to external monitoring and
// can be disabled. See Log for details about the format of the logged data.
func Debug(ctx context.Context, entry Entry) {
	l := fromctx(ctx)
	if atomic.LoadUint32(&l.debug) == 1 {
		l.w.Debug(format(ctx, entry)) // nolint: errcheck, gosec, ignore logging errors.
	}
}

// Log logs an entry. The entry will be formatted as a JSON object with five
// members.
//
//     - Timestamp    an ISO8601 timestamp of the time the entry was logged,
//     - ConnectionID an unique network connection identifier, only used in
//                    request handlers,
//     - SessionID    an unique session identifier, only used in request
//                    handlers after a session ID has been established,
//     - ID           an unique log entry identifier, each ID must be used only
//                    once in the codebase, and
//     - Entry        a JSON encoding of the log entry.
//
// The log entry will be encoded to JSON according to the following rules:
//
//     - nested values which implement the Entry interface will be encoded as a
//       JSON object with two members: "ID", the unique identifier of the
//       nested entry, and "Entry", the encoding of the nested entry according
//       to these rules, excluding any interface checking,
//
//     - values of type Sensitive will be encoded as a JSON object with two
//       members: "Hash", a SHA-256 hash of the value, and "Length", the length
//       of the original value,
//
//     - values of type time.Time will be formatted according to RFC 3339 with
//       either second (if the nanosecond part is zero) or nanosecond precision,
//
//     - values of type pkix.Name and pkix.RDNSequence will be encoded to a
//       string according to RFC 4514,
//
//     - values of type *x509.Certificate will be encoded as a JSON object with
//       members containing the certificate's subject name, issuer name,
//       hexadecimal serial number, and SHA-256 hash of its DER encoding,
//
//     - values which implement the error interface will be converted to a
//       string using Error() and processed as a string value,
//
//     - values which implement the fmt.Stringer interface will be converted to
//       a string using String() and processed as a string value.
//
//     - byte slices will be Base64-encoded to a string and processed as a
//       string value,
//
//     - if a string value exceeds the maximum length, a JSON object will be
//       encoded in its place with two members: "Truncated", the string value
//       truncated to the maximum length, and "FullLength", the byte length of
//       the original value,
//
//     - string values (including truncated ones) will be URL-encoded according
//       to RFC 3986,
//
//     - values of kind bool, int, int8, int16, int32, int64, uint, uint8,
//       uint16, uint32, uint64, float32, and float64 will be encoded as JSON
//       booleans or numbers,
//
//     - slices and arrays will be encoded as JSON arrays with these rules
//       applied to each value,
//
//     - maps with key strings will be encoded as JSON objects with these rules
//       applied to each name and value, with the exception that if a name
//       exceeds the maximum length it will not be truncated, but cause an
//       error,
//
//     - structs will be encoded as JSON objects with field names as member
//       names and these rules applied to each value, and
//
//     - all other values are unsupported and will cause an error.
//
// As noted, when encoding a nested entry, interface checks are skipped. First,
// this avoids recursing into the first rule, and second, we wish to ignore the
// Error and String methods of log entries to get the actual structure of the
// entry and not just a string representation.
//
// In case an error occurs during log entry encoding a plaintext string will be
// logged instead, containing the time, connection ID, session ID, entry ID,
// and the error that occurred. The unencoded entry will not be included.
func Log(ctx context.Context, entry Entry) {
	fromctx(ctx).w.Info(format(ctx, entry)) // nolint: errcheck, gosec, ignore logging errors.
}

// Error logs an error entry. By default, the entry is logged with severity
// "error", which indicates an error in the normal flow of the application.
// However, if the entry was caused by an error wrapped with Alert, then
// severity "alert" is used instead, indicating an error not directly tied to
// the application, but of a higher-level, e.g., when an external service is
// not reachable or writing to disk fails.
//
// See Log for details about the format of the logged data.
func Error(ctx context.Context, entry ErrorEntry) {
	l := fromctx(ctx)
	f := l.w.Err
	if errors.CausedBy(entry, new(alert)) != nil {
		f = l.w.Alert
	}
	f(format(ctx, entry)) // nolint: errcheck, gosec, ignore logging errors.
}

type alert struct {
	err error
}

func (a alert) Cause() error  { return a.err }
func (a alert) Error() string { return a.err.Error() }

// Alert wraps err, marking that it should be logged with severity "alert".
func Alert(err error) error {
	return alert{err}
}

type message struct {
	Timestamp    string
	ConnectionID string `json:",omitempty"`
	SessionID    string `json:",omitempty"`
	ID           string
	Entry        *json.RawMessage `json:",omitempty"`
}

// format formats the entry to a string as described by Log.
func format(ctx context.Context, entry Entry) string {
	now := fromctx(ctx).now().Format(time.RFC3339Nano)

	var cid, sid string
	if val := ctx.Value(connIDKey); val != nil {
		cid = val.(string)
	}
	if val := ctx.Value(sessIDKey); val != nil {
		sid = val.(string)
	}

	// Get an encoder from the pool and encode the log entry.
	e := encPool.Get().(*encoder)
	err := e.encode(reflect.ValueOf(entry), false)
	if err != nil {
		err = fmt.Errorf("failed to encode: %v", err)
	}

	var b []byte
	if err == nil {
		// On success, marshal the entire log entry.
		encoded := e.Bytes()
		m := (*json.RawMessage)(&encoded)

		// Omit empty entries for readability.
		if empty(encoded) {
			m = nil
		}
		b, err = json.Marshal(message{now, cid, sid, entry.EntryID(), m})
	}

	// Put the encoder back into the pool regardless of success and only
	// after the entry has been marshaled. Otherwise we introduce a race
	// condition where another goroutine can get the same encoder from the
	// pool and overwrite the data in its buffer.
	e.Reset()
	encPool.Put(e)

	if err != nil {
		// This should never happen, but we do not want to panic here
		// if it does: log a plain error. Avoid logging the the entry
		// since it may contain unencoded information.
		return fmt.Sprintf("failed to format log message: "+
			"time=%s, cid=%s, sid=%s, id=%s, err=%v",
			now, cid, sid, entry.EntryID(), err)
	}
	return string(b)
}

type encoder struct {
	bytes.Buffer
	*json.Encoder
}

var encPool = sync.Pool{New: func() interface{} {
	e := new(encoder)
	e.Encoder = json.NewEncoder(e)
	e.SetEscapeHTML(false)
	return e
}}

// encode traverses v and encodes it to JSON according to the rules described
// in Log. If nest is false, then checking if v implements Entry will be
// skipped: used when we are already nesting v and now want to encode its
// value.
func (e *encoder) encode(v reflect.Value, nest bool) (err error) {
	// Check for specific types.
	switch t := v.Interface().(type) {
	case alert:
		// Do not log alert, which is only used for tagging entries.
		return e.encode(reflect.ValueOf(t.Cause()), nest)
	case Entry:
		if nest {
			return e.encodeNested(t)
		}
	case Sensitive:
		return e.encodeSensitive(t)
	case time.Time:
		layout := time.RFC3339
		if t.Nanosecond() > 0 {
			layout = time.RFC3339Nano
		}
		return e.encodeString(t.Format(layout))
	case pkix.RDNSequence:
		return e.encodeRDNSequence(t)
	case pkix.Name:
		t.ExtraNames = t.Names // Include unrecognized names.
		return e.encodeRDNSequence(t.ToRDNSequence())
	case *x509.Certificate:
		return e.encodeCertificate(t)
	case error:
		return e.encodeString(t.Error())
	case fmt.Stringer:
		return e.encodeString(t.String())
	}

	// Check for supported kinds.
	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:

		// Numbers have a pre-defined format and limited length, so
		// encode unmodified.
		return e.Encode(v.Interface())

	case reflect.String:
		return e.encodeString(v.String())

	case reflect.Slice:
		// Check if we are encoding a byte slice: encode to Base64.
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return e.encodeBase64(v.Bytes())
		}
		fallthrough
	case reflect.Array:
		return e.encodeArray(v)

	case reflect.Map:
		return e.encodeMap(v)

	case reflect.Struct:
		return e.encodeStruct(v)

	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return e.Encode(nil)
		}
		return e.encode(v.Elem(), nest)

	default:
		// Although the encoding/json package can handle more cases
		// than listed here (e.g., any type that implements
		// json.Marshaler), we do not support them as log entries.
		return fmt.Errorf("unsupported log entry type: %s", v.Type())
	}
}

// encodeSensitive encodes a value of type Sensitive into a JSON object as
// described in Log.
func (e *encoder) encodeSensitive(sensitive Sensitive) (err error) {
	h := sha256.Sum256(sensitive)
	return e.Encode(struct {
		Hash   []byte
		Length int
	}{
		h[:],
		len(sensitive),
	})
}

// encodeRDNSequence encodes a value of type pkix.RDNSequence into a string.
func (e *encoder) encodeRDNSequence(rdns pkix.RDNSequence) (err error) {
	s, err := cryptoutil.EncodeRDNSequence(rdns)
	if err != nil {
		return fmt.Errorf("failed to encode RDN sequence: %v", err)
	}
	return e.encodeString(s)
}

// encodeCertificate encodes a value of type *x509.Certificate into a JSON
// object as described in Log.
func (e *encoder) encodeCertificate(c *x509.Certificate) (err error) {
	if c == nil {
		return e.Encode(nil)
	}

	hash := sha256.Sum256(c.Raw)
	return e.encodeStruct(reflect.ValueOf(struct {
		Subject pkix.Name
		Issuer  pkix.Name
		Serial  string
		Hash    []byte
	}{
		Subject: c.Subject,
		Issuer:  c.Issuer,
		Serial:  c.SerialNumber.Text(16),
		Hash:    hash[:],
	}))
}

// encodeString encodes a string into a JSON value as described in Log.
func (e *encoder) encodeString(s string) (err error) {
	if len(s) > maxlen {
		return e.Encode(struct {
			Truncated  string
			FullLength int
		}{
			url.QueryEscape(s[:maxlen]),
			len(s),
		})
	}

	return e.Encode(url.QueryEscape(s))
}

// encodeBase64 encodes a byte slice into Base64. It does this manually instead
// of using encoding/json, because a type alias of []byte with a MarshalJSON
// method could override this behavior.
//
// nolint: errcheck, gosec, writing to bytes.Buffer always returns nil and
// base64.Encoder only returns errors when writing to the underlying stream
// fails.
func (e *encoder) encodeBase64(data []byte) (err error) {
	e.WriteByte('"')
	b := base64.NewEncoder(base64.StdEncoding, e)
	b.Write(data)
	b.Close()
	e.WriteByte('"')
	return
}

// encodeArray encodes an array or slice into a JSON array with encoded values.
func (e *encoder) encodeArray(v reflect.Value) (err error) {
	e.WriteByte('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			e.WriteByte(',')
		}
		if err = e.encode(v.Index(i), true); err != nil {
			return
		}
	}
	e.WriteByte(']')
	return
}

// encodeMap encodes a map with string keys into a JSON object with encoded
// names and values.
func (e *encoder) encodeMap(v reflect.Value) (err error) {
	// Allow only string keys. encoding/json package also allows numbers
	// and encoding.TextMarshalers, but we have no need for that.
	if v.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("unsupported log entry map key type: %s", v.Type().Key())
	}

	e.WriteByte('{')
	for i, key := range v.MapKeys() {
		if i > 0 {
			e.WriteByte(',')
		}
		name := key.String()
		if len(name) > maxlen {
			return fmt.Errorf("map key too long: %d", len(name))
		}
		if err = e.encodeString(name); err != nil {
			return
		}
		e.WriteByte(':')
		if err = e.encode(v.MapIndex(key), true); err != nil {
			return
		}
	}
	e.WriteByte('}')
	return
}

// encodeStruct encodes a struct into a JSON object with encoded values.
func (e *encoder) encodeStruct(v reflect.Value) (err error) {
	e.WriteByte('{')
	for i := 0; i < v.NumField(); i++ {
		if i > 0 {
			e.WriteByte(',')
		}
		if err = e.Encode(v.Type().Field(i).Name); err != nil {
			return
		}
		e.WriteByte(':')
		if err = e.encode(v.Field(i), true); err != nil {
			return
		}
	}
	e.WriteByte('}')
	return
}

// encodeNested encodes a nested log entry into a JSON object as described in
// Log.
func (e *encoder) encodeNested(n Entry) (err error) {
	e.WriteString(`{"ID":`)
	if err = e.Encode(n.EntryID()); err != nil {
		return
	}
	afterID := e.Len()

	e.WriteString(`,"Entry":`)
	beforeEntry := e.Len()

	if err = e.encode(reflect.ValueOf(n), false); err != nil {
		return
	}

	// In case we encoded an empty entry, then just output the ID.
	encoded := e.Bytes()[beforeEntry:]
	if empty(encoded) {
		e.Truncate(afterID)
	}

	e.WriteByte('}')
	return
}

// empty checks if the JSON encoding in b corresponds to a null value or empty
// object.
func empty(b []byte) bool {
	// json.Encoder.Encode adds a newline after each encode, which is why
	// null will have a newline at the end.
	s := string(b)
	return s == "null\n" || s == "{}"
}
