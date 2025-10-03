// Package ecqp implements prime-order subgroup of following NIST elliptic curves:
//
//	NIST-P256
//	NIST-P384
//	NIST-P521
package ecqp

import (
	"crypto/elliptic"
	"fmt"
	"math/big"

	"tivi.io/core/math/group/internal"
)

// init registers ECqP group in a group.registry.
func init() {
	registerWellKnownGroups()
}

const (
	// logPrefix is a prefix for logging records.
	logPrefix = "group/ecqp"
)

// shortWeierstrassFunc is a short Weierstrass equation that solves for Y^2, given X coordinate
// and has a form of
//
//	y^2 = x^3 + Ax + B
//
// where A and B are constants.
func shortWeierstrassFunc(x *big.Int, curve elliptic.Curve) *big.Int {
	int3 := big.NewInt(3) //nolint:mnd

	// x^3
	x3 := new(big.Int).Exp(x, int3, curve.Params().P)

	// A = P - 3
	A := new(big.Int).Sub(curve.Params().P, int3)

	// Ax
	A.Mul(A, x)

	// x^3 + Ax
	x3.Add(x3, A)

	// (x^3 + Ax + B) mod P
	x3.Add(x3, curve.Params().B)
	x3.Mod(x3, curve.Params().P)

	// y^2
	return x3
}

// unmarshal is a copy of elliptic.Unmarshal with an ability to distinguish
// between invalid, uncompressed, infinity and IsOnCurve point errors.
// Returns x, y, nil if point is valid (infinity point is also considered valid).
func unmarshal(point []byte, curve elliptic.Curve) (x, y *big.Int, err error) { //nolint:nonamedreturns
	fieldByteLen := (curve.Params().BitSize + 7) / 8 //nolint:mnd

	// Uncompressed point slice is 50/50 <--> X/Y coordinates
	if len(point) != 1+2*fieldByteLen {
		return nil, nil, fmt.Errorf("%s: invalid elliptic curve point format", logPrefix)
	}

	// Is point in uncompressed form?
	if point[0] != 0x04 { //nolint:mnd
		return nil, nil, fmt.Errorf("%s: not an uncompressed elliptic curve point", logPrefix)
	}

	p := curve.Params().P
	x = new(big.Int).SetBytes(point[1 : 1+fieldByteLen])
	y = new(big.Int).SetBytes(point[1+fieldByteLen:])

	// Point outside a curve
	if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
		return nil, nil, fmt.Errorf("%s: point outside an elliptic curve", logPrefix)
	}

	if err := isOnCurve(curve, x, y); err != nil {
		return nil, nil, err
	}

	return x, y, nil
}

// marshal is a copy of elliptic.Marshal with an ability to distinguish
// between invalid, uncompressed, infinity and IsOnCurve point errors.
// Returns marshalled elliptic curve point.
func marshal(curve elliptic.Curve, x, y *big.Int) ([]byte, error) {
	fieldByteLen := (curve.Params().BitSize + 7) / 8 //nolint:mnd

	if err := isOnCurve(curve, x, y); err != nil {
		return nil, err
	}

	// 1 byte for uncompressed format byte and each coordinate has a size of field order
	point := make([]byte, 1+2*fieldByteLen)
	point[0] = 0x04 // uncompressed point

	// Fill slice 50/50 <--> X/Y coordinates
	x.FillBytes(point[1 : 1+fieldByteLen])
	y.FillBytes(point[1+fieldByteLen : 1+2*fieldByteLen])

	return point, nil
}

// isOnCurve returns nil if (x,y) on an elliptic curve.
// (0,0) is considered to be point at infinity.
func isOnCurve(curve elliptic.Curve, x, y *big.Int) error {
	// Check point is at infinity (0, 0)
	if isAtInfinity(x, y) {
		return nil
	}

	// Point at infinity is rejected (0,0)
	if !curve.IsOnCurve(x, y) {
		return internal.NotOnCurve{Prefix: logPrefix}
	}

	return nil
}

// isAtInfinity returns true if point is at infinity, i.e. (0,0).
func isAtInfinity(x, y *big.Int) bool {
	if x.Cmp(new(big.Int)) == 0 && y.Cmp(new(big.Int)) == 0 {
		return true
	}

	return false
}
