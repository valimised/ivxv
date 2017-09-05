package cookie

import (
	"bytes"
	"testing"
)

// zero creates a new cookie manager with a zero key and zero nonce: this
// ensures predictable output.
func zero(t *testing.T) *C {
	c, err := New(make([]byte, 16))
	if err != nil {
		t.Fatal("new failed:", err)
	}
	c.nonce = make([]byte, len(c.nonce))
	return c
}

var (
	data   = []byte("foobar")
	cookie = []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x65, 0xe7, 0xb5, 0xac, 0x01, 0xc4, 0xe8, 0xc9, 0xd8, 0x17, 0x8d, 0x1a,
		0xfc, 0x1e, 0xe1, 0x19, 0xd8, 0x39, 0xfd, 0x94, 0x7c, 0x6d,
	}
)

func TestCreate(t *testing.T) {
	if got := zero(t).Create(data); !bytes.Equal(got, cookie) {
		t.Errorf("unexpected cookie value: %x", got)
	}
}

func TestOpen(t *testing.T) {
	if got, err := zero(t).Open(cookie); err != nil {
		t.Error("open failed:", err)
	} else if !bytes.Equal(got, data) {
		t.Errorf("unexpected data value: %q", got)
	}
}

func TestIncrementNonce(t *testing.T) {
	c := zero(t)
	empty := make([]byte, len(c.nonce))

	one := make([]byte, len(c.nonce))
	one[len(one)-1] = 1

	ff := make([]byte, len(c.nonce))
	ff[len(ff)-1] = 0xff

	carry := make([]byte, len(c.nonce))
	carry[len(carry)-2] = 1

	tests := []struct {
		name   string
		before []byte
		after  []byte
	}{
		{"one", empty, one},
		{"carry", ff, carry},
		{"wrap", bytes.Repeat([]byte{0xff}, len(c.nonce)), empty},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			copy(c.nonce, test.before)
			c.Create(nil)
			if !bytes.Equal(c.nonce, test.after) {
				t.Errorf("unexpected nonce: %x, want %x", c.nonce, test.after)
			}
		})
	}
}
