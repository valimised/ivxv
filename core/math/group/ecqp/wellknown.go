package ecqp

import (
	"crypto/elliptic"

	"tivi.io/core/math/group"
)

const fieldEncodingBits = 10

// registerWellKnownGroups registers popular NIST curve implementations in a group.registry.
// Names of NIST curve implementations are defined by TIVI Core library, in order to keep
// backwards compatibility with elliptic library.
func registerWellKnownGroups() {
	curves := map[string]elliptic.Curve{
		"NIST-P256": elliptic.P256(),
		"NIST-P384": elliptic.P384(),
		"NIST-P521": elliptic.P521(),
	}

	// Register curves in group.registry
	for name, curve := range curves {
		group.Register(&ecqPGroup{name: name, Curve: curve, fieldEncodingBits: fieldEncodingBits})
	}
}
