// Package modqp represents multiplicative subgroup of quadratic residues of following RFC3526 groups:
//
//	2048-bit
//	3072-bit
//	4096-bit
//	8192-bit
//
// Each subgroup contains only quadratic residue elements.
//
// For each subgroup, exponents (scalars) are in a range of [1, q), whereas elements are in [1, p) range.
package modqp

const (
	// logPrefix is a prefix for logging records.
	logPrefix = "group/modqp"
)

// init registers ModqP groups in a group.registry.
func init() {
	registerWellKnownGroups()
}
