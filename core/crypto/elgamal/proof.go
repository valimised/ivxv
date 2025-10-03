package elgamal

import (
	"fmt"

	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"

	"tivi.io/core/crypto/elgamal/nizkp"
	"tivi.io/core/math/group"
)

// DecryptionProof holds an ElGamal non-interactive zero-knowledge proof and its related metadata.
type DecryptionProof struct {
	ciphertext *Ciphertext
	decryption *Decryption
	proof      *nizkp.Proof
}

func NewDecryptionProof(ciphertext *Ciphertext, decryption *Decryption, proof *nizkp.Proof) *DecryptionProof {
	return &DecryptionProof{
		ciphertext: ciphertext,
		decryption: decryption,
		proof:      proof,
	}
}

func (p *DecryptionProof) Ciphertext() *Ciphertext {
	return p.ciphertext
}

func (p *DecryptionProof) Decryption() *Decryption {
	return p.decryption
}

func (p *DecryptionProof) Proof() *nizkp.Proof {
	return p.proof
}

// MarshalASN1 ASN.1 marshals ElGamal non-interactive zero-knowledge proof as
//
//	p ::= SEQUENCE {
//		Ciphertext	CIPHERTEXT
//		Decryption	DECRYPTION
//		Proof	PROOF
//	}
//
//	CIPHERTEXT ::= {
//		a	GROUP ELEMENT
//		b	GROUP ELEMENT
//	}
//
//	DECRYPTION ::= GROUP ELEMENT
//
//	PROOF ::= {
//		challenge	INTEGER
//		response	INTEGER
//	}
//
//	GROUP ELEMENT ::= CHOICE {
//		ModPElement	INTEGER
//		NistElement	OCTET STRING
//		Edwards25519Element	OCTET STRING
//	}
func (p *DecryptionProof) MarshalASN1() (der []byte, err error) {
	ctDer, err := p.ciphertext.MarshalASN1()
	if err != nil {
		return nil, fmt.Errorf("cannot ASN.1 marshal ElGamal proof ciphertext: %v", err)
	}

	decDer, err := p.decryption.Value().Marshal()
	if err != nil {
		return nil, fmt.Errorf("cannot ASN.1 marshal ElGamal proof decryption: %v", err)
	}

	transDer, err := p.proof.MarshalASN1()
	if err != nil {
		return nil, fmt.Errorf("cannot ASN.1 marshal ElGamal proof transcript: %v", err)
	}

	var builder cryptobyte.Builder
	builder.AddASN1(asn1.SEQUENCE, func(builder *cryptobyte.Builder) {
		builder.AddBytes(ctDer)
		builder.AddBytes(decDer)
		builder.AddBytes(transDer)
	})

	der, err = builder.Bytes()
	if err != nil {
		return nil, fmt.Errorf("cannot ASN.1 marshal ElGamal proof: %v", err)
	}

	return
}

func ASN1UnmarshalDecryptionProof(g group.Group, data []byte) (*DecryptionProof, error) {
	der := cryptobyte.String(data)
	var sequence, ctDer, decDer, transDer cryptobyte.String

	if !der.ReadASN1(&sequence, asn1.SEQUENCE) {
		return nil, fmt.Errorf("cannot ASN.1 unmarshal ElGamal proof SEQUENCE")
	}

	if !sequence.ReadAnyASN1Element(&ctDer, nil) {
		return nil, fmt.Errorf("cannot ASN.1 unmarshal ElGamal proof ciphertext")
	}

	if !sequence.ReadAnyASN1Element(&decDer, nil) {
		return nil, fmt.Errorf("cannot ASN.1 unmarshal ElGamal proof decryption")
	}

	if !sequence.ReadAnyASN1Element(&transDer, nil) {
		return nil, fmt.Errorf("cannot ASN.1 unmarshal ElGamal proof transcript")
	}

	if !der.Empty() || !sequence.Empty() {
		return nil, fmt.Errorf("trailing bytes while ASN.1 unmarshalling ElGamal proof")
	}

	ciphertext, err := ASN1UnmarshalCiphertext(g, ctDer)
	if err != nil {
		return nil, fmt.Errorf("cannot create ElGamal ciphertext from ASN.1 unmarshalled ElGamal proof ciphertext: %v", err)
	}

	decrypted, err := ASN1UnmarshalDecryption(g, decDer)
	if err != nil {
		return nil, fmt.Errorf("cannot create ElGamal decryption from ASN.1 unmarshalled ElGamal proof decryption: %v", err)
	}

	transcript, err := nizkp.ASN1UnmarshalProof(g, transDer)
	if err != nil {
		return nil, fmt.Errorf("cannot create ElGamal transcript from ASN.1 unmarshalled ElGamal proof transcript: %v", err)
	}

	return NewDecryptionProof(ciphertext, decrypted, transcript), nil
}
