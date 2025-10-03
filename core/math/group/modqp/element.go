package modqp

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/cryptobyte"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/math/group"
	"tivi.io/core/math/group/internal"
)

type modqPElement struct {
	g     *modqPGroup
	value *big.Int
}

// Inverse returns E2 that satisfies
//
//	E * E2 = 1
func (E *modqPElement) Inverse() group.Element { //nolint:revive
	E2 := E.g.identity()

	E2.value.ModInverse(E.value, E.g.fieldOrder)

	return E2
}

// Scale returns
//
//	E ^ s
func (E *modqPElement) Scale(s *group.Scalar) (group.Element, error) { //nolint:revive
	E2 := E.g.identity()

	if s.Modulo().Cmp(E.g.groupOrder) != 0 {
		return nil, internal.ScalarModCmpOrder{Prefix: logPrefix}
	}

	E2.value.Exp(E.value, s.Value(), E.g.fieldOrder)

	return E2, nil
}

// Op returns
//
//	E * E2
func (E *modqPElement) Op(E2 group.Element) (group.Element, error) {
	E3, err := internal.CastElementTo[*modqPElement](E2)
	if err != nil {
		return nil, err
	}

	err = E.g.Equal(E3.g)
	if err != nil {
		return nil, err
	}

	E4 := E.g.identity()
	E4.value.Mul(E.value, E3.value)
	E4.value.Mod(E4.value, E4.g.fieldOrder)

	return E4, nil
}

// Equal returns nil if E == E2.
func (E *modqPElement) Equal(E2 group.Element) error {
	E3, err := internal.CastElementTo[*modqPElement](E2)
	if err != nil {
		return err
	}

	if err = E.g.Equal(E3.g); err != nil {
		return err
	}

	if E.value.Cmp(E3.value) != 0 {
		return fmt.Errorf("%s: different element values", logPrefix)
	}

	return nil
}

// Decode decodes a message from group element.
func (E *modqPElement) Decode() (padded []byte, err error) { //nolint:revive
	// Ensure that this group element actually belongs to the correct group.
	if err := E.g.IsGroupElement(E); err != nil {
		return nil, err
	}

	// All group elements whose value > ModqP group generator order are considered to be quadratic non-residues
	decoded := new(big.Int).Set(E.value)
	if E.value.Cmp(E.g.Order()) > 0 {
		// Make a value to be a quadratic residue
		decoded.Sub(E.g.FieldOrder(), E.value)
	}

	msgByteLen := (decoded.BitLen() + 7) / 8
	maxByteLen := E.g.maxPaddedMsgByteLen() - 1 //((E.g.Order().BitLen() + 7) / 8) - 1
	// Ensure that group element is padded up until Sophie-Germain q value bytelen
	// -1 since padding header 0x00 byte is stripped by big.Int
	if msgByteLen != maxByteLen {
		return nil, internal.PaddedByteLen{Prefix: logPrefix, MsgByteLen: msgByteLen, ExpectedByteLen: maxByteLen}
	}

	padded = make([]byte, E.g.maxPaddedMsgByteLen())
	decoded.FillBytes(padded)
	return
}

// Marshal ASN.1 marshals group element as
//
//	E ::= INTEGER
//
// where INTEGER is group element value
func (E *modqPElement) Marshal() (der asn_1.DER, err error) { //nolint:revive
	var builder cryptobyte.Builder
	builder.AddASN1BigInt(E.value)

	return builder.Bytes()
}
