package bdoc

import (
	"bytes"
	"io"

	"ivxv.ee/safereader"
)

// readAtCloser is a wrapper around the io.ReaderAt returned by toReaderAt
// which on close releases any resources allocated by it.
type readAtCloser struct {
	io.ReaderAt
	buffer *bytes.Buffer
}

func (ratc *readAtCloser) close() {
	release(ratc.buffer)
	ratc.buffer = nil
	ratc.ReaderAt = nil // Reader references ratc.buffer.Bytes() so set nil.
}

// toReaderAt transforms an io.Reader into an io.ReaderAt. This is achieved by
// first checking whether the io.Reader already implements io.ReaderAt. If not,
// r is read until EOF or the safe limit is reached and the read data is
// returned in a bytes.Reader.
//
// toReaderAt also finds the size of the data to be read, either by calling
// Size or Seek on r or again just by reading until EOF or the safe limit is
// reached. An error is returned if the size is greater than limit.
//
// The resulting io.ReaderAt is wrapped in a readAtCloser, whose Close method
// frees any resources acquired during this process. Therefore the Close method
// of ratc MUST be called when it is no longer used.
func toReaderAt(r io.Reader, limit int64) (ratc *readAtCloser, size int64, err error) {
	// Convert r into an io.ReaderAt.
	var ok bool
	ratc = new(readAtCloser)
	ratc.ReaderAt, ok = r.(io.ReaderAt)
	if !ok {
		// r is not an io.ReaderAt: read everything into memory and
		// wrap in a bytes.Reader.
		ratc.buffer = buffer()
		if size, err = ratc.buffer.ReadFrom(safereader.New(r, limit)); err != nil {
			ratc.close()
			err = ReaderAtReadBufferError{Err: err}
			return
		}
		ratc.ReaderAt = bytes.NewReader(ratc.buffer.Bytes())
		return // We have both ratc and size.
	}

	// Determine the size of ratc.ReaderAt.
	type sizer interface {
		Size() int64
	}

	switch s := ratc.ReaderAt.(type) {
	case sizer:
		size = s.Size()
	case io.Seeker:
		// Seek to end to get length and return to current location.
		var current int64
		if current, err = s.Seek(0, io.SeekCurrent); err != nil {
			err = ReaderAtSeekCurrentError{Err: err}
			return
		}
		if size, err = s.Seek(0, io.SeekEnd); err != nil {
			err = ReaderAtSeekEndError{Err: err}
			return
		}
		if _, err = s.Seek(current, io.SeekStart); err != nil {
			err = ReaderAtSeekStartError{Err: err}
			return
		}
	default:
		// As a fallback, just read everything into memory, determine
		// the size, and discard the read data. This alters the read
		// offset, but ReadAt does not depend on it anyway.
		buf := buffer()
		defer release(buf)
		if size, err = buf.ReadFrom(safereader.New(r, limit)); err != nil {
			err = ReaderAtReadSizeError{Err: err}
			return
		}
	}

	if size > limit {
		err = ReaderAtSizeError{Size: size, Limit: limit}
	}
	return
}
