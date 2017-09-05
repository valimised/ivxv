package bdoc

import (
	"bytes"
	"io"

	"ivxv.ee/safereader"
)

// readAtCloser is a wrapper around the io.ReaderAt returned by toReaderAt
// which on close releases any resources allocated by it.
type readAtCloser struct {
	reader io.ReaderAt
	buf    []byte
	sr     *safereader.SafeReader
}

func (ratc readAtCloser) ReadAt(b []byte, off int64) (n int, err error) {
	return ratc.reader.ReadAt(b, off)
}

func (ratc *readAtCloser) close() {
	if ratc.buf != nil {
		ratc.sr.Recover(ratc.buf)
	}
}

// toReaderAt transforms an io.Reader into an io.ReaderAt. This is achieved by
// first checking whether the io.Reader already implements io.ReaderAt. If not,
// r is read until EOF and the read data is returned in a bytes.Reader.
//
// toReaderAt also finds the size of the data to be read, either by calling
// Size or Seek on r or again just by reading until EOF.
//
// sr is used as the pool of buffers for reading data and to ensure that we do
// not read more than expected.
//
// The resulting io.ReaderAt is wrapped in a readAtCloser, whose Close method
// frees any resources allocated during this process. Therefore the Close
// method of ratc MUST be called when it is no longer used.
func toReaderAt(r io.Reader, sr *safereader.SafeReader) (ratc *readAtCloser, size int64, err error) {
	type sizer interface {
		Size() int64
	}

	// Convert r into an io.ReaderAt.
	var ok bool
	ratc = new(readAtCloser)
	ratc.reader, ok = r.(io.ReaderAt)
	if !ok {
		// r is not an io.ReaderAt: read everything into memory and
		// wrap in a bytes.Reader.
		var buf []byte
		if buf, err = sr.Read(r); err != nil {
			err = ReaderAtReadBufError{Err: err}
			return
		}
		ratc.reader = bytes.NewReader(buf)
		ratc.buf = buf // Free the read buffer on close.
		ratc.sr = sr   // SafeReader to recover buf to.
	}

	// Determine the size of ratc.reader.
	switch s := ratc.reader.(type) {
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
		// the size, and discard the read data.
		//
		// bytes.Reader implements sizer so we do not have to worry
		// about reading everything into memory a second time if r was
		// not an io.ReaderAt.
		var buf []byte
		if buf, err = sr.Read(r); err != nil {
			err = ReaderAtReadSizeError{Err: err}
			return
		}
		size = int64(len(buf))
		sr.Recover(buf)
	}
	return
}
