package safereader

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

// data is the stream of data being read in tests.
const data = "0123456789abcdef"

func reader(n uint64) io.Reader {
	return strings.NewReader(data[:n])
}

func read(t *testing.T, s *SafeReader, r io.Reader) []byte {
	b, err := s.Read(r)
	if err != nil {
		t.Fatal("read error:", err)
	}
	return b
}

func TestReadLimit(t *testing.T) {
	tests := []struct {
		max uint64
		ns  []uint64
	}{
		{0, []uint64{0, 1}},
		{10, []uint64{0, 5, 10, 11}},
	}

	for _, test := range tests {
		s := New(test.max)
		for _, n := range test.ns {
			t.Run(fmt.Sprintf("%d of %d", n, test.max), func(t *testing.T) {
				b, err := s.Read(reader(n))

				if uint64(n) > s.max {
					if err != s.err {
						t.Errorf("unexpected Read error: got %v, want %v",
							err, s.err)
					}
					return
				}

				if err != nil {
					t.Fatal("read error:", err)
				}
				if want := data[:n]; string(b) != want {
					t.Errorf("unexpected data: got %q, want %q", string(b), want)
				}
			})
		}
	}
}

func TestReadAll(t *testing.T) {
	s := New(uint64(len(data)))

	tests := []struct {
		name string
		r    io.Reader
	}{
		{"Reader", reader(s.max)},
		{"DataErrReader", iotest.DataErrReader(reader(s.max))},
		{"HalfReader", iotest.HalfReader(reader(s.max))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if b := string(read(t, s, test.r)); b != data {
				t.Errorf("unexpected data: got %q, want %q", b, data)
			}
		})
	}
}

func TestReadError(t *testing.T) {
	s := New(uint64(len(data)))

	// Ad-hoc solution for a reader which returns an error.
	r := iotest.TimeoutReader(iotest.HalfReader(reader(s.max)))
	if _, err := s.Read(r); err != iotest.ErrTimeout {
		t.Errorf("unexpected Read error: got %v, want %v", err, iotest.ErrTimeout)
	}
}

func TestRecover(t *testing.T) {
	s := New(uint64(len(data)))

	t.Run("half", func(t *testing.T) {
		s.Recover(read(t, s, reader(s.max/2)))
	})

	t.Run("exact", func(t *testing.T) {
		s.Recover(read(t, s, reader(s.max)))
	})

	t.Run("grown", func(t *testing.T) {
		b := read(t, s, reader(s.max))
		b = append(b, 0)
		s.Recover(b)
	})
}
