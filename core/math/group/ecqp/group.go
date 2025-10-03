package ecqp

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
	"math/bits"

	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"

	asn_1 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group"
	"tivi.io/core/math/group/internal"
)

type ecqPGroup struct {
	name string
	// fieldEncodingBits is amount of bits reserved for encoding message into an
	// elliptic curve point. Should be 1 < fieldEncodingBits < 21.
	fieldEncodingBits uint8
	// Only use unnamed fields, like this one, if methods doesn't collide.
	elliptic.Curve
}

func (G *ecqPGroup) Name() string {
	return G.name
}

// Order returns elliptic curve N parameter.
func (G *ecqPGroup) Order() *big.Int {
	return new(big.Int).Set(G.Params().N)
}

// FieldOrder returns elliptic curve P parameter.
func (G *ecqPGroup) FieldOrder() *big.Int {
	return new(big.Int).Set(G.Params().P)
}

// Identity returns elliptic curve point at (0, 0).
func (G *ecqPGroup) Identity() group.Element {
	return G.identity()
}

// identity is a helper method for Identity, in order to
// return implementation ecqPElement and not group.Element interface.
func (G *ecqPGroup) identity() *ecqPElement {
	return &ecqPElement{
		g: G,
		x: new(big.Int),
		y: new(big.Int),
	}
}

// Generator returns elliptic curve base point (Gx, Gy).
func (G *ecqPGroup) Generator() group.Element {
	return &ecqPElement{
		g: G,
		x: new(big.Int).Set(G.Params().Gx),
		y: new(big.Int).Set(G.Params().Gy),
	}
}

// Equal returns nil if generator and field orders are the same.
func (G *ecqPGroup) Equal(G2 group.Group) error {
	if G.Order().Cmp(G2.Order()) != 0 {
		return internal.DifferentOrder{Prefix: logPrefix}
	}

	if G.FieldOrder().Cmp(G2.FieldOrder()) != 0 {
		return internal.DifferentFieldOrder{Prefix: logPrefix}
	}

	return nil
}

// PadBytes pads a data into a byte-aligned slice that looks like:
//
//	0b0111...0XX...X
//
// where 0b01 is a padding head, 11... are padding bits, ...0 is a padding end and XX...X are msg bits.
func (G *ecqPGroup) PadBytes(msg []byte) ([]byte, error) {
	fieldBitLen := G.Params().BitSize

	// Max padded msg size, reserve padding head (2), padding end (1) and field encoding bits
	maxMsgBitLen := fieldBitLen - 2 - 1 - int(G.fieldEncodingBits)

	// Msg is byte-aligned, i.e. 0b110110 will be represented as 0b00110110 (0 bits included)
	msgBitLen := len(msg) * 8 //nolint:mnd
	if msgBitLen > maxMsgBitLen {
		return nil, internal.MessageBitLenTooLarge{Prefix: logPrefix, MsgBitLen: msgBitLen, MaxMsgBitLen: maxMsgBitLen}
	}

	// Padding head and end bits are included in paddingBitLen.
	//
	// -1 here is due to left shift nature. Suppose you wish
	// to shift 8 bits to the left, so you would expect to get 0b10000000,
	// however when you do 1<<8 you get 0b00000001 0b00000000 and that
	// is because we start with initial 1 value, which is 0b00000001.
	// So -1 ensures that we actually shift only desired amount of bits to
	// the left.
	paddingBitLen := fieldBitLen - 1 - msgBitLen - int(G.fieldEncodingBits)

	padding := big.NewInt(1)
	// 0b100...000
	padding.Lsh(padding, uint(paddingBitLen)) //nolint:gosec
	// Sets most-significant bit (msb) and least-significant bit (lsb) to 0,
	// this operation creates padding head and padding end bits 0b011...110
	padding.Sub(padding, big.NewInt(2)) //nolint:mnd
	// Appends msg bytes to the end of padding (head+padding+end)
	return append(padding.Bytes(), msg...), nil // 0b011...110 || msg
}

// PadBits uses the same padding schema as PadBytes, but in a more efficient way,
// see group.PadBits for details.
func (G *ecqPGroup) PadBits(msg []byte) ([]byte, error) {
	// We should leave space for elliptic curve field encoding bits
	return group.PadBits(msg, uint64(G.FieldOrder().BitLen())-uint64(G.fieldEncodingBits)) //nolint:wrapcheck,gosec
}

// UnpadBytes strips padding bytes from padded and returns initial message.
func (G *ecqPGroup) UnpadBytes(padded []byte) ([]byte, error) {
	// When 0xFE (padding end) is msb then convert 0b11111110 to 0b01111111
	msb := padded[0] >> 1

	// Check that first byte can only contain sequence of 0 bits
	// followed by sequence of 1 bits (0b00001111), and combinations such as
	// 0b00101111 are not possible
	if bits.LeadingZeros8(msb) != 8-bits.OnesCount8(msb) {
		return nil, internal.UnexpectedPaddingHeader{Prefix: logPrefix, Byte: msb}
	}

	// Will change first byte to 0xFF, unless first byte is 0xFE
	padded[0] |= 0xFE

	// Loop exactly until 0xFE padding end byte found,
	// or until there are no bytes left to read from padded
	for i, b := range padded {
		switch b {
		// skip
		case 0xFF: //nolint:mnd
		// padding end
		case 0xFE: //nolint:mnd
			// +1, because current i position is at 0x00 padding end, so msg starts from next byte
			return padded[i+1:], nil
		default:
			return nil, internal.UnexpectedPaddingByte{Prefix: logPrefix, Byte: b, Index: i}
		}
	}

	return nil, internal.NoEncodedMessage{Prefix: logPrefix}
}

// UnpadBits unpads a message that was padded in PadBits, see group.UnpadBits
// for details.
func (G *ecqPGroup) UnpadBits(padded []byte) ([]byte, error) {
	return group.UnpadBits(padded) //nolint:wrapcheck
}

// IsGroupElement checks that E is on an elliptic curve.
func (G *ecqPGroup) IsGroupElement(E group.Element) error {
	E1, err := internal.CastElementTo[*ecqPElement](E)
	if err != nil {
		return err
	}

	// Groups equal
	if err := G.Equal(E1.g); err != nil {
		return err
	}

	if err := isOnCurve(G, E1.x, E1.y); err != nil {
		return err
	}

	return nil
}

// Encode encodes a msg into an elliptic curve point using short Weierstrass equation.
func (G *ecqPGroup) Encode(msg []byte) (group.Element, error) { //nolint:funlen
	x := new(big.Int).SetBytes(msg)

	maxMsgBitLen := G.FieldOrder().BitLen() - int(G.fieldEncodingBits)

	if x.BitLen() > maxMsgBitLen {
		return nil, internal.MessageBitLenTooLarge{Prefix: logPrefix, MsgBitLen: x.BitLen(), MaxMsgBitLen: maxMsgBitLen}
	}

	// We allocate bits for field encoding
	x = x.Lsh(x, uint(G.fieldEncodingBits)) // 0b...000

	// Limit for field encoding attempts, < 2^fieldEncodingBits
	limit := 2 << (G.fieldEncodingBits - 1) //nolint:mnd

	// result of short Weierstrass equation is y^2
	var result *big.Int

	// Loop until x is a quadratic residue or iterations limit is reached
	for range limit {
		result = shortWeierstrassFunc(x, G.Params())

		// Is result a perfect square?
		if big.Jacobi(result, G.Params().P) == 1 {
			break
		}

		//nolint: godox
		// TODO: uncomment me to use random field encoding bits
		//// This mask will set to 1 all field encoding bits.
		// bitMask := new(big.Int).SetInt64(1 << fieldEncodingBits)
		// bitMask.Sub(bitMask, big.NewInt(1))
		//
		// random, err := rand.Int(rand.Reader, bitMask)
		// if err != nil {
		//	return nil, err
		// }
		//
		//// Will insert random value into allocated field encoding bits space
		// x.Or(x, random)

		x.Add(x, big.NewInt(1)) // 0b...0[+1]
	}

	// First solution of short Weierstrass equation is sqrt(result),
	// becomes (X,Y)
	y1 := new(big.Int).ModSqrt(result, G.Params().P)

	// This case should not happen, but rather be safe
	if y1 == nil {
		return nil, fmt.Errorf("%s: short Weierstrass equation solved to 0", logPrefix)
	}

	// Second solution of Weierstrass equation sqrt(y2),
	// becomes (X,-Y)
	y2 := new(big.Int).Neg(y1)
	y2.Mod(y2, G.FieldOrder())

	// We have to pick an Y value for an elliptic curve point (X,Y)
	y := new(big.Int)

	// Make decision which one to pick - take smallest Y
	// https://github.com/verificatum/verificatum-vcr/blob/97974cfc4ebbb323e49396222823e226cae2bebe/src/java/com/verificatum/arithm/ECqPGroup.magic#L485
	if y2.Cmp(y1) < 0 {
		y.Set(y2)
	} else {
		y.Set(y1)
	}

	E2 := G.identity()
	E2.x.Set(x)
	E2.y.Set(y)

	return E2, nil
}

// ElementOf ASN.1 unmarshalls group element and ensures that it belongs to the group.
func (G *ecqPGroup) ElementOf(der asn_1.DER) (group.Element, error) {
	octetString := cryptobyte.String(der)

	var point cryptobyte.String

	if !octetString.ReadASN1(&point, asn1.OCTET_STRING) {
		return nil, internal.NotASN1OctetString{Prefix: logPrefix}
	}

	if !octetString.Empty() {
		return nil, internal.ASN1TrailingBytes{Prefix: logPrefix}
	}

	// Unmarshal elliptic curve point
	x, y, err := unmarshal(point, G)
	if err != nil {
		return nil, err
	}

	E := G.identity()
	E.x.Set(x)
	E.y.Set(y)

	// We must ensure that element actually belongs to this group
	if err := G.IsGroupElement(E); err != nil {
		return nil, err
	}

	return E, nil
}
