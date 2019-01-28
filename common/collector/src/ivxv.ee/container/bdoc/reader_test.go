package bdoc

import (
	"bytes"
	"io"
	"testing"
)

// testReader only implements io.Reader.
type testReader bytes.Reader

func newReader(data []byte) *testReader {
	return (*testReader)(bytes.NewReader(data))
}

func (r *testReader) Read(p []byte) (int, error) {
	return (*bytes.Reader)(r).Read(p)
}

// testReaderAt only implements io.Reader and io.ReaderAt.
type testReaderAt struct{ *testReader }

func newReaderAt(data []byte) testReaderAt {
	return testReaderAt{newReader(data)}
}

func (r *testReaderAt) ReadAt(p []byte, off int64) (int, error) {
	return (*bytes.Reader)(r.testReader).ReadAt(p, off)
}

// testReadAtSeeker only implements io.Reader, io.ReaderAt and io.Seeker.
type testReadAtSeeker struct{ testReaderAt }

func (r testReadAtSeeker) Seek(offset int64, whence int) (int64, error) {
	return (*bytes.Reader)(r.testReader).Seek(offset, whence)
}

// testReadAtSizer only implements io.Reader, io.ReaderAt and sizer.
type testReadAtSizer struct{ testReaderAt }

func (r testReadAtSizer) Size() int64 {
	return (*bytes.Reader)(r.testReader).Size()
}

var data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

func TestToReaderAt(t *testing.T) {
	tests := []struct {
		name   string
		reader io.Reader
	}{
		{"Reader", newReader(data)},
		{"ReaderAt", newReaderAt(data)},
		{"ReadAtSeeker", testReadAtSeeker{newReaderAt(data)}},
		{"ReadAtSizer", testReadAtSizer{newReaderAt(data)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ratc, size, err := toReaderAt(test.reader, int64(len(data)))
			if err != nil {
				t.Fatal("toReaderAt error:", err)
			}
			ratc.close()
			if size != int64(len(data)) {
				t.Fatalf("unexpected size: want %d, got %d", len(data), size)
			}
		})
	}
}
