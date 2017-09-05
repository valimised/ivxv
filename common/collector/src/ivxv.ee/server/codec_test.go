package server

import (
	"reflect"
	"testing"

	"ivxv.ee/errors"
)

func TestCheckSize(t *testing.T) {
	type Nested struct {
		String string `size:"6"`
	}

	tests := []struct {
		name     string
		x        interface{}
		expected error
	}{
		{"simple", struct {
			String  string `size:"6"`
			Bytes   []byte `size:"4"`
			Ignored int64  `size:"1"`
		}{
			"foobar",
			[]byte{0xde, 0xad, 0xbe, 0xef},
			256,
		}, nil},

		{"nested", struct {
			Nested Nested
		}{
			Nested{"foobar"},
		}, nil},

		{"simple too big", struct {
			String string `size:"4"`
		}{
			"foobar",
		}, new(FieldSizeError)},

		{"simple pointer too big", struct {
			Bytes *[]byte `size:"2"`
		}{
			&[]byte{0xde, 0xad, 0xbe, 0xef},
		}, new(FieldSizeError)},

		{"nested too big", struct {
			Nested Nested
		}{
			Nested{"foobarbaz"},
		}, new(NestedFieldSizeError)},

		{"nested pointer too big", struct {
			Nested *Nested
		}{
			&Nested{"foobarbaz"},
		}, new(NestedFieldSizeError)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkSize(reflect.ValueOf(test.x))
			if err != test.expected && errors.CausedBy(err, test.expected) == nil {
				t.Errorf("unexpected error: got %v, want cause %T", err, test.expected)
			}
		})
	}
}
