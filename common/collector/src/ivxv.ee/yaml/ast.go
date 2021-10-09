package yaml

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Node is a node in the YAML AST. It is of concrete type Scalar, Sequence, or
// Mapping.
type Node interface {
	// apply sets v to the value of the node.
	apply(v reflect.Value)

	// equal compares a Node to another Node of the same type, including
	// the path. Only used for testing.
	equal(other Node) bool
}

// apply calls n.apply(v), but gracefully handles cases where n is nil.
func apply(n Node, v reflect.Value) {
	if n == nil {
		v = indirect(v)
		v.Set(reflect.Zero(v.Type()))
		return
	}
	n.apply(v)
}

// equal calls n.equal(other), but gracefully handles cases where n is nil.
func equal(n, other Node) bool {
	if n == nil {
		return other == nil
	}
	return n.equal(other)
}

// Scalar represents a YAML scalar.
type Scalar struct {
	path  string
	value string
}

// apply sets the value of a string, byte slice, int, or bool to the scalar.
func (s Scalar) apply(v reflect.Value) {
	v = indirect(v)
	switch v.Kind() {
	case reflect.String:
		v.SetString(s.String())
	case reflect.Int64:
		i, err := strconv.ParseInt(s.String(), 10, 64)
		if err != nil {
			s.error(InvalidInt64Error{Value: s, Err: err})
		}
		v.SetInt(i)
	case reflect.Uint64:
		u, err := strconv.ParseUint(s.String(), 10, 64)
		if err != nil {
			s.error(InvalidUint64Error{Value: s, Err: err})
		}
		v.SetUint(u)
	case reflect.Bool:
		switch s.String() {
		case "true":
			v.SetBool(true)
		case "false":
			v.SetBool(false)
		default:
			s.error(InvalidBoolError{Value: s})
		}
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			b, err := base64.StdEncoding.DecodeString(s.String())
			if err != nil {
				s.error(InvalidBase64Error{Value: s, Err: err})
			}
			v.Set(reflect.ValueOf(b))
			break
		}
		fallthrough
	default:
		if !reflect.TypeOf(s).AssignableTo(v.Type()) {
			s.error(InvalidScalarError{Type: v.Type()})
		}
		v.Set(reflect.ValueOf(s))
	}
}

func (s Scalar) equal(other Node) bool {
	o, ok := other.(Scalar)
	return ok && s.String() == o.String() && s.path == o.path
}

func (s Scalar) String() string {
	return s.value
}

func (s Scalar) error(err error) {
	stop(ApplyScalarError{Path: s.path, Err: err})
}

// Sequence represents a YAML sequence.
type Sequence struct {
	path     string
	elements []Node
}

// apply sets the elements of a slice or array.
func (s Sequence) apply(v reflect.Value) {
	v = indirect(v)
	switch v.Kind() {
	case reflect.Slice:
		if v.Cap() >= len(s.elements) {
			v.SetLen(len(s.elements))
		} else {
			v.Set(reflect.MakeSlice(v.Type(), len(s.elements), len(s.elements)))
		}
	case reflect.Array:
		if v.Len() != len(s.elements) {
			s.error(InvalidArrayLengthError{Len: v.Len(), Expected: len(s.elements)})
		}
	default:
		if !reflect.TypeOf(s).AssignableTo(v.Type()) {
			s.error(InvalidSliceError{Type: v.Type()})
		}
		v.Set(reflect.ValueOf(s))
		return
	}

	for i, n := range s.elements {
		e := reflect.New(v.Type().Elem()).Elem()
		apply(n, e)
		v.Index(i).Set(e)
	}
}

func (s Sequence) equal(other Node) bool {
	o, ok := other.(Sequence)
	if !ok || len(s.elements) != len(o.elements) || s.path != o.path {
		return false
	}
	if s.elements == nil || o.elements == nil {
		return s.elements == nil && o.elements == nil
	}
	for i, v := range o.elements {
		if !equal(s.elements[i], v) {
			return false
		}
	}
	return true
}

func (s Sequence) error(err error) {
	stop(ApplySequenceError{Path: s.path, Err: err})
}

// Mapping represents a YAML mapping.
type Mapping struct {
	path  string
	pairs map[string]Node
}

// apply sets the fields of a struct or entries of a map.
func (m Mapping) apply(v reflect.Value) {
	v = indirect(v)
	switch v.Kind() {
	case reflect.Struct:
		for k, n := range m.pairs {
			var f reflect.Value

			// Try an exact match first.
			if exported(k) {
				f = v.FieldByName(k)
			}

			// Otherwise try case-insensitive comparison.
			if !f.IsValid() {
				lk := strings.ToLower(k)
				f = v.FieldByNameFunc(func(name string) bool {
					return exported(name) && strings.ToLower(name) == lk
				})
			}
			if !f.IsValid() {
				// Ignore unspecified fields.
				continue
			}
			e := reflect.New(f.Type()).Elem()
			apply(n, e)
			f.Set(e)
		}
	case reflect.Map:
		ktype := v.Type().Key()
		if !reflect.TypeOf("").ConvertibleTo(ktype) {
			m.error(KeyNotConvertibleToStringError{Type: ktype})
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
		for k, n := range m.pairs {
			e := reflect.New(v.Type().Elem()).Elem()
			apply(n, e)
			v.SetMapIndex(reflect.ValueOf(k).Convert(ktype), e)
		}
	default:
		if !reflect.TypeOf(m).AssignableTo(v.Type()) {
			m.error(InvalidMappingError{Type: v.Type()})
		}
		v.Set(reflect.ValueOf(m))
	}
}

func (m Mapping) equal(other Node) bool {
	o, ok := other.(Mapping)
	if !ok || len(m.pairs) != len(o.pairs) || m.path != o.path {
		return false
	}
	if m.pairs == nil || o.pairs == nil {
		return m.pairs == nil && o.pairs == nil
	}
	for k, v := range o.pairs {
		if !equal(m.pairs[k], v) {
			return false
		}
	}
	return true
}

func (m Mapping) error(err error) {
	stop(ApplyMappingError{Path: m.path, Err: err})
}

// indirect dereferences p until it finds a non-pointer. If any nil values are
// encountered on the way, new pointers are allocated.
func indirect(p reflect.Value) (v reflect.Value) {
	v = p
	for {
		if v.Kind() != reflect.Ptr {
			return
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
}

// exported checks if the variable name is exported.
func exported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
