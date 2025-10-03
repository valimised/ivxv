package crypto

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/math/group"
)

var (
	// ElGamalEncryptionOID https://datatracker.ietf.org/doc/html/draft-rfced-info-pgutmann-00#section-2.
	ElGamalEncryptionOID = func() asn1.ObjectIdentifier { return asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3029, 2, 1} }
	// DistributedElGamalEncryptionOID is a custom OID, doesn't collide with existing OIDs.
	DistributedElGamalEncryptionOID = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 3029, 2, 1, 9999}
	// https://datatracker.ietf.org/doc/rfc8692/
	EcdsaWithSHAKE256 = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 6, 33}
)

// ASN1Marshaller is an interface for marshalling any data to ASN.1 bytes.
type ASN1Marshaller interface {
	// Marshal marshals this instance into ASN.1 bytes.
	Marshal() (asn_1.DER, error)
}

// AlgorithmIdentifierParameters is an interface for accessing key parameters' data.
type AlgorithmIdentifierParameters interface {
	// Group returns abstract group parsed from key parameters.
	Group() group.Group

	// Algorithm returns OID parsed from key parameters.
	Algorithm() asn1.ObjectIdentifier

	// ToPKIXAlgorithmIdentifier converts this key parameters into PKIX format.
	ToPKIXAlgorithmIdentifier() (pkix.AlgorithmIdentifier, error)
}

// KeyInfo is an interface for accessing key metadata.
type KeyInfo interface {
	ASN1Marshaller

	// Parameters returns this key parameters.
	Parameters() AlgorithmIdentifierParameters

	// Fingerprint returns this key fingerprint.
	Fingerprint() (uint64, error)
}

// EncryptionKey is an interface for a key that is used for data encryption.
type EncryptionKey interface {
	KeyInfo
	Encrypter
	EncryptionVerifier
	ProofVerifier
}

// Encrypter is an interface for plaintext encryption.
type Encrypter interface {
	KeyInfo

	// Encrypt uses random group scalar s to encrypt a plaintext.
	// Returns an ASN.1 marshalled ciphertext.
	Encrypt(rand *group.Scalar, plaintext []byte) ([]byte, error)
}

type DecrypterWithRandomness interface {
	KeyInfo

	DecryptWithRandomness(r *group.Scalar, ciphertext []byte, checkDecodable bool) (Decryption, error)
}

// EncryptionVerifier is an interface
type EncryptionVerifier interface {
	KeyInfo
	Verify(proof []byte, salt []byte) error
}

// DecryptionKey is an interface for a key that is used in ciphertext decryption.
type DecryptionKey interface {
	KeyInfo
	Decrypter
	Prover
	ProvableDecrypter
	ProofProver

	// EncryptionKey returns an encryption key that belongs to this decryption key.
	EncryptionKey() EncryptionKey
}

// Decryption is an interface for any decrypted data.
type Decryption interface {
	ASN1Marshaller

	// Value returns result of a decryption.
	Value() group.Element

	// Plaintext converts result of a decryption into UTF-8 byte slice.
	Plaintext() ([]byte, error)
}

// Decrypter is an interface for ciphertext decryption.
type Decrypter interface {
	KeyInfo

	// Decrypt decrypts a ciphertext. If checkDecodable is true then ciphertext
	// is also validated for decodability property, which allows to detect
	// malformed plaintexts. Use checkDecodable, when plaintext length
	// varies, otherwise keep it false (e.g. homomorphic encryption).
	Decrypt(ciphertext []byte, checkDecodable bool) (Decryption, error)
}

// ProvableDecrypter is an interface for ciphertext decryption with a proof of correct decryption.
type ProvableDecrypter interface {
	KeyInfo

	// ProvableDecrypt decrypts a ciphertext and provides a proof of correct decryption.
	// It uses cryptographic salt if there is a need to add uniqueness to the proof,
	// otherwise keep it nil. See Decrypter for details.
	ProvableDecrypt(s *group.Scalar, ciphertext []byte, salt []byte, checkDecodable bool) (dec Decryption, proof []byte, err error)
}

type Prover interface {
	KeyInfo
	Prove(rand *group.Scalar, ciphertext []byte, dec Decryption, salt []byte) (proof []byte, err error)
}

// ProofOpts is options that both prover and verifier should agree on in order
// to determine which set of proofs will be generated.
type ProofOpts any

// KeyPairProofOpts is a proof that ensures verifier that public key, used in vote encryption
// is actually a pair of a private key used in vote decryption.
type KeyPairProofOpts struct{}

func NewKeyPairProofOpts() KeyPairProofOpts {
	return KeyPairProofOpts{}
}

// CiphertextProofOpts is a proof that ensures verifier that ciphertext (encrypted vote) produced
// by public key is actually a decrypted vote produced by a private key.
type CiphertextProofOpts struct {
	// Ciphertext is an encrypted vote.
	ciphertext []byte

	// Decrypted is a decrypted vote.
	decrypted []byte
}

func NewCiphertextProofOpts(ciphertext []byte, decrypted []byte) CiphertextProofOpts {
	return CiphertextProofOpts{
		ciphertext: ciphertext,
		decrypted:  decrypted,
	}
}

func (cpo *CiphertextProofOpts) Ciphertext() []byte {
	return cpo.ciphertext
}

func (cpo *CiphertextProofOpts) Decrypted() []byte {
	return cpo.decrypted
}

// ProofProver proves to verifier different kind of proofs, governed by ProofOpts.
type ProofProver interface {
	// ProofProve proves to verifier different kind of proofs, governed by opts.
	// Salt is optional parameter and if not nil, then is used as cryptographic salt
	// to make a proof unique among others.
	ProofProve(rand *group.Scalar, salt []byte, opts []ProofOpts) (proof []byte, err error)
}

// ProofVerifier verifies prover's different kind of proofs, governed by ProofOpts.
type ProofVerifier interface {
	// ProofVerify verifiers prover's different kind of proofs, governed by opts.
	// As for ProofProver, salt is optional parameter and if not nil, then is used as
	// cryptographic salt to make a proof unique among others.
	ProofVerify(proof []byte, salt []byte, opts []ProofOpts) (err error)
}

// DecryptionSharesCombiner is an interface of distributed encryption scheme
// to combine all partial Decryptions into a single Decryption. Only after
// combining all Decryptions you may run Plaintext() on a resulting Decryption.
type DecryptionSharesCombiner interface {
	// DecryptionSharesCombine combines all Decryptions decs of a ciphertext
	// into a single Decryption.
	DecryptionSharesCombine(ciphertext []byte, decs ...[]byte) (Decryption, error)
}

type EncryptionPrivateKeyShare interface {
	DecryptionKey
	Index() *group.Scalar
	Share() *group.Scalar
}

// EncryptionPrivateKeyShareRegenerator is an interface of a distributed
// encryption to regenerate a single private key share.
type EncryptionPrivateKeyShareRegenerator interface {
	// RegeneratePrivateKeyShare regenerates a single private key share by only
	// index of the missing private key share and threshold amount of other
	// shares given.
	RegeneratePrivateKeyShare(index *group.Scalar, shares []EncryptionPrivateKeyShare) (EncryptionPrivateKeyShare, error)
}

// HomomorphicEncrypter is an interface for homomorphic encryption.
//
// In homomorphic encryption, a plaintext is called as marks. marks is an
// ordered list of candidates with a bool value representing is particular
// candidate is chosen or not e.g. {Alice, Bob, Eve} can be represented as
// {false, true, false}, if Bob is chosen. Mark is therefore is a single candidate.
//
// In homomorphic encryption, a ciphertext is a choice.
type HomomorphicEncrypter interface {
	Encrypt(pk EncryptionKey, marks []bool) (choice []byte, err error)
}

// HomomorphicAggregator is an interface for aggregating all choices into a single
// choice without decrypting each choice. It is possible due to homomorphic property
// of each choice.
//
// During aggregation, it is also possible to check each choice proofs:
//
//	a) range proof
//	b) mark proof
//
// Where range proof confirms that a single choice's marks' sum equals to 1
// (range proof), and each mark is either 0 or 1 (mark proof).
//
// After aggregation a single choice is sent to the HomomorphicTallier for a
// decryption.
type HomomorphicAggregator interface {
	// Aggregate sums up all DER marshalled choices into a single DER
	// marshalled aggregated choice.
	//
	// Use withVerify=true when you sum up choices that has not been aggregated yet,
	// and use withVerify=false otherwise. That is because aggregated choice
	// has already range proof > 1 and therefore verification of this proof
	// will fail.
	Aggregate(pk EncryptionKey, withVerify bool, choices ...[]byte) (aggregated []byte, err error)
}

// HomomorphicVerifier is an interface for verifying a choice proof.
type HomomorphicVerifier interface {
	Verify(pk EncryptionKey, proof []byte, salt []byte) error
}

// HomomorphicDecryption is a result of homomorphic tally.
//
// Result of a homomorphic tally is a decrypted choice. a choice holds
// aggregated marks inside. marks is an ordered list of candidates.
type HomomorphicDecryption interface {
	ASN1Marshaller

	// Count is an amount of a particular mark per a choice, e.g. suppose
	// we have candidate list (marks): list={Alice, Bob, Eve}. Then two voters
	// John and Michelle. They have voted as follows: John's is {true, false, false}
	// and Michelle's is {true, false, false}. After HomomorphicAggregator we
	// got {true(x2), false(x0), false(x0)}. After HomomorphicTallier we got
	// {2, 0, 0}, so Count() for a list[0] (candidate Alice) will be 2, and
	// for list[1] (candidate Bob) will be 0.
	Count() uint64

	// Proofs are all proofs per mark. Consider the same example in Count(),
	// the Proofs() for candidate Alice will be all proofs during HomomorphicTallier
	// for candidate Alice (list[0]).
	Proofs() (proofs [][]byte, err error)
}

// HomomorphicTallier is an interface for tallying choices, i.e. providing
// following info per a single choice:
//
//	a) how many times this choice has been chosen?
//	b) proofs of all marks for this choice
type HomomorphicTallier interface {
	// Tally decrypts a choice with attached DER marshalled proof for each
	// DecryptionKey sk, and verifies that proof with each EncryptionKey
	// pk, i.e. provable decryption is done by each sk[i] and proof is verified
	// by each pk[i].
	//
	// MaxCount is used to restrict iterations for HomomorphicDecryption.
	Tally(key []DecryptionKey, pkey []EncryptionKey, choice []byte, maxCount int, salt []byte) (tallies []HomomorphicDecryption, err error)
}

// HomomorphicEncryption is an interface to support homomorphic encryption
// for regular/distributed encryption schemes.
type HomomorphicEncryption interface {
	HomomorphicEncrypter
	HomomorphicAggregator
	HomomorphicVerifier
	HomomorphicTallier
}

type SignatureVerifier interface {
	KeyInfo
	Verify(rand *group.Scalar, signature, data []byte) error
}

type SigningPublicKey interface {
	KeyInfo
	SignatureVerifier
}

type Signer interface {
	KeyInfo
	Sign(rand *group.Scalar, data []byte) ([]byte, error)
}

type SigningPrivateKey interface {
	KeyInfo
	Signer
	PublicKey() SigningPublicKey
}
