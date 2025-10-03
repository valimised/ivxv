// Package internal provides group internal low-level functionalities.
package internal

import (
	"errors"

	"tivi.io/core/math/group"
)

// CastGroupTo casts group.Group to T.
func CastGroupTo[T any](G group.Group) (T, error) {
	var g T
	var ok bool
	if g, ok = G.(T); !ok {
		return g, errors.New("group/internal: cannot cast group.Group to T type")
	}
	return g, nil
}

// CastElementTo casts group.Element to T.
func CastElementTo[T any](E group.Element) (T, error) {
	var e T
	var ok bool
	if e, ok = E.(T); !ok {
		return e, errors.New("group/internal: cannot cast group.Element to T type")
	}
	return e, nil
}
