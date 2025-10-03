// Package local tests that gen does not match locally declared types.
package local

func x() {
	type InnerLocal struct {
		Field int
	}

	type OuterLocal struct {
		Field int
	}

	_ = InnerLocal{Field: 0}
}

var _ = OuterLocal{Field: 0}
