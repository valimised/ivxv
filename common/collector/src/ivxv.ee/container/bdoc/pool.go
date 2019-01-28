package bdoc

import (
	"bytes"
	"sync"
)

// Shared pool of buffers used by other functions in this package. Great care
// must be taken to not retain any references to the buffer after releasing it
// for reuse, especially to any byte slices returned by Buffer.Bytes().
var buffers = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func buffer() *bytes.Buffer {
	return buffers.Get().(*bytes.Buffer)
}

func release(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		buffers.Put(buf)
	}
}
