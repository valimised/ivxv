package modqp

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/cryptobyte"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/math/group"
	"tivi.io/core/math/group/internal"
)

type modqPGroup struct {
	name      string
	generator *big.Int
	// groupOrder is Scalar modulo.
	groupOrder *big.Int
	// fieldOrder is Element modulo.
	fieldOrder *big.Int
}

func (G *modqPGroup) Name() string {
	return G.name
}

// Order returns Sophie-Germain prime value q.
func (G *modqPGroup) Order() *big.Int {
	return new(big.Int).Set(G.groupOrder)
}

// FieldOrder returns safe prime value p.
func (G *modqPGroup) FieldOrder() *big.Int {
	return new(big.Int).Set(G.fieldOrder)
}

// Identity returns 1.
func (G *modqPGroup) Identity() group.Element {
	return G.identity()
}

// identity returns 1.
//
// identity is a helper method for Identity, in order to
// return concrete implementation modqPElement and not group.Element interface.
func (G *modqPGroup) identity() *modqPElement {
	return &modqPElement{
		g:     G,
		value: big.NewInt(1),
	}
}

// Generator returns group generator value g, most likely 2.
func (G *modqPGroup) Generator() group.Element {
	return &modqPElement{
		g:     G,
		value: new(big.Int).Set(G.generator),
	}
}

func (G *modqPGroup) Equal(G2 group.Group) error {
	if G.Order().Cmp(G2.Order()) != 0 {
		return internal.DifferentOrder{Prefix: logPrefix}
	}

	if G.FieldOrder().Cmp(G2.FieldOrder()) != 0 {
		return internal.DifferentFieldOrder{Prefix: logPrefix}
	}

	G3, err := internal.CastGroupTo[*modqPGroup](G2)
	if err != nil {
		return err
	}

	if G.generator.Cmp(G3.generator) != 0 {
		return fmt.Errorf("%s: groups have different generators", logPrefix)
	}

	return nil
}

// PadBytes pads a data into a byte slice that looks like:
//
//	{0x00 0x01 0xFF 0xFF ... 0xFF 0x00 X X ... X}
//
// where 0x00, 0x01 are required and 0xFF are optional bytes.
// X X ... X denotes msg bytes, which are filled up until padded msg end.
func (G *modqPGroup) PadBytes(msg []byte) ([]byte, error) {
	// Max msg bytelen that can be padded
	maxMsgByteLen := G.maxPaddedMsgByteLen()

	// 3 bytes are reserved for padding head 0x00 0x01 and end 0x00 bytes
	if len(msg) > maxMsgByteLen-2-1 {
		return nil, internal.MessageByteLenTooLarge{Prefix: logPrefix, MsgByteLen: len(msg), MaxMsgByteLen: maxMsgByteLen - 2 - 1}
	}

	padded := make([]byte, maxMsgByteLen)

	// Padding head 0x00 and 0x01 bytes
	padded[0] = 0x00
	padded[1] = 0x01
	// Calculate how many 0xFF bytes will be inserted between padding head and end bytes
	paddingEnd := maxMsgByteLen - 1 - len(msg)
	// Padding end 0x00 byte
	padded[paddingEnd] = 0x00

	// Fill slice with 0xFF padding bytes, right after padding head and up until padding end bytes
	for i := range padded[2:paddingEnd] {
		padded[2+i] = 0xFF
	}

	// Copy msg bytes right after padding end 0x00 byte
	copy(padded[maxMsgByteLen-len(msg):], msg)

	return padded, nil
}

// PadBits uses bit padding schema, instead of byte padding as done in Pad.
// Bit padding is more efficient, see group.PadBits for details.
func (G *modqPGroup) PadBits(msg []byte) (padded []byte, err error) {
	return group.PadBits(msg, uint64(G.groupOrder.BitLen()))
}

// UnpadBytes strips padding bytes from padded and returns initial message.
func (G *modqPGroup) UnpadBytes(padded []byte) ([]byte, error) {
	if len(padded) != G.maxPaddedMsgByteLen() {
		return nil, internal.PaddedByteLen{Prefix: logPrefix, MsgByteLen: len(padded), ExpectedByteLen: G.maxPaddedMsgByteLen()}
	}

	if padded[0] != 0x00 {
		return nil, fmt.Errorf("%s: first byte %x is not equal to expected 0x00", logPrefix, padded[0])
	}

	if padded[1] != 0x01 {
		return nil, fmt.Errorf("%s: second byte %x is not equal to expected 0x01", logPrefix, padded[1])
	}

	// Loop exactly until 0x00 padding end byte found,
	// or until there are no bytes left to read from padded
	for i, p := range padded[2:] {
		switch p {
		case 0xFF: // skip
		case 0x00: // end of padding
			return padded[i+2+1:], nil
		default:
			return nil, internal.UnexpectedPaddingByte{Prefix: logPrefix, Byte: p, Index: i}
		}
	}

	return nil, internal.NoEncodedMessage{Prefix: logPrefix}
}

// UnpadBits unpads a message that was padded in PadBits, see group.UnpadBits
// for details.
func (G *modqPGroup) UnpadBits(padded []byte) (msg []byte, err error) {
	return group.UnpadBits(padded)
}

func (G *modqPGroup) IsGroupElement(E group.Element) error {
	E1, err := internal.CastElementTo[*modqPElement](E)
	if err != nil {
		return err
	}

	if err = G.Equal(E1.g); err != nil {
		return err
	}

	if E1.value.Cmp(big.NewInt(0)) < 1 {
		return internal.ElementLessThanOne{Prefix: logPrefix}
	}

	if E1.value.Cmp(G.fieldOrder) > 0 {
		return fmt.Errorf("%s: element > safe prime p value", logPrefix)
	}

	if big.Jacobi(E1.value, G.fieldOrder) != 1 {
		return fmt.Errorf("%s: element  is not a quadratic residue", logPrefix)
	}

	return nil
}

// Encode encodes msg into group element. Msg can be any value from a [1, Order()] range.
// After encoding, all group elements from a (Order(), FieldOrder()) range are quadratic
// non-residues, and group elements from a [1, Order()] range are quadratic residues.
func (G *modqPGroup) Encode(msg []byte) (group.Element, error) {
	E := new(big.Int).SetBytes(msg)

	if E.Cmp(big.NewInt(0)) < 1 {
		return nil, internal.ElementLessThanOne{Prefix: logPrefix}
	}

	if E.Cmp(G.groupOrder) > 0 {
		return nil, fmt.Errorf("%s: padded message value cannot be greater than generator order", logPrefix)
	}

	// Check if E is a quadratic residue
	switch big.Jacobi(E, G.fieldOrder) {
	case 0:
		// Should never happen, since we check on 0
		return nil, fmt.Errorf("%s: padded message value is 0", logPrefix)
	case -1:
		// E is not a quadratic residue.
		// Subtracting also means that now E becomes > group generator order,
		// which therefore means that all group element which are > group generator
		// order are quadratic non-residues, and those which value is <= group
		// generator order are quadratic residues
		E.Sub(G.fieldOrder, E)
	}

	// Guaranteed to be quadratic residue
	E2 := G.identity()
	E2.value.Set(E)

	return E2, nil
}

// ElementOf ASN.1 unmarshalls group element and ensures that it belongs to the group.
func (G *modqPGroup) ElementOf(der asn_1.DER) (group.Element, error) {
	integer := cryptobyte.String(der)
	E := new(big.Int)

	if !integer.ReadASN1Integer(E) {
		return nil, internal.NotASN1Integer{Prefix: logPrefix}
	}

	if !integer.Empty() {
		return nil, internal.ASN1TrailingBytes{Prefix: logPrefix}
	}

	E2 := G.identity()
	E2.value.Set(E)
	E2.value.Mod(E2.value, E2.g.fieldOrder)

	// We must ensure that element actually belongs to this group
	if err := G.IsGroupElement(E2); err != nil {
		return nil, err
	}

	return E2, nil
}

// maxPaddedMsgByteLen returns non-inclusive upper limit of a maximum bytes
// length that padded message may have to be encoded into a ModqP group element.
func (G *modqPGroup) maxPaddedMsgByteLen() int {
	return (G.groupOrder.BitLen() + 7) / 8
}
