// Package key provides implementation for ElGamal non-interactive zero-knowledge proof
// key pair commitment. Key pair commitment ensures a verifier, that encryption key is actually
// a valid pair to decryption key.
package key

import (
	"fmt"

	"tivi.io/core/crypto/elgamal/nizkp"
	"tivi.io/core/math/group"
)

const (
	keyPairCommitmentsCount = 1
)

// Commitment is a key pair commitment that ensures a verifier, that encryption key is actually
// a valid pair to decryption key.
type Commitment struct {
	// g is an abstract group that both keys belong to.
	g group.Group

	// pub is a public element of a key.
	pub group.Element
}

// NewKeyPairCommitment returns new ElGamal non-interactive zero-knowledge proof key pair commitment.
func NewKeyPairCommitment(g group.Group, pub group.Element) *Commitment {
	return &Commitment{
		g:   g,
		pub: pub,
	}
}

// Count returns 1.
func (kpc *Commitment) Count() uint64 {
	return keyPairCommitmentsCount
}

// Create creates prover's ElGamal non-interactive zero-knowledge proof key pair commitment.
func (kpc *Commitment) Create(s *group.Scalar) ([]group.Element, error) {
	keyc, err := kpc.g.Generator().Scale(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create prover's ElGamal non-interactive zero-knowledge proof key pair commitment: %v", err)
	}

	return []group.Element{keyc}, nil
}

// Verify verifies prover's ElGamal non-interactive zero-knowledge proof key pair commitment.
func (kpc *Commitment) Verify(t *nizkp.Proof) ([]group.Element, error) {
	// Recreate prover's key commitment and verify it
	keyc, err := nizkp.Verify(kpc.g.Generator(), kpc.pub, t.Challenge(), t.Response())
	if err != nil {
		return nil, fmt.Errorf("failed to verify prover's ElGamal non-interactive zero-knowledge proof key pair commitment: %v", err)
	}

	return []group.Element{keyc}, nil
}
