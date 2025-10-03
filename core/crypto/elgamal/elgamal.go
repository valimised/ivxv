package elgamal

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
	"golang.org/x/crypto/sha3"
	"tivi.io/core/crypto/elgamal/nizkp"
	"tivi.io/core/crypto/elgamal/nizkp/ciphertext"
	"tivi.io/core/crypto/elgamal/nizkp/key"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/math/group"
)

// PublicKey is ElGamal public key.
type PublicKey struct {
	// parameters are ElGamal key parameters.
	parameters *Parameters
	// public is a public component of ElGamal key.
	public group.Element
}

// NewPublicKey returns new ElGamal public key.
func NewPublicKey(parameters *Parameters, public group.Element) *PublicKey {
	return &PublicKey{
		parameters: parameters,
		public:     public,
	}
}

func (pkey *PublicKey) Public() group.Element {
	return pkey.public
}

// Parameters returns ElGamal public key parameters.
func (pkey *PublicKey) Parameters() crypto.AlgorithmIdentifierParameters {
	return pkey.parameters
}

// Fingerprint returns ElGamal public key fingerprint.
func (pkey *PublicKey) Fingerprint() (uint64, error) {
	der, err := pkey.Public().Marshal()
	if err != nil {
		return 0, fmt.Errorf("failed to ASN.1 marshal ElGamal public element: %v", err)
	}

	digest := sha256.Sum256(der)

	// Output only first 24 bytes of SHA256 digest
	return binary.BigEndian.Uint64(digest[24:]), nil
}

// Encrypt creates ASN.1 marshalled ElGamal ciphertext from plaintext.
// Process of creation is following:
//  1. pad a plaintext
//  2. encode a padded plaintext into group element
//  3. encrypt group element and return ElGamal ciphertext
//  4. ASN.1 marshal ElGamal ciphertext
//
// s is considered to be a random scalar.
func (pkey *PublicKey) Encrypt(s *group.Scalar, plaintext []byte) (der []byte, err error) {
	// PadBytes a plaintext
	padded, err := pkey.Parameters().Group().PadBytes(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to pad a plaintext: %v", err)
	}

	// Encode a padded plaintext
	encoded, err := pkey.Parameters().Group().Encode(padded)
	if err != nil {
		return nil, fmt.Errorf("failed to encode a padded plaintext: %v", err)
	}

	// Encrypt encoded group element and return ElGamal ciphertext
	ct, err := pkey.EncryptGroupElement(s, encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt encoded group element: %v", err)
	}

	// ASN.1 marshal ElGamal ciphertext
	der, err = ct.MarshalASN1()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal ciphertext: %v", err)
	}
	return
}

// EncryptGroupElement creates ASN.1 marshalled ElGamal ciphertext from group element.
//
// s is considered to be a random scalar.
func (pkey *PublicKey) EncryptGroupElement(s *group.Scalar, msg group.Element) (*Ciphertext, error) {
	// g {^,*} s
	A, err := pkey.Parameters().Group().Generator().Scale(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create ElGamal ciphertext a element: %v", err)
	}

	// y {^,*} s
	B, err := pkey.Public().Scale(s)
	if err != nil {
		return nil, fmt.Errorf("failed to scale ElGamal public element by random scalar: %v", err)
	}

	// b {*,+} msg
	B, err = B.Op(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ElGamal ciphertext b element: %v", err)
	}

	return NewCiphertext(A, B), nil
}

// Verify verifies ASN.1 marshalled ElGamal proof of correct decryption.
//
// See ProofChallenge for salt description.
func (pkey *PublicKey) Verify(proof, salt []byte) error {
	// decryption proof is a non-interactive zero-knowledge proof
	pr, err := ASN1UnmarshalDecryptionProof(pkey.Parameters().Group(), proof)
	if err != nil {
		return fmt.Errorf("failed to ASN.1 unmarshal ElGamal decryption proof: %v", err)
	}

	// Recreate prover's key commitment
	keyc, err := nizkp.Verify(pkey.Parameters().Group().Generator(), pkey.public, pr.Proof().Challenge(), pr.Proof().Response())
	if err != nil {
		return fmt.Errorf("failed to recreate prover's key commitment: %v", err)
	}

	dec := pr.decryption.Value().Inverse()

	// Create message commitment from ElGamal ciphertext b element
	msgc2, err := pr.ciphertext.B().Op(dec)
	if err != nil {
		return fmt.Errorf("failed to create message commitment from ElGamal ciphertext b element: %v", err)
	}

	// Recreate prover's message commitment
	msgc, err := nizkp.Verify(pr.ciphertext.A(), msgc2, pr.Proof().Challenge(), pr.Proof().Response())
	if err != nil {
		return fmt.Errorf("failed to recreate prover's message commitment: %v", err)
	}

	// Recreate prover's ElGamal proof challenge
	challenge, err := pkey.ProofChallenge(pr.ciphertext, pr.decryption.Value(), msgc, keyc, salt)
	if err != nil {
		return fmt.Errorf("failed to recreate prover's ElGamal proof challenge: %v", err)
	}

	// Recreated prover's challenge should be equal to the challenge that prover has actually calculated
	err = challenge.Equal(pr.Proof().Challenge())
	if err != nil {
		return fmt.Errorf("recreated challenge on verifier side doesn't match the one generated on a prover side: %v", err)
	}

	return nil
}

// ProofChallenge computes the non-interactive challenge (hash).
// Prover generates a challenge and verifier recomputes the same
// challenge on its side and then compares two challenges.
//
// Hash is created using Shake256 and output is always non-deterministically variable length.
//
// Salt is used to add uniqueness to the challenge. So that you can create different challenges
// for the same input to the ProofChallenge, just by providing different salt for each invocation.
func (pkey *PublicKey) ProofChallenge(ct *Ciphertext, dec, msgc, keyc group.Element, salt ...[]byte) (*group.Scalar, error) {
	pubDer, err := pkey.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal public key: %v", err)
	}

	ctDer, err := ct.MarshalASN1()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal ciphertext: %v", err)
	}

	decDer, err := dec.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal decryption: %v", err)
	}

	msgcDer, err := msgc.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal proof message commitment: %v", err)
	}

	keycDer, err := keyc.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal proof key commitment: %v", err)
	}

	// Custom string, acts as a challenge header
	hashID := []byte("DECRYPTION")

	// If salt == nil, then
	// digest = HASH(hashID || pubDer || ctDer ... || keycDer)
	//
	// otherwise,
	// digest = HASH(hashID || pubDer || ctDer ... || keycDer || salt)
	var builder cryptobyte.Builder
	builder.AddASN1(asn1.SEQUENCE, func(builder *cryptobyte.Builder) {
		builder.AddASN1(asn1.PrintableString, func(builder *cryptobyte.Builder) {
			builder.AddBytes(hashID)
		})
		builder.AddBytes(pubDer)
		builder.AddBytes(ctDer)
		builder.AddBytes(decDer)
		builder.AddBytes(msgcDer)
		builder.AddBytes(keycDer)
		if len(salt) > 0 {
			builder.AddASN1(asn1.SEQUENCE, func(builder *cryptobyte.Builder) {
				for _, ex := range salt {
					builder.AddASN1OctetString(ex)
				}
			})
		}
	})

	der, err := builder.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal proof challenge: %v", err)
	}

	// Variable length hash
	hash := sha3.NewShake256()

	// Shake instance does not fail
	hash.Write(der) //nolint: errcheck

	// Create group scalar from hash bytes
	challenge, err := group.ScalarValueOfReader(hash, pkey.Parameters().Group().Order())
	if err != nil {
		return nil, fmt.Errorf("failed to create group scalar from hashed ASN.1 marshalled ElGamal proof challenge: %v", err)
	}

	return challenge, nil
}

// ASN1Marshal ASN.1 marshals ElGamal public key as
//
//	pkey ::= GROUP ELEMENT
//
//	GROUP ELEMENT ::= CHOICE {
//		ModqPElement	INTEGER
//		ECqPElement	OCTET STRING
//		Edwards25519Element	OCTET STRING
//	}
func (pkey *PublicKey) Marshal() (der asn_1.DER, err error) {
	der, err = pkey.Public().Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal public element: %v", err)
	}
	return
}

// ASN1UnmarshalPublicElement ASN.1 unmarshalls ElGamal public element.
func ASN1UnmarshalPublicElement(g group.Group, der asn_1.DER) (pub group.Element, err error) {
	pub, err = g.ElementOf(der)
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 unmarshal ElGamal public element: %v", err)
	}
	return
}

// PrivateKey is ElGamal private key.
type PrivateKey struct {
	// private is a private component of the key.
	private   *group.Scalar
	publicKey *PublicKey
}

// NewPrivateKey returns new ElGamal private key.
//
// s is considered to be a random scalar.
func NewPrivateKey(params *Parameters, s *group.Scalar) (*PrivateKey, error) {
	G := params.Group().Generator()

	// g {^,*} s
	Y, err := G.Scale(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create ElGamal public element: %v", err)
	}

	return &PrivateKey{
		private:   s,
		publicKey: NewPublicKey(params, Y),
	}, nil
}

// Parameters returns ElGamal private key parameters.
func (key *PrivateKey) Parameters() crypto.AlgorithmIdentifierParameters {
	return key.publicKey.parameters
}

// Fingerprint is unimplemented for ElGamal private key.
func (key *PrivateKey) Public() *PublicKey {
	return key.publicKey
}

// Fingerprint is unimplemented for ElGamal private key.
func (key *PrivateKey) Fingerprint() (uint64, error) {
	return 0, nil
}

// PublicKey returns ElGamal public key.
func (key *PrivateKey) EncryptionKey() crypto.EncryptionKey {
	return key.publicKey
}

// Decrypt decrypts ASN.1 marshalled ElGamal ciphertext into a group element.
func (key *PrivateKey) Decrypt(ciphertext []byte, checkDecodable bool) (crypto.Decryption, error) {
	ct, err := ASN1UnmarshalCiphertext(key.Public().Parameters().Group(), ciphertext)
	if err != nil {
		return nil, fmt.Errorf("cannot ASN.1 unmarshal ElGamal ciphertext: %v", err)
	}

	return decrypt(key.Parameters().Group(), ct.A(), key.private, ct, checkDecodable)
}

func (key *PublicKey) DecryptWithRandomness(r *group.Scalar, ciphertext []byte, checkDecodable bool) (crypto.Decryption, error) {
	ct, err := ASN1UnmarshalCiphertext(key.Parameters().Group(), ciphertext)
	if err != nil {
		return nil, fmt.Errorf("cannot ASN.1 unmarshal ElGamal ciphertext: %v", err)
	}

	gr, err := key.Parameters().Group().Generator().Scale(r)
	if err != nil {
		return nil, err
	}

	if err = ct.A().Equal(gr); err != nil {
		return nil, err
	}

	return decrypt(key.Parameters().Group(), key.Public(), r, ct, checkDecodable)
}

func decrypt(g group.Group, base group.Element, s *group.Scalar, ct *Ciphertext, checkDecodable bool) (crypto.Decryption, error) {
	// Y {^,*} x
	A, err := base.Scale(s)
	if err != nil {
		return nil, fmt.Errorf("cannot scale ElGamal public key element: %v", err)
	}

	// -A
	A = A.Inverse()

	// B {*,+} AInv
	B, err := ct.B().Op(A)
	if err != nil {
		return nil, fmt.Errorf("cannot recreate ElGamal ciphertext B element: %v", err)
	}

	if checkDecodable {
		if _, err = B.Decode(); err != nil {
			return nil, fmt.Errorf("ElGamal ciphertext B group element is not decodable: %v", err)
		}
	}

	return NewDecryption(g, B), nil
}

// ProvableDecrypt decrypts ASN.1 marshalled ElGamal ciphertext into a group element, and
// also provides proofs of correct decryption.
//
// s is considered to be a random scalar.
//
// See ProofChallenge for salt description.
func (key *PrivateKey) ProvableDecrypt(s *group.Scalar, ciphertext, salt []byte, checkDecodable bool) (dec crypto.Decryption, proof []byte, err error) {
	dec, err = key.Decrypt(ciphertext, checkDecodable)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt an ElGamal ciphertext: %v", err)
	}

	proof, err = key.Prove(s, ciphertext, dec, salt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create an ElGamal proof: %v", err)
	}
	return
}

// Prove proves that given ASN.1 marshalled ElGamal ciphertext and corresponding to it
// decryption are correct.
//
// s is considered to be a random scalar.
//
// See ProofChallenge for salt description.
func (key *PrivateKey) Prove(s *group.Scalar, ciphertext []byte, dec crypto.Decryption, salt []byte) ([]byte, error) {
	ct, err := ASN1UnmarshalCiphertext(key.Parameters().Group(), ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 unmarshal ElGamal ciphertext: %v", err)
	}

	// Create msg commitment a {^,*} s
	msgc, err := ct.A().Scale(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create message commitment: %v", err)
	}

	// Create key commitment g {^,*} s
	keyc, err := key.Parameters().Group().Generator().Scale(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create key commitment: %v", err)
	}

	// Create proof challenge (h) for verifier
	challenge, err := key.publicKey.ProofChallenge(ct, dec.Value(), msgc, keyc, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to create an ElGamal proof challenge: %v", err)
	}

	// r = h * x
	response, err := challenge.Mul(key.private)
	if err != nil {
		return nil, fmt.Errorf("failed to multiply ElGamal proof challenge by ElGamal private element: %v", err)
	}

	// r = r + s
	response, err = response.Add(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create an ElGamal proof response: %v", err)
	}

	der, err := NewDecryptionProof(ct, NewDecryption(key.Parameters().Group(), dec.Value()), nizkp.NewProof(challenge, response)).MarshalASN1()
	if err != nil {
		panic(err)
	}

	return der, nil
}

func (pkey *PublicKey) NewCommitments(g group.Group, opts []crypto.ProofOpts) ([]nizkp.Commitment, uint64, error) {
	var count uint64
	commitments := make([]nizkp.Commitment, len(opts))

	for i, opt := range opts {
		switch op := opt.(type) {
		case crypto.KeyPairProofOpts:
			commitments[i] = key.NewKeyPairCommitment(g, pkey.Public())
			count += commitments[i].Count()
		case crypto.CiphertextProofOpts:
			ct, err := ASN1UnmarshalCiphertext(g, op.Ciphertext())
			dec, err := ASN1UnmarshalDecryption(g, op.Decrypted())
			commitments[i], err = ciphertext.NewCiphertextCommitment(g, ct.A(), ct.B(), dec.Value())
			if err != nil {
				return nil, 0, fmt.Errorf("failed to create ciphertext commitments: %v", err)
			}
			count += commitments[i].Count()
		default:
			return nil, 0, fmt.Errorf("unsupported proof type")
		}
	}

	return commitments, count, nil
}

func (key *PrivateKey) ProofProve(s *group.Scalar, salt []byte, opts []crypto.ProofOpts) (der []byte, err error) {
	optz, count, err := key.Public().NewCommitments(key.Parameters().Group(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to prove key pair: %v", err)
	}

	// Size of array
	commitments := make([]group.Element, count)
	var added int

	for _, optzz := range optz {
		commit, err := optzz.Create(s)
		if err != nil {
			return nil, fmt.Errorf("failed to prove key pair: %v", err)
		}

		for j, commitment := range commit {
			commitments[added+j] = commitment
		}

		added++
	}

	// Create proof challenge (h) for verifier
	challenge, err := nizkp.Challenge(key.Parameters().Group(), commitments, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to create an ElGamal proof challenge: %v", err)
	}

	// r = h * x
	response, err := challenge.Mul(key.private)
	if err != nil {
		return nil, fmt.Errorf("failed to multiply ElGamal proof challenge by ElGamal private element: %v", err)
	}

	// r = r + s
	response, err = response.Add(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create an ElGamal proof response: %v", err)
	}

	t := nizkp.NewProof(challenge, response)

	der, err = t.MarshalASN1()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal key pair proof: %v", err)
	}
	return
}

func (pkey *PublicKey) ProofVerify(proof []byte, salt []byte, opts []crypto.ProofOpts) (err error) {
	//// decryption proof is a non-interactive zero-knowledge proof
	t, err := nizkp.ASN1UnmarshalProof(pkey.Parameters().Group(), proof)
	if err != nil {
		return fmt.Errorf("failed to ASN.1 unmarshal ElGamal decryption proof: %v", err)
	}

	optz, count, err := pkey.NewCommitments(pkey.Parameters().Group(), opts)
	if err != nil {
		return fmt.Errorf("failed to prove key pair: %v", err)
	}

	// Size of array
	commitments := make([]group.Element, count)
	var added int

	for _, optzz := range optz {
		commitmentz, err := optzz.Verify(t)
		if err != nil {
			return fmt.Errorf("failed to prove key pair: %v", err)
		}
		for j, commitment := range commitmentz {
			commitments[added+j] = commitment
		}
		added++
	}

	// Create proof challenge (h) for verifier
	challenge, err := nizkp.Challenge(pkey.Parameters().Group(), commitments, salt)
	if err != nil {
		return fmt.Errorf("failed to create an ElGamal proof challenge: %v", err)
	}

	// Recreated prover's challenge should be equal to the challenge that prover has actually calculated
	err = challenge.Equal(t.Challenge())
	if err != nil {
		return fmt.Errorf("recreated challenge on verifier side doesn't match the one generated on a prover side: %v", err)
	}
	return
}

// Marshal ASN.1 marshals ElGamal private key as
//
//	key ::= INTEGER
func (key *PrivateKey) Marshal() (der asn_1.DER, err error) {
	der, err = key.private.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal ElGamal private element: %v", err)
	}
	return
}

// ASN1UnmarshalPrivateElement ASN.1 unmarshalls ElGamal private element.
func ASN1UnmarshalPrivateElement(g group.Group, data []byte) (priv *group.Scalar, err error) {
	priv, err = group.UnmarshalScalar(data, g.Order())
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 unmarshal ElGamal private element: %v", err)
	}
	return
}
