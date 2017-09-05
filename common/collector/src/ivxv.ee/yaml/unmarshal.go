package yaml

import (
	"io"
	"reflect"
)

type yamlError struct {
	err error
}

// stop panics with a yamlError to stop the current operation.
//
// Instead of having to return an error through many layers of recursive
// function calls, use a panic to unwind the stack until stopped is deferred.
func stop(err error) {
	panic(yamlError{err})
}

// stopped checks if the current operation was stopped and returns the error if
// it did. Otherwise returns the provided error, so that it does not get
// overwritten if no panic happened. Only useful inside a deferred function
// like this:
//
//     defer func() {
//         err = stopped(err, recover())
//     }()
//
func stopped(err error, recovered interface{}) error {
	if recovered != nil {
		yerr, ok := recovered.(yamlError)
		if !ok {
			panic(recovered)
		}
		err = yerr.err
	}
	return err
}

// Apply sets the value of v to that of the YAML Node n. n must be assignable
// to v, with the following exceptions:
//
//     - Scalar nodes can be applied to strings, byte slices (in which case the
//       content is automatically base64-decoded), int64s, uint64s, and
//       booleans,
//     - Sequence nodes can be applied to slices and arrays, and
//     - Mapping nodes can be applied to maps with string keys and structs (the
//       mapping keys must must case-insensitively match fields in the struct).
//
func Apply(n Node, v interface{}) (err error) {
	defer func() {
		err = stopped(err, recover())
	}()

	r := reflect.ValueOf(v)
	if r.Kind() != reflect.Ptr || r.IsNil() {
		return ApplyInvalidTargetError{Type: r.Type()}
	}
	apply(n, r)
	return
}

// Unmarshal calls Parse on the YAML-serialized data and calls Apply to apply
// the returned root Node to v.
//
// If data contains container references, then the data in container is used.
func Unmarshal(data io.Reader, container map[string][]byte, v interface{}) (err error) {
	root, err := Parse(data, container)
	if err != nil {
		return
	}
	return Apply(root, v)
}
