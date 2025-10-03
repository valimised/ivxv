// Package duplicate tests that gen reports duplicate literals.
package duplicate

var (
	x = Duplicate{Field: 0, Description: "Duplicate"}
	y = Duplicate{Field: 0, Description: "Duplicate"}
)
