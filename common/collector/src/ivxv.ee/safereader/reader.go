/*
Package safereader implements safely reading from streams up to a maximum size.
*/
package safereader // import "ivxv.ee/safereader"

import (
	"io"
)

// LimitExceededError is returned when reading from a stream with more data
// than the specified limit.
var _ = LimitExceededError{Limit: 0}

type safereader struct {
	r io.Reader
	n int64

	// The LimitExceededError returned from Read will be identical for an
	// instance of safereader so create it once and reuse.
	err LimitExceededError
}

// New returns a Reader that reads from r but returns a LimitExceededError
// after n bytes are read. This behavior is extremely similar to
// io.LimitedReader which does the same but returns io.EOF. With safereader it
// is possible to distinguish if there were more bytes in r or it returned
// io.EOF after exactly n bytes.
//
// Due to the implementation, if more than n bytes are requested, the returned
// Reader reads n+1 bytes from r. Callers must keep this in mind if they wish
// to continue using r after having depleted the returned Reader.
func New(r io.Reader, n int64) io.Reader {
	s := &safereader{r: r, n: n}
	s.err.Limit = n
	return s
}

func (s *safereader) Read(p []byte) (n int, err error) {
	if s.n < 0 {
		// Limit was exceeded by previous calls to Read.
		return 0, s.err
	}

	// If more than n bytes are requested, then attempt to read one extra
	// byte. This is allowed by io.Reader which says that implementations
	// can use all of p as scratch space.
	if int64(len(p)) > s.n {
		p = p[:s.n+1]
	}
	n, err = s.r.Read(p)

	if int64(n) > s.n {
		// Limit has been exceeded.
		n = int(s.n) // Only report s.n bytes as safe to use. Downcast is safe since n > s.n.
		err = s.err  // Replace underlying error with a LimitExceededError.
		s.n = -1     // Mark n as exceeded.
		return
	}

	s.n -= int64(n)
	return
}
