package internal

import (
	"ivxv.ee/common/collector/crypto/elgamal"

	"tivi.io/core/math/group"
)

type ASN1CiphertextVerifier interface {
	// ASN1CiphertextVerify verifies ASN.1 ciphertext correctness, without
	// decrypting it.
	// We wish to secure offline applications from malformed ciphertexts.
	ASN1CiphertextVerify(ciphertext []byte) error
}

type ElGamalASN1CiphertextVerifier struct {
	g group.Group
}

func NewElGamalASN1CiphertextVerifier(g group.Group) *ElGamalASN1CiphertextVerifier {
	return &ElGamalASN1CiphertextVerifier{g: g}
}

func (e *ElGamalASN1CiphertextVerifier) ASN1CiphertextVerify(data []byte) error {
	_, err := elgamal.UnmarshalCiphertext(e.g, data)
	return err
}
