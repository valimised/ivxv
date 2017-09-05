// Package declared tests that gen does not match already declared types.
package declared

type Declared struct {
	Field int
}

var x = Declared{Field: 0}
