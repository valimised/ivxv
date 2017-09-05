package yaml

import (
	"os"
	"reflect"
	"testing"
)

// unmarshal is the test struct to unmarshal into.
type unmarshal struct {
	String string
	Int    int64
	Bool   bool

	Slice []string
	Array [3]string

	Map    map[string]string
	Struct struct {
		Foo string
		Baz string
	}

	Introspection Node

	Nested []struct {
		Key   string
		Hello string
	}
}

func compare(t *testing.T, name string, have, want interface{}) bool {
	if have != want {
		t.Errorf(`unexpected %s: "%v", want "%v"`, name, have, want)
		return false
	}
	return true
}

func gets(m map[string]string, k string) (s string) {
	s, ok := m[k]
	if !ok {
		s = "<no such key>"
	}
	return
}

func TestUnmarshal(t *testing.T) {
	f, err := os.Open("testdata/unmarshal.yaml")
	if err != nil {
		t.Fatal("failed to open test YAML:", err)
	}

	var v unmarshal
	if err := Unmarshal(f, nil, &v); err != nil {
		t.Fatal("failed to unmarshal test YAML:", err)
	}

	compare(t, "string", v.String, "hello, world!")
	compare(t, "int", v.Int, int64(-123))
	compare(t, "bool", v.Bool, true)

	if compare(t, "slice length", len(v.Slice), 3) {
		compare(t, "slice[0]", v.Slice[0], "first")
		compare(t, "slice[1]", v.Slice[1], "second")
		compare(t, "slice[2]", v.Slice[2], "third")
	}
	compare(t, "array[0]", v.Array[0], "fourth")
	compare(t, "array[1]", v.Array[1], "fifth")
	compare(t, "array[2]", v.Array[2], "sixth")

	compare(t, "map.key", gets(v.Map, "key"), "value")
	compare(t, "map.hello", gets(v.Map, "hello"), "world!")
	compare(t, "struct.foo", v.Struct.Foo, "bar")
	compare(t, "struct.baz", v.Struct.Baz, "qux")

	if compare(t, "introspection type", reflect.TypeOf(v.Introspection),
		reflect.TypeOf(Mapping{})) {

		m := v.Introspection.(Mapping)
		compare(t, "introspection.scalar",
			m.pairs["scalar"].(Scalar).String(), "hello, world!")

		sn := m.pairs["sequence"]
		if compare(t, "introspection.sequence type", reflect.TypeOf(sn),
			reflect.TypeOf(Sequence{})) {

			s := sn.(Sequence)
			if compare(t, "introspection.sequence length", len(s.elements), 2) {
				compare(t, "introspection.sequence[0]",
					s.elements[0].(Scalar).String(), "first")
				compare(t, "introspection.sequence[1]",
					s.elements[1].(Scalar).String(), "second")
			}
		}
	}

	if compare(t, "nested length", len(v.Nested), 2) {
		compare(t, "nested[0].key", v.Nested[0].Key, "value")
		compare(t, "nested[0].hello", v.Nested[0].Hello, "world!")
		compare(t, "nested[1].key", v.Nested[1].Key, "foobar")
		compare(t, "nested[1].hello", v.Nested[1].Hello, "maailm!")
	}
}
