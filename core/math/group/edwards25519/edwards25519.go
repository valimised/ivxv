// Package edwards25519 implements prime-order subgroup of Edwards25519 elliptic curve,
// using Ristretto255 g.
package edwards25519

import (
	"fmt"
	"github.com/gtank/ristretto255"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
	"math/big"
	asn_1 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group/internal"

	"tivi.io/core/math/group"
)

const (
	// Maximum encoded message bytes size for Edwards25519 curve
	maxEncodedMsgBytesLen = 32
	// Maximum message bytes size for Edwards25519 curve
	maxMsgBytesLen = 29
)

// init registers Edwards25519 g in a group.registry.
func init() {
	group.Register(new(edwards25519Group))
}

var (
	// l is an order of an Edwards25519 curve base point (b), i.e. generator order
	// Ref: https://datatracker.ietf.org/doc/html/rfc8032#section-5.1
	l = []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x14, 0xde, 0xf9, 0xde, 0xa2, 0xf7, 0x9c, 0xd6, 0x58, 0x12, 0x63, 0x1a, 0x5c, 0xf5, 0xd3, 0xed}
	// p is an order of an Edwards25519 curve field, i.e. g field order
	// Ref: https://datatracker.ietf.org/doc/html/rfc8032#section-5.1
	p = []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xed}
)

type edwards25519Group struct{} //nolint:revive

func (G *edwards25519Group) Name() string {
	return "Edwards25519"
}

// Order returns Ristretto255 group L parameter.
func (G *edwards25519Group) Order() *big.Int {
	return new(big.Int).SetBytes(l)
}

// FieldOrder returns Ristretto255 group P parameter.
func (G *edwards25519Group) FieldOrder() *big.Int {
	return new(big.Int).SetBytes(p)
}

// Identity returns Edwards25519 curve point at (X=0, Y=1, Z=1, T=0).
func (G *edwards25519Group) Identity() group.Element {
	return G.identity()
}

// identity returns Edwards25519 curve point at (X=0, Y=1, Z=1, T=0).
//
// identity is a helper method for Identity, in order to
// return concrete implementation edwards25519Element and not group.Element.
func (G *edwards25519Group) identity() *edwards25519Element {
	return &edwards25519Element{
		g: G,
		p: ristretto255.NewElement(),
	}
}

// Generator returns Edwards25519 curve b parameter, i.e. base point.
func (G *edwards25519Group) Generator() group.Element {
	return &edwards25519Element{
		g: G,
		p: ristretto255.NewElement().Base(),
	}
}

func (G *edwards25519Group) Equal(G2 group.Group) error {
	if G.Order().Cmp(G2.Order()) != 0 {
		return fmt.Errorf("group/edwards25519: %s", internal.ErrDifferentGenOrder)
	}

	if G.FieldOrder().Cmp(G2.FieldOrder()) != 0 {
		return fmt.Errorf("group/edwards25519: %s", internal.ErrDifferentFieldOrder)
	}

	return nil
}

// Pad is unimplemented for Edwards25519 group. Padding is done in EncodeMessage.
func (G *edwards25519Group) PadBytes(msg []byte) ([]byte, error) {
	return msg, nil
}

// PadBits is unimplemented for Edwards25519 group. Padding is done in EncodeMessage.
func (G *edwards25519Group) PadBits(msg []byte) ([]byte, error) {
	return msg, nil
}

// Unpad is unimplemented for Edwards25519 group. Unpadding is done in Decode.
func (G *edwards25519Group) UnpadBytes(padded []byte) (msg []byte, err error) {
	return padded, nil
}

// UnpadBits is unimplemented for Edwards25519 group. Unpadding is done in Decode.
func (G *edwards25519Group) UnpadBits(padded []byte) (msg []byte, err error) {
	return padded, nil
}

// IsGroupElement is unimplemented for Edwards25519 group and always returns nil.
func (G *edwards25519Group) IsGroupElement(_ group.Element) error {
	return nil
}

// EncodeMessage encodes a msg, whose byte length is <= 29, into Edwards25519 curve point.
func (G *edwards25519Group) Encode(msg []byte) (group.Element, error) {
	E := G.identity()

	// 32 - 29 = 3, these 3 bytes are reserved and must always present in encoded message
	if len(msg) > maxMsgBytesLen {
		return nil, fmt.Errorf("group/edwards25519: %s: %d", internal.ErrLargeMessageBytes, maxMsgBytesLen)
	}

	// initialized with 0, i.e. []byte{0, 0, 0, ...}
	var buf [maxEncodedMsgBytesLen]byte

	// []byte{0, 0, len(msg), 0, 0, ...}
	buf[2] = byte(len(msg))

	// if len(msg) == 29, then
	// []byte{0, 0, len(msg), M, E, S, S, a, g, E}
	//
	// if len(msg) < 29, then
	// []byte{0, 0, len(msg), M, E, S, S, a, g, E, 0, 0, ..., 0}
	copy(buf[3:], msg)

	// Loop until buf[1] != 255 or msg is successfully encoded (err == nil)
	var err error

	// TODO: possibly infinite loop
	for ; buf[1] != 0xFF; buf[1]++ {
		err = E.p.Decode(buf[:])
		if err == nil {
			return E, nil
		}
	}

	return nil, fmt.Errorf("group/edwards25519: cannot encode a message into Edwards25519 curve point: %v", err)
}

// ElementOfASN1 ASN.1 unmarshalls group element and ensures that it belongs to the group.
func (G *edwards25519Group) ElementOf(der asn_1.DER) (group.Element, error) {
	octetString := cryptobyte.String(der)
	var ristrettoElement cryptobyte.String

	if !octetString.ReadASN1(&ristrettoElement, asn1.OCTET_STRING) {
		return nil, fmt.Errorf("group/edwards25519: %s", internal.ErrNotASN1OctetString)
	}

	if !octetString.Empty() {
		return nil, fmt.Errorf("group/edwards25519: %s", internal.ErrTrailingBytesASN1)
	}

	E := G.identity()

	if err := E.p.Decode(ristrettoElement); err != nil {
		return nil, fmt.Errorf("group/edwards25519: cannot decode Ristretto255 element into Edwards25519 curve point: %v", err)
	}

	// Always nil
	if err := G.IsGroupElement(E); err != nil {
		return nil, err
	}

	return E, nil
}

type edwards25519Element struct { //nolint:revive
	g *edwards25519Group
	// p is an underlying Ristretto255 g element.
	p *ristretto255.Element
}

// Inverse returns E2 that satisfies
//
//	E * E2 = 1
func (E *edwards25519Element) Inverse() group.Element { //nolint:revive
	E2 := E.g.identity()

	E2.p.Negate(E.p)

	return E2
}

// Scale returns
//
//	E * s.
func (E *edwards25519Element) Scale(s *group.Scalar) (group.Element, error) { //nolint:revive
	E2 := E.g.identity()

	if s.Modulo().Cmp(E.g.Order()) != 0 {
		return nil, fmt.Errorf("group/edwards25519: %s", internal.ErrScalarModCmpGenOrder)
	}

	s2, err := toRistrettoScalar(s)
	if err != nil {
		return nil, err
	}

	E2.p.ScalarMult(s2, E.p)

	return E2, nil
}

// Op returns
//
//	E1 + E2.
func (E *edwards25519Element) Op(E2 group.Element) (group.Element, error) {
	E3, err := internal.CastElementTo[*edwards25519Element](E2)
	if err != nil {
		return nil, err
	}

	err = E.g.Equal(E3.g)
	if err != nil {
		return nil, err
	}

	E4 := E.g.identity()

	E4.p.Add(E.p, E3.p)

	return E4, nil
}

// Equal return nil if E == E2.
func (E *edwards25519Element) Equal(E2 group.Element) error {
	E3, err := internal.CastElementTo[*edwards25519Element](E2)
	if err != nil {
		return err
	}

	if err = E.g.Equal(E3.g); err != nil {
		return err
	}

	if E.p.Equal(E3.p) != 1 {
		return fmt.Errorf("group/edwards25519: different Ristretto255 elements")
	}

	return nil
}

// DecodeMessage decodes a message from Edwards25519 curve point.
func (E *edwards25519Element) Decode() ([]byte, error) { //nolint:revive
	// Ensure that this group element actually belongs to the correct group.
	if err := E.g.IsGroupElement(E); err != nil {
		return nil, err
	}

	buf := E.p.Encode(nil)
	if buf[0] != 0x00 {
		return nil, fmt.Errorf("group/edwards25519: first byte of decoded message should be 0x00")
	}

	// Ensure that decoded message bytelen isn't > 29
	if buf[2] > 0x1D {
		return nil, fmt.Errorf("group/edwards25519: %s: %d", internal.ErrLargeMessageBytes, maxMsgBytesLen)
	}

	// +3 because we set slice end index, buf[2] returns msg bytelen, but we
	// start slicing from index=3, therefore our end index becomes buf[2]+3
	return buf[3 : buf[2]+3], nil
}

// Marshal ASN.1 marshals group element as
//
//	E ::= OCTET STRING
//
// where OCTET STRING is Edwards25519 curve point.
func (E *edwards25519Element) Marshal() (der asn_1.DER, err error) { //nolint:revive
	var builder cryptobyte.Builder
	builder.AddASN1OctetString(E.p.Encode(nil))

	der, err = builder.Bytes()
	if err != nil {
		return nil, err
	}
	return
}

// toRistrettoScalar converts from big-endian Edwards25519 scalar
// to little-endian Ristretto255 scalar.
func toRistrettoScalar(s *group.Scalar) (rs *ristretto255.Scalar, err error) {
	var bigEndian [32]byte
	s.Value().FillBytes(bigEndian[:])

	var littleEndian [32]byte
	for i := 0; i < 32; i++ {
		littleEndian[31-i] = bigEndian[i]
	}

	rs = ristretto255.NewScalar()

	err = rs.Decode(littleEndian[:])
	if err != nil {
		return nil, err
	}

	return
}
