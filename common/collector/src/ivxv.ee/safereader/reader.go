/*
Package safereader provides an implementation for safely reading byte slices
from streams up to a maximum size.
*/
package safereader // import "ivxv.ee/safereader"

import (
	"fmt"
	"io"
	"sync"
)

// SafeReader allows safely reading byte slices from streams of unknown length
// by limiting the number of bytes allowed.
//
// It essentially combines io.LimitReader, ioutil.ReadAll, and sync.Pool. It
// reads from a stream until EOF is encountered or a predefined limit is
// exceeded, and returns the data read. A free list is kept of byte slices, so
// that if they are recovered after use, they can be reused. Because of this, a
// SafeReader should be allocated and reused instead of constantly creating new
// short-lived SafeReaders.
//
// SafeReader is safe (and encouraged) for concurrent use.
type SafeReader struct {
	max uint64

	// pool will contain byte slices with length max + 1. This way we can
	// read data and check for extra bytes in one shot.
	pool sync.Pool

	// The LimitExceededError returned from Read will be identical for an
	// instance of SafeReader so create it once and reuse.
	err LimitExceededError
}

// New creates a new SafeReader which allows reading up to max bytes.
func New(max uint64) (s *SafeReader) {
	s = &SafeReader{
		max: max,
		pool: sync.Pool{New: func() interface{} {
			b := make([]byte, max+1)
			return &b
		}},
	}
	s.err.Limit = max
	return
}

// LimitExceededError is returned when reading from a stream with more data
// than the specified limit.
var _ = LimitExceededError{Limit: 0}

// Read reads from r until EOF and returns the data read. If the length of data
// exceeds the maximum specified at creation, LimitExceededError is returned --
// in this case maximum + 1 bytes are read from r. If a non-EOF error is return
// by r, then it is returned as is, i.e., not wrapped in another error.
//
// Unlike Read methods in the standard library, if err is not nil, then b is
// nil, not the bytes read before the error occurred. We do not want to process
// partial data.
//
// When the returned byte slice is no longer needed, then it should be returned
// to the SafeReader via Recover for reuse.
func (s *SafeReader) Read(r io.Reader) (b []byte, err error) {
	b = *(s.pool.Get().(*[]byte))

	n, err := io.ReadFull(r, b)
	switch err {
	case io.EOF, io.ErrUnexpectedEOF:
		// Less than max + 1 bytes read: set the length of b to the
		// number of bytes read and reset err.
		b = b[:n]
		err = nil
	case nil:
		// max + 1 data read: the limit has been exceeded.
		err = s.err
	}

	if err != nil {
		s.Recover(b)
		b = nil
	}
	return
}

// Recover recovers a byte slice returned by Read for reuse by the SafeReader.
//
// Note! The data referenced by b must not be used by the caller after Recover.
// It can be overwritten at any moment.
func (s *SafeReader) Recover(b []byte) {
	// Allow users to grow the byte slices, but not shrink.
	if uint64(cap(b)) <= s.max {
		panic(fmt.Sprintf("byte slice too small to recover: cap %d, want %d", cap(b), s.max+1))
	}
	b = b[:s.max+1]
	s.pool.Put(&b)
}
