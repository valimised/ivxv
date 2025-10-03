package elgamal

import (
	"fmt"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/math/group"
)

// Decryption is ElGamal decrypted group element.
type Decryption struct {
	g       group.Group
	element group.Element
}

// NewDecryption creates new ElGamal decrypted group element.
func NewDecryption(g group.Group, element group.Element) *Decryption {
	return &Decryption{
		g:       g,
		element: element,
	}
}

// Value returns ElGamal decrypted group element.
func (d *Decryption) Value() group.Element {
	return d.element
}

// Plaintext decodes a plaintext from ElGamal decrypted group element.
func (d *Decryption) Plaintext() (plaintext []byte, err error) {
	unpadded, err := d.element.Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode padded plaintext from ElGamal decrypted group element: %v", err)
	}

	plaintext, err = d.g.UnpadBytes(unpadded)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad plaintext from padded plaintext: %v", err)
	}
	return
}

// ASN1Marshal ASN.1 marshals ElGamal decrypted group element as
//
//	d ::= GROUP ELEMENT
//
//	GROUP ELEMENT ::= CHOICE {
//		ModqPElement	INTEGER
//		ECqPElement	OCTET STRING
//		Edwards25519Element	OCTET STRING
//	}
func (d *Decryption) Marshal() (der asn_1.DER, err error) {
	der, err = d.element.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal decrypted group element: %v", err)
	}
	return
}

// ASN1UnmarshalDecryption ASN.1 unmarshalls ElGamal decrypted group element.
func ASN1UnmarshalDecryption(g group.Group, data []byte) (*Decryption, error) {
	E, err := g.ElementOf(data)
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 unmarshal ElGamal decrypted group element: %v", err)
	}

	return NewDecryption(g, E), nil
}
