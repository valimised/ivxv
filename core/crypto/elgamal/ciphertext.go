package elgamal

import (
	"fmt"
	asn_1 "tivi.io/core/encoding/asn1"

	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"

	"tivi.io/core/math/group"
)

// Ciphertext is ElGamal ciphertext.
type Ciphertext struct {
	a group.Element
	b group.Element
}

// NewCiphertext creates new ElGamal ciphertext with provided a and b group elements.
func NewCiphertext(A, B group.Element) *Ciphertext {
	return &Ciphertext{
		a: A,
		b: B,
	}
}

func (ct *Ciphertext) A() group.Element {
	return ct.a
}

func (ct *Ciphertext) B() group.Element {
	return ct.b
}

// ASN1Marshal ASN.1 marshals ElGamal ciphertext as
//
//	ct ::= SEQUENCE {
//		a	GROUP ELEMENT
//		b	GROUP ELEMENT
//	}
//
//	GROUP ELEMENT ::= CHOICE {
//		ModqPElement	INTEGER
//		ECqPElement	OCTET STRING
//		Edwards25519Element	OCTET STRING
//	}
func (ct *Ciphertext) MarshalASN1() (der []byte, err error) {
	A, err := ct.a.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal ciphertext a element: %v", err)
	}

	B, err := ct.b.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal ciphertext b element: %v", err)
	}

	var builder cryptobyte.Builder
	builder.AddASN1(asn1.SEQUENCE, func(builder *cryptobyte.Builder) {
		builder.AddBytes(A)
		builder.AddBytes(B)
	})

	der, err = builder.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal ciphertext: %v", err)
	}
	return
}

// ASN1UnmarshalCiphertext ASN.1 unmarshalls ElGamal ciphertext.
func ASN1UnmarshalCiphertext(g group.Group, data []byte) (*Ciphertext, error) {
	der := cryptobyte.String(data)
	var sequence, ADer, BDer cryptobyte.String

	if !der.ReadASN1(&sequence, asn1.SEQUENCE) {
		return nil, fmt.Errorf("failed to ASN.1 unmarshal ElGamal ciphertext ASN.1 SEQUENCE")
	}

	if !sequence.ReadAnyASN1Element(&ADer, nil) {
		return nil, fmt.Errorf("failed to ASN.1 unmarshal ElGamal ciphertext a element")
	}

	if !sequence.ReadAnyASN1Element(&BDer, nil) {
		return nil, fmt.Errorf("failed to ASN.1 unmarshal ElGamal ciphertext b element")
	}

	if !der.Empty() || !sequence.Empty() {
		return nil, fmt.Errorf("trailing bytes left while ASN.1 unmarshalling ElGamal ciphertext")
	}

	A, err := g.ElementOf(asn_1.DER(ADer))
	if err != nil {
		return nil, fmt.Errorf("failed to create group element from ASN.1 marshalled ElGamal ciphertext a element: %v", err)
	}

	B, err := g.ElementOf(asn_1.DER(BDer))
	if err != nil {
		return nil, fmt.Errorf("failed to create group element from ASN.1 marshalled ElGamal ciphertext b element: %v", err)
	}

	return NewCiphertext(A, B), nil
}
