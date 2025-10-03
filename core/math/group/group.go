// Package group defines an interface for finite cyclic abstract groups.
// An abstract group is a set that has the following properties:
//
//	a) has operation either * or +
//	b) applying operation on any of the group elements should always return a group element
//	c) each group element has an inverse element in the group
//	d) group has an identity element
//	e) all operations are associative
package group

import (
	"fmt"
	"math/big"

	asn_1 "tivi.io/core/encoding/asn1"

	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
)

// Group is a finite cyclic abstract group. You use Group interface to
// create new Element.
type Group interface {
	// Name returns name of this group.
	Name() string

	// Order returns order of this group generator and is used as a Scalar modulo.
	Order() *big.Int

	// FieldOrder returns order of this group field and is used as Element modulo.
	FieldOrder() *big.Int

	// Identity returns this group identity element.
	Identity() Element

	// Generator returns this group generator element.
	Generator() Element

	// Equal returns nil if this group equals to g.
	Equal(G Group) error

	// PadBytes pads a msg up to the maximum allowed byte length.
	// Maximum allowed byte length decision is up to the implementation.
	PadBytes(msg []byte) (padded []byte, err error)

	// PadBits pads a msg up to the maximum allowed bit length.
	// Maximum allowed bit length decision is up to the implementation.
	PadBits(msg []byte) (padded []byte, err error)

	// UnpadBytes unpads a msg from padded byte slice.
	UnpadBytes(padded []byte) (msg []byte, err error)

	// UnpadBits unpads a msg from padded bit slice.
	UnpadBits(padded []byte) (msg []byte, err error)

	// IsGroupElement return nil if E belongs to this group.
	IsGroupElement(E Element) error

	// Encode encodes a message into this group element.
	Encode(msg []byte) (Element, error)

	// ElementOf ASN.1 unmarshalls group element into Element.
	ElementOf(der asn_1.DER) (Element, error)
}

// Element represents a Group element.
type Element interface {
	// Inverse returns E, which is an inverse of this group element such that condition is met
	//	this * E = identity
	// for multiplicative group
	//	this + E = identity
	// for additive group.
	Inverse() (E Element)

	// Scale returns group element whose value is
	//	this ^ s
	// for multiplicative group
	//	this * s
	// for additive group.
	Scale(s *Scalar) (Element, error)
	// Op returns a result of group operation whose value is
	//	this * E
	// for multiplicative group
	//	this + E
	// for additive group.
	Op(E Element) (Element, error)

	// Equal returns nil if this group element equals to E.
	Equal(E Element) error

	// Decode decodes a message from this group element.
	Decode() ([]byte, error)

	// Marshal returns this group element ASN.1 marshalled.
	Marshal() (asn_1.DER, error)
}

// RandomElement returns random element from a group.
func RandomElement(G Group) (Element, error) {
	s, err := RandomScalar(G.Order())
	if err != nil {
		return nil, err
	}

	return G.Generator().Scale(s)
}

// Marshal ASN.1 marshals group as
//
//	g ::= PrintableString
//
// where PrintableString is a group name.
func Marshal(G Group) (asn_1.DER, error) {
	var builder cryptobyte.Builder
	builder.AddASN1(asn1.PrintableString, func(builder *cryptobyte.Builder) {
		builder.AddBytes([]byte(G.Name()))
	})

	return builder.Bytes()
}

// Unmarshal ASN.1 unmarshalls group.
func Unmarshal(der asn_1.DER) (Group, error) {
	printableString := cryptobyte.String(der)
	var name []byte
	if !printableString.ReadASN1Bytes(&name, asn1.PrintableString) {
		return nil, fmt.Errorf("math/group: group is not an ASN.1 PrintableString")
	}

	if !printableString.Empty() {
		return nil, fmt.Errorf("math/group: incomplete ASN.1 group")
	}

	return Get(string(name))
}
