package safereader_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"ivxv.ee/safereader"
)

// ParseBigEndian parses a big-endian encoded unsigned integer from r. Returns
// an error if r has more bytes than an uint64 can hold.
func ParseBigEndian(r io.Reader) (u uint64, err error) {
	b, err := ioutil.ReadAll(safereader.New(r, 8))
	if err != nil {
		return
	}

	for _, o := range b {
		u = u<<8 | uint64(o)
	}
	return
}

func Example() {
	u, err := ParseBigEndian(bytes.NewReader([]byte{0xde, 0xad, 0xbe, 0xef}))
	if err != nil {
		fmt.Println("error parsing 0xdeadbeef:", err)
		return
	}
	fmt.Println(u)

	_, err = ParseBigEndian(bytes.NewReader(make([]byte, 9)))
	if err != nil {
		fmt.Println("error parsing nine bytes:", err)
		return
	}

	// Output:
	// 3735928559
	// error parsing nine bytes: ivxv.ee/safereader.LimitExceededError{Limit:8}
}
