// Package ciphertext provides implementation for ElGamal non-interactive zero-knowledge proof
// ciphertext commitment. Ciphertext commitment ensures a verifier, that ciphertext and decrypted
// ciphertext have a relationship.
package ciphertext

import (
	"fmt"

	"tivi.io/core/crypto/elgamal/nizkp"
	"tivi.io/core/math/group"
)

const (
	ciphertextCommitmentsCount = 4
)

// Commitment is a ciphertext commitment that ensures a verifier, that ciphertext and decrypted
// ciphertext have a relationship.
type Commitment struct {
	// g is an abstract group that both ciphertext and decrypted ciphertext belong to.
	g group.Group

	// a is an ElGamal ciphertext A parameter.
	a group.Element

	// b is an ElGamal ciphertext B parameter.
	b group.Element

	// dec is an ElGamal decrypted ciphertext.
	dec group.Element
}

// NewCiphertextCommitment returns new ElGamal non-interactive zero-knowledge proof ciphertext commitment.
func NewCiphertextCommitment(g group.Group, a group.Element, b group.Element, dec group.Element) (*Commitment, error) {
	return &Commitment{
		g:   g,
		a:   a,
		b:   b,
		dec: dec,
	}, nil
}

// Count returns 4.
func (cc *Commitment) Count() uint64 {
	return ciphertextCommitmentsCount
}

// Create creates prover's ElGamal non-interactive zero-knowledge proof ciphertext commitment.
func (cc *Commitment) Create(s *group.Scalar) ([]group.Element, error) {
	cipherc, err := cc.a.Scale(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create prover's ElGamal non-interactive zero-knowledge proof ciphertext commitment: %v", err)
	}

	return []group.Element{cipherc, cc.a, cc.b, cc.dec}, nil
}

func (cc *Commitment) Verify(proof *nizkp.Proof) ([]group.Element, error) {
	dec := cc.dec.Inverse()

	// Parse encoded message out of ElGamal decryption
	msg, err := cc.b.Op(dec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encoded message out of ElGamal decryption: %v", err)
	}

	// Recreate prover's ciphertext commitment and verify it
	cipherc, err := nizkp.Verify(cc.a, msg, proof.Challenge(), proof.Response())
	if err != nil {
		return nil, fmt.Errorf("failed to verify prover's ElGamal non-interactive zero-knowledge proof ciphertext commitment: %v", err)
	}

	return []group.Element{cipherc, cc.a, cc.b, cc.dec}, nil
}
