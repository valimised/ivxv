package ecqp

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/cryptobyte"

	asn_1 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group"
	"tivi.io/core/math/group/internal"
)

type ecqPElement struct {
	g *ecqPGroup
	// x is an x coordinate of an elliptic curve.
	x *big.Int
	// y is a y coordinate of an elliptic curve.
	y *big.Int
}

// Inverse returns E2 that satisfies
//
//	E * E2 = 1
func (E *ecqPElement) Inverse() group.Element {
	E2 := E.g.identity()

	E2.x.Set(E.x)
	E2.y.Neg(E.y)
	E2.y.Mod(E2.y, E2.g.Params().P)

	return E2
}

// Scale returns
//
//	E * s
func (E *ecqPElement) Scale(s *group.Scalar) (group.Element, error) {
	E2 := E.g.identity()

	if s.Modulo().Cmp(E.g.Order()) != 0 {
		return nil, internal.ScalarModCmpOrder{Prefix: logPrefix}
	}

	// ScalarMult requires s to be in Big-Endian form
	x, y := E2.g.ScalarMult(E.x, E.y, s.Value().Bytes())

	E2.x.Set(x)
	E2.y.Set(y)

	return E2, nil
}

// Op returns
//
//	E + E2
func (E *ecqPElement) Op(E2 group.Element) (group.Element, error) {
	E3, err := internal.CastElementTo[*ecqPElement](E2)
	if err != nil {
		return nil, err
	}

	if err := E.g.Equal(E3.g); err != nil {
		return nil, err
	}

	E4 := E.g.identity()

	x, y := E4.g.Add(E.x, E.y, E3.x, E3.y)

	E4.x.Set(x)
	E4.y.Set(y)

	return E4, nil
}

// Equal return nil if E == E2.
func (E *ecqPElement) Equal(E2 group.Element) error {
	E3, err := internal.CastElementTo[*ecqPElement](E2)
	if err != nil {
		return err
	}

	if err := E.g.Equal(E3.g); err != nil {
		return err
	}

	if E.x.Cmp(E3.x) != 0 {
		return fmt.Errorf("%s: different elliptic curve point X coordinates", logPrefix)
	}

	if E.y.Cmp(E3.y) != 0 {
		return fmt.Errorf("%s: different elliptic curve point Y coordinates", logPrefix)
	}

	return nil
}

// Decode decodes a message from elliptic curve point X coordinate.
func (E *ecqPElement) Decode() ([]byte, error) {
	// Ensure that this group element actually belongs to the correct group.
	if err := E.g.IsGroupElement(E); err != nil {
		return nil, err
	}

	// Ensure that group element is padded up until elliptic curve field size.
	// -1 since padding head 0 bit is stripped by big.Int
	if E.x.BitLen() != E.g.FieldOrder().BitLen()-1 {
		return nil, internal.PaddedBitLen{Prefix: logPrefix, MsgBitLen: E.x.BitLen(), ExpectedBitLen: E.g.FieldOrder().BitLen() - 1}
	}

	// Strip field encoding bits
	return new(big.Int).Rsh(new(big.Int).Set(E.x), uint(E.g.fieldEncodingBits)).Bytes(), nil
}

// Marshal ASN.1 marshals group element as
//
//	E ::= OCTET STRING
//
// where OCTET STRING is uncompressed elliptic curve point.
func (E *ecqPElement) Marshal() (asn_1.DER, error) {
	point, err := marshal(E.g, E.x, E.y)
	if err != nil {
		return nil, err
	}

	var builder cryptobyte.Builder

	builder.AddASN1OctetString(point)

	return builder.Bytes() //nolint:wrapcheck
}
