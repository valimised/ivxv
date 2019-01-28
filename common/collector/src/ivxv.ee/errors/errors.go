/*
Package errors provides interfaces and methods for working with hierarchical
errors.
*/
package errors // import "ivxv.ee/errors"

import "reflect"

// Causer wraps the basic Cause method.
//
// All errors which were caused by another error should implement this
// interface and the Cause method should return the causing error.
//
// The returned error must be non-nil. If the implementing error was not caused
// by another error, then it should not implement Causer at all.
type Causer interface {
	Cause() error
}

// Walk walks the causes of err, calling f on each error, including err itself.
// It stops once it reaches an error which does not implement Causer. If f
// returns a non-nil error, then Walk cancels and returns that error, otherwise
// nil. Walk does nothing if err is nil.
func Walk(err error, f func(error) error) error {
	var ret error
	for ret == nil && err != nil {
		ret = f(err)
		c, ok := err.(Causer)
		if !ok {
			break
		}
		err = c.Cause()
	}
	return ret
}

// CausedBy checks if the error itself is of the given type or if it was caused
// by an error of that type. It returns the top-most error that is of the
// requested type or nil. As a conveinience, pointers to error types also match
// the pointed error types so that CausedBy(Error{}, new(Error)) returns a
// non-nil result.
func CausedBy(err error, v interface{}) (found error) {
	if v == nil {
		return nil
	}
	needle := reflect.TypeOf(v)
	for needle.Kind() == reflect.Ptr {
		needle = needle.Elem()
	}

	Walk(err, func(err error) error { // nolint: errcheck, gosec, only found or nil is returned.
		hay := reflect.TypeOf(err)
		for hay.Kind() == reflect.Ptr {
			hay = hay.Elem()
		}
		if hay == needle {
			found = err
			return found // Just to cancel the walk.
		}
		return nil
	})
	return
}
