package yaml

import (
	"bytes"
	"reflect"
	"testing"

	"ivxv.ee/errors"
)

// Helper functions for creating Nodes without path information.
func scalar(value string) Scalar            { return Scalar{value: value} }
func sequence(elements ...Node) Sequence    { return Sequence{elements: elements} }
func mapping(pairs map[string]Node) Mapping { return Mapping{pairs: pairs} }

func expect(t *testing.T, cause error, recovered interface{}) {
	if err := stopped(nil, recovered); err == nil || errors.CausedBy(err, cause) == nil {
		t.Errorf("unexpected error: %v; want cause %T", err, cause)
	}
}

func TestScalar_Apply(t *testing.T) {
	t.Run("bool", func(t *testing.T) {
		var b bool
		v := reflect.ValueOf(&b).Elem()

		if scalar("true").apply(v); !b {
			t.Error("unexpected boolean value: false")
		}

		if scalar("false").apply(v); b {
			t.Error("unexpected boolean value: true")
		}

		defer func() {
			expect(t, new(InvalidBoolError), recover())
		}()
		scalar("garbage").apply(v)
	})

	t.Run("byte slice", func(t *testing.T) {
		var b []byte
		v := reflect.ValueOf(&b).Elem()

		if scalar("3q2+7w==").apply(v); !bytes.Equal(b, []byte{0xde, 0xad, 0xbe, 0xef}) {
			t.Error("unexpected byte slice value:", b)
		}

		defer func() {
			expect(t, new(InvalidBase64Error), recover())
		}()
		scalar("garbage").apply(v)
	})

	t.Run("invalid", func(t *testing.T) {
		var s []string
		v := reflect.ValueOf(&s).Elem()

		defer func() {
			expect(t, new(InvalidScalarError), recover())
		}()
		scalar("- hello, world!").apply(v)
	})

	t.Run("int64", func(t *testing.T) {
		var i int64
		v := reflect.ValueOf(&i).Elem()

		if scalar("-123").apply(v); i != -123 {
			t.Error("unexpected integer value:", i)
		}

		defer func() {
			expect(t, new(InvalidInt64Error), recover())
		}()
		scalar("garbage").apply(v)
	})

	t.Run("uint64", func(t *testing.T) {
		var u uint64
		v := reflect.ValueOf(&u).Elem()

		if scalar("456").apply(v); u != 456 {
			t.Error("unexpected unsigned integer value:", u)
		}

		defer func() {
			expect(t, new(InvalidUint64Error), recover())
		}()
		scalar("-789").apply(v)
	})

	t.Run("string", func(t *testing.T) {
		var s string
		v := reflect.ValueOf(&s).Elem()

		if scalar("hello, world!").apply(v); s != "hello, world!" {
			t.Error("unexpected string value:", s)
		}
	})
}

func TestSequence_Apply(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		var a [3]string
		v := reflect.ValueOf(&a).Elem()

		sequence(scalar("foo"), scalar("bar"), scalar("baz")).apply(v)
		if a[0] != "foo" || a[1] != "bar" || a[2] != "baz" {
			t.Error("unexpected array value:", a)
		}

		defer func() {
			expect(t, new(InvalidArrayLengthError), recover())
		}()
		sequence(scalar("hello"), scalar("world!")).apply(v)
	})

	t.Run("invalid", func(t *testing.T) {
		var s string
		v := reflect.ValueOf(&s).Elem()

		defer func() {
			expect(t, new(InvalidSliceError), recover())
		}()
		sequence(scalar("hello, world!")).apply(v)
	})

	t.Run("slice", func(t *testing.T) {
		var s []string
		v := reflect.ValueOf(&s).Elem()

		q := sequence(scalar("foo"), scalar("bar"), scalar("baz"))
		q.apply(v)
		if len(s) != len(q.elements) {
			t.Errorf("unexpected slice length: %d; want %d", len(s), len(q.elements))
		}
		if s[0] != "foo" || s[1] != "bar" || s[2] != "baz" {
			t.Error("unexpected slice value:", s)
		}

		// Check that we do not allocate a new slice if we have
		// enough cap already.
		c := cap(s)
		q = sequence(scalar("hello"), scalar("world!"))
		q.apply(v)
		if len(s) != len(q.elements) {
			t.Errorf("unexpected slice length: %d; want %d", len(s), len(q.elements))
		}
		if cap(s) != c {
			t.Errorf("unexpected slice capacity: %d; want %d", cap(s), c)
		}
		if s[0] != "hello" || s[1] != "world!" {
			t.Error("unexpected slice value:", s)
		}
	})
}

func TestMapping_Apply(t *testing.T) {
	t.Run("convertible", func(t *testing.T) {
		type convertible string
		var m map[convertible]string
		v := reflect.ValueOf(&m).Elem()

		// Just check if this panics or not.
		mapping(map[string]Node{"hello": scalar("world!")}).apply(v)
	})

	t.Run("invalid", func(t *testing.T) {
		var s string
		v := reflect.ValueOf(&s).Elem()

		defer func() {
			expect(t, new(InvalidMappingError), recover())
		}()
		mapping(map[string]Node{"hello": scalar("world!")}).apply(v)
	})

	t.Run("map", func(t *testing.T) {
		var m map[string]string
		v := reflect.ValueOf(&m).Elem()

		g := mapping(map[string]Node{"hello": scalar("world!"), "foo": scalar("bar")})
		g.apply(v)
		if len(m) != len(g.pairs) {
			t.Errorf("unexpected map size: %d; want %d", len(m), len(g.pairs))
		}
		for k, e := range g.pairs {
			if v, ok := m[k]; !ok {
				t.Errorf("missing key: %q", k)
			} else if v != e.(Scalar).String() {
				t.Errorf("unexpected value for key %q: %q; want %q", k, e, v)
			}
		}
	})

	t.Run("struct", func(t *testing.T) {
		var s struct {
			Foo     string
			Bar     string
			bar     string // To check that the exported value gets set, not this one.
			garbage string // To check that unexported values do not get set.
		}

		v := reflect.ValueOf(&s).Elem()

		g := mapping(map[string]Node{"Foo": scalar("hello"), "bar": scalar("world!")})
		g.apply(v)
		if s.Foo != "hello" || s.Bar != "world!" || len(s.bar) > 0 {
			t.Errorf("unexpected struct values: {Foo:%q Bar:%q bar:%q}; "+
				"want {Foo:%q Bar:%q bar:%q}",
				s.Foo, s.Bar, s.bar, g.pairs["Foo"], g.pairs["bar"], "")
		}

		// Check that extra fields are ignored.
		mapping(map[string]Node{"garbage": scalar("hello, world!")}).apply(v)
		if len(s.garbage) > 0 {
			t.Errorf("value assigned to unexported field: {garbage:%q}", s.garbage)
		}
	})
}
