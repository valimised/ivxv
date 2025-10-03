package homomorphic

import (
	"encoding/asn1"
	"reflect"
	"strconv"
	asn_11 "tivi.io/core/encoding/asn1"

	"golang.org/x/crypto/cryptobyte"
	asn_1 "golang.org/x/crypto/cryptobyte/asn1"
	"golang.org/x/crypto/sha3"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	"tivi.io/core/crypto/elgamal/distributed"
	"tivi.io/core/math/group"
)

// ElGamalHomomorphicEncryption is an implementation of crypto.HomomorphicEncryption.
//
// Both regular and distributed encryption schemes are supported under a single
// implementation.
type ElGamalHomomorphicEncryption struct {
	g         group.Group
	threshold uint64
}

// NewElGamalHomomorphicEncryption is a constructor for ElGamalHomomorphicEncryption.
//
// Don't initialize ElGamalHomomorphicEncryption as
//
//	enc := NewElGamalHomomorphicEncryption{}
//
// instead use this constructor
//
//	enc := NewElGamalHomomorphicEncryption()
//
// By doing so, you can be always sure, that code will not panic due to uninitialized
// fields.
//
// Also note, that ElGamalHomomorphicEncryption is not a pointer, why? That is
// because we don't modify ElGamalHomomorphicEncryption outside the
// NewElGamalHomomorphicEncryption constructor and, in addition to, it has
// struct fields that are either relatively small values or pointers as well.
func NewElGamalHomomorphicEncryption(g group.Group, threshold uint64) *ElGamalHomomorphicEncryption {
	return &ElGamalHomomorphicEncryption{
		g:         g,
		threshold: threshold,
	}
}

// Encrypt implements crypto.HomomorphicEncrypter.
//
// Encrypt uses public key pk to assist in homomorphic encryption of marks, it
// returns an DER marshalled Choice.
func (d *ElGamalHomomorphicEncryption) Encrypt(pk crypto.EncryptionKey, marks []bool) ([]byte, error) {
	elgamalPk, ok := pk.(*elgamal.PublicKey)
	if !ok {
		return nil, EncryptElGamalHomomorphicEncryptionCastTypeError{
			Expected: reflect.TypeOf(new(elgamal.PublicKey)),
			Got:      reflect.TypeOf(pk),
		}
	}

	choice, err := Encrypt(elgamalPk, marks)
	if err != nil {
		return nil, EncryptElGamalHomomorphicEncryptionError{Err: err}
	}

	choiceBytes, err := choice.Marshal()
	if err != nil {
		return nil, EncryptElGamalHomomorphicEncryptionASN1MarshalChoiceError{Err: err}
	}

	return choiceBytes, nil
}

// Aggregate implements crypto.HomomorphicAggregator.
//
// Aggregate will use homomorphic property of choices and sum them together into
// a resulting DER marshalled Choice. When withVerify is true, then each
// choices[i] will be verified. Verification of a choice consists of:
//
// a) check that each mark is either 0 or 1
//
// b) check that sum of all marks per this particular choice is 1
func (d *ElGamalHomomorphicEncryption) Aggregate(pk crypto.EncryptionKey, withVerify bool, choices ...[]byte) ([]byte, error) {
	elgamalPk, ok := pk.(*elgamal.PublicKey)
	if !ok {
		return nil, AggregateElGamalHomomorphicEncryptionCastTypeError{
			Expected: reflect.TypeOf(new(elgamal.PublicKey)),
			Got:      reflect.TypeOf(pk),
		}
	}

	aggregated, err := Aggregate(elgamalPk, withVerify, choices...)
	if err != nil {
		return nil, AggregateElGamalHomomorphicEncryptionError{Err: err}
	}

	aggregatedBytes, err := aggregated.Marshal()
	if err != nil {
		return nil, AggregateElGamalHomomorphicEncryptionASN1MarshalAggregatedError{Err: err}
	}

	return aggregatedBytes, nil
}

// Verify implements crypto.HomomorphicVerifier.
//
// Verify uses pk to verify a choice.
func (d *ElGamalHomomorphicEncryption) Verify(pk crypto.EncryptionKey, choice, extra []byte) error {
	err := pk.Verify(choice, extra)
	if err != nil {
		return VerifyElGamalHomomorphicEncryptionError{Err: err}
	}

	return nil
}

// Tally implements crypto.HomomorphicTallier.
//
// Privs and Pubs are intentionally passed as slices in order to support both
// regular and distributed encryption schemes, when with distributed you pass
// private and public key shares, whereas with distributed you only pass a
// single private and a public keys.
//
// Aggregated is an DER marshalled Choice, that has already been processed
// by Aggregate.
//
// MaxCount is an amount of iterations for DiscreteLog.
//
// Extra is bytes that we can add to the proof challenge (hash) creation in
// order to distinguish between different set of proofs, i.e. set of proofs,
// where identifier is hashed inside (in cryptography we call it
// "to add a salt").
//
// Tally returns an ordered slice (order is predefined on aggregation and is
// consistent due to Go slice order guarantees). Each element inside a slice
// has:
//
// a) count() - which represent how many times this candidate was marked
//
// b) Proofs() - which represent DER marshalled proofs of each mark for this candidate
func (d *ElGamalHomomorphicEncryption) Tally(privs []crypto.DecryptionKey, pubs []crypto.EncryptionKey, aggregated []byte, maxCount int, extra []byte) ([]crypto.HomomorphicDecryption, error) {
	choice, err := ASN1UnmarshalChoice(d.g, aggregated)
	if err != nil {
		return nil, TallyElGamalHomomorphicEncryptionASN1UnmarshalChoiceError{Err: err}
	}

	homomorphicDecrypted := make([]crypto.HomomorphicDecryption, len(choice.marks))
	proofsBytes := make([][]byte, len(privs))
	decryptedBytes := make([][]byte, len(privs))

	var combinedDecrypted crypto.Decryption

	for i, mark := range choice.marks {
		cipherBytes, err := mark.ciphertext.MarshalASN1()
		if err != nil {
			return nil, TallyElGamalHomomorphicEncryptionASN1MarshalCipherError{Err: err}
		}

		for j, priv := range privs {
			var decrypted crypto.Decryption
			rand, err := group.RandomScalar(d.g.Order())
			decrypted, proofsBytes[j], err = priv.ProvableDecrypt(rand, cipherBytes, extra, false)
			if err != nil {
				return nil, TallyElGamalHomomorphicEncryptionProvableDecryptError{Err: err}
			}

			decryptedBytes[j], err = decrypted.Marshal()
			if err != nil {
				return nil, TallyElGamalHomomorphicEncryptionASN1MarshalDecryptionError{Err: err}
			}

			err = pubs[j].Verify(proofsBytes[j], nil)
			if err != nil {
				return nil, TallyElGamalHomomorphicEncryptionVerifyError{Err: err}
			}
		}

		// This will have no effect for non-distributed encryption scheme
		decryptionCombiner := distributed.NewDecryptionShareCombiner(d.g, d.threshold)

		// This will have no effect for non-distributed encryption scheme.
		//
		// But for distributed one, it will combine all partial decryptions
		// into a single one. That single one can actually be decoded to a
		// meaningful plaintext.
		combinedDecrypted, err = decryptionCombiner.DecryptionSharesCombine(cipherBytes, decryptedBytes...)
		if err != nil {
			return nil, TallyElGamalHomomorphicEncryptionCombineDecryptedSharesError{Err: err}
		}

		// Find such exponent that can satisfy an equation:
		//	group^exponent = combinedDecrypted.value()
		// where group is a group generator
		exponent, err := DiscreteLog(d.g, maxCount, combinedDecrypted.Value())
		if err != nil {
			return nil, TallyDistributedPrivateKeyDiscreteLogError{Err: err}
		}

		homomorphicDecrypted[i] = &DistributedResult{
			count: uint64(exponent),
			proof: proofsBytes,
		}
	}

	return homomorphicDecrypted, nil
}

// Choice holds voter's encrypted preferences, or simply put, Choice is an
// ordered list of candidates (marks), where mark is a binary value 0 (candidate is
// not chosen) or 1 (candidate is chosen). Each mark has a proof, that confirms
// to the prover, that this mark can only be 0 or 1. Also, each choice has a
// range proof, which confirms to the prover, that sum of all marks is exactly 1.
// This means, that each choice could only have a single 1 mark and rest marks
// should be zeroes, i.e. {0, 0, 0, 1}, but not {0, 1, 0, 1}, and also not
// {0, 0, 0, 0}.
type Choice struct {
	// marks is a list of encrypted marks and proofs, which confirms to the
	// prover that a mark can only be 0 or 1.
	marks []*Mark
	// proof holds a range proof that confirms to the prover that sum of all
	// marks is exactly 1.
	proof *RangeProof
}

// NewChoice is a constructor for Choice.
//
// Don't initialize Choice as
//
//	choice := &nil
//
// instead use this constructor
//
//	choice := NewChoice()
//
// By doing so, you can be always sure, that code will not panic due to uninitialized
// fields.
func NewChoice(marks []*Mark, rangeProof *RangeProof) (*Choice, error) {
	if marks == nil {
		return nil, NewChoiceMarksIsNilError{}
	}

	return &Choice{
		marks: marks,
		proof: rangeProof,
	}, nil
}

// Verify verifies the correctness of a homomorphic Choice:
//
// a) check that each mark is either 0 or 1
//
// b) check that range proof is a sum of all marks
func (ch *Choice) Verify(pub *elgamal.PublicKey) (err error) {
	var markSumA = pub.Parameters().Group().Identity()
	var markSumB = pub.Parameters().Group().Identity()

	for i, mark := range ch.marks {
		// VerifyAll that mark is either 0 or 1
		if err = mark.Verify(pub); err != nil {
			return VerifyChoiceVerifyMarkError{Err: err, Index: strconv.Itoa(i)}
		}

		// Sum all mark ephemeral values, in terms of abstract algebra
		// we do following:
		// Element = Element {*} Element, where {*} is a group operation.
		//
		// As you can see the result of sum is always a group element
		markSumA, err = markSumA.Op(mark.ciphertext.A())
		if err != nil {
			return VerifyChoiceEphemeralOpError{Err: err, Index: strconv.Itoa(i)}
		}

		// Same for mark blinded message - sum all them up
		markSumB, err = markSumB.Op(mark.ciphertext.B())
		if err != nil {
			return VerifyChoiceBlindedMsgOpError{Err: err, Index: strconv.Itoa(i)}
		}
	}

	// Create a group element by given generator group and range proof.
	// Please note, that range proof is a sum of all marks, and generator
	// is a group element, that is used for generating any other group
	// element.
	// This actually means, that rangeProofElement is an element which is
	// obtained by summing all marks.
	rangeProofElement, err := pub.EncryptGroupElement(ch.proof.value, pub.Parameters().Group().Generator())
	if err != nil {
		return VerifyChoiceEncryptWithEncodedMsgAndEphemeralError{Err: err}
	}

	// Obviously, since range proof element is a sum of all marks per choice, then
	// ephemeral value of mark, should be the same as ephemeral value of a range
	// proof element.
	// NB! If it doesn't make sense, read this function comments again with
	// patience.
	err = rangeProofElement.A().Equal(markSumA)
	if err != nil {
		return VerifyChoiceEphemeralEqualError{Err: err}
	}

	// Same for blinded message of a mark, it should be the same as for range
	// proof element
	err = rangeProofElement.B().Equal(markSumB)
	if err != nil {
		return VerifyChoiceBlindedMsgEqualError{Err: err}
	}

	return nil
}

// ASN1Marshal marshals ch as
//
//	ch ::= SEQUENCE {
//		SEQUENCE {
//			ch.marks[0]
//			ch.marks[1]
//			...
//		}
//		ch.RangeProof
//	}
func (ch *Choice) Marshal() (asn_11.DER, error) {
	var err error

	var c cryptobyte.Builder

	rangeProofBytes, err := ch.proof.value.Marshal()
	if err != nil {
		return nil, ASN1MarshalChoiceASN1MarshalRangeProofError{Err: err}
	}

	marksBytes := make([][]byte, len(ch.marks))
	for i, mark := range ch.marks {
		marksBytes[i], err = mark.MarshalASN1()
		if err != nil {
			return nil, ASN1MarshalChoiceASN1MarshalMarkError{Err: err}
		}
	}

	c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
			for _, markBytes := range marksBytes {
				c.AddBytes(markBytes)
			}
		})
		c.AddBytes(rangeProofBytes)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalChoiceBytesError{Err: err}
	}

	return b, nil
}

// Mark is a candidate that can be either chosen (true) or not (false).
type Mark struct {
	// ciphertext holds the encrypted value of a mark.
	ciphertext *elgamal.Ciphertext
	// binaryProof holds the proof that the encrypted value is either 0 or 1.
	binaryProof *BinaryValueProof
}

// NewMark is a constructor for Mark.
//
// Don't initialize Mark as
//
//	mark := &Mark{}
//
// instead use this constructor
//
//	mark := NewMark()
//
// By doing so, you can be always sure, that code will not panic due to uninitialized
// fields.
func NewMark(pubkey *elgamal.PublicKey, value bool) (*Mark, *group.Scalar, error) {
	// M is a mark value, initialized as 0
	var M = group.ZeroScalar(pubkey.Parameters().Group().Order())
	if value {
		// This mark (candidate) has been chosen (1)
		M = group.OneScalar(pubkey.Parameters().Group().Order())
	}

	// group^M
	gM, err := pubkey.Parameters().Group().Generator().Scale(M)
	if err != nil {
		return nil, nil, NewMarkEncodeMessageError{Err: err}
	}

	R, err := group.RandomScalar(pubkey.Parameters().Group().Order())
	if err != nil {
		return nil, nil, NewMarkCreateEphemeralError{Err: err}
	}

	// Encrypt mark
	// ct = (ct1, ct2) = (generator^response, y^response generator^bigM)
	ct, err := pubkey.EncryptGroupElement(R, gM)
	if err != nil {
		return nil, nil, NewMarkEncryptWithEncodedMsgAndEphemeralError{Err: err}
	}

	// Create binary proof for a mark in order to later prove that
	// this mark has only either false or true
	proof, err := NewBinaryValueProof(pubkey, value, ct, R)
	if err != nil {
		return nil, nil, NewMarkNewBinaryValueProofError{Err: err}
	}

	return &Mark{
		ciphertext:  ct,
		binaryProof: proof,
	}, R, nil
}

// Verify verifies the correctness of binary value proof.
func (m *Mark) Verify(pub *elgamal.PublicKey) (err error) {
	var a, b [2]group.Element

	a[0], b[0], err = getPreCommitment(m, false, pub)
	if err != nil {
		return VerifyMarkGetPreCommitmentIfMarkIsNotChosenError{Err: err}
	}

	a[1], b[1], err = getPreCommitment(m, true, pub)
	if err != nil {
		return VerifyMarkGetPreCommitmentIfMarkIsChosenError{Err: err}
	}

	candidateChallenge, err := proofChallenge(pub.Parameters().Group(), pub, m.ciphertext, a, b)
	if err != nil {
		return VerifyMarkProofChallengeError{Err: err}
	}

	var challenge = group.ZeroScalar(pub.Parameters().Group().Order())
	challenge, err = challenge.Add(m.binaryProof.challenge0)
	if err != nil {
		return VerifyMarkChallengeAddBinaryProofChallenge0Error{Err: err}
	}

	challenge, err = challenge.Add(m.binaryProof.challenge1)
	if err != nil {
		return VerifyMarkChallengeAddBinaryProofChallenge1Error{Err: err}
	}

	if err = candidateChallenge.Equal(challenge); err != nil {
		return VerifyMarkEqualError{Err: err}
	}

	return nil
}

// ASN1Marshal marshals m as
//
//	m ::= SEQUENCE {
//		m.ciphertext
//		m.binaryProof
//	}
func (m *Mark) MarshalASN1() ([]byte, error) {
	cipherBytes, err := m.ciphertext.MarshalASN1()
	if err != nil {
		return nil, ASN1MarshalMarkASN1MarshalCipherError{Err: err}
	}

	binaryProofBytes, err := m.binaryProof.MarshalASN1()
	if err != nil {
		return nil, ASN1MarshalMarkASN1MarshalBinaryProofError{Err: err}
	}

	var c cryptobyte.Builder

	c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(cipherBytes)
		c.AddBytes(binaryProofBytes)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalMarkBytesError{Err: err}
	}

	return b, nil
}

type BinaryValueProof struct {
	challenge0 *group.Scalar
	response0  *group.Scalar
	challenge1 *group.Scalar
	response1  *group.Scalar
}

// NewBinaryValueProof represents OR proof.
// For a primer on OR composition of proof, see Section 5.2.4 of Berry Schoenmakers' lectures notes
// on cryptographic protocols: https://www.win.tue.nl/~berry/CryptographicProtocols/LectureNotes.pdf
func NewBinaryValueProof(opts *elgamal.PublicKey, value bool, ct *elgamal.Ciphertext, ephemeral *group.Scalar) (*BinaryValueProof, error) {
	// cast `value` to uint64 and big.Int for later use
	var M uint64
	bigM := group.ZeroScalar(opts.Parameters().Group().Order())
	// value=1 == voter has chosen this mark (candidate)
	if value {
		M = 1
		bigM = group.OneScalar(opts.Parameters().Group().Order())
	}

	// There are two of each variable:
	// one for the true Mark, and one for the opposite Mark.
	var a [2]group.Element // commitments (elements)
	var b [2]group.Element // commitments (elements)
	var c [2]*group.Scalar // challenges (scalars)
	var r [2]*group.Scalar // responses (scalars)

	// Sample the ephemeral proof randomness.
	s, err := group.RandomScalar(opts.Parameters().Group().Order())
	if err != nil {
		return nil, NewBinaryValueProofEphemeralProofError{Err: err}
	}

	// Compute the commitments for the case for which we have a witness.

	// a[M] = group^s
	a[M], err = opts.Parameters().Group().Generator().Scale(s)
	if err != nil {
		return nil, NewBinaryValueProofScaleGeneratorByEphemeralError{Err: err}
	}

	// b[M] = y^s
	b[M], err = opts.Public().Scale(s)
	if err != nil {
		return nil, NewBinaryValueProofScalePublicByEphemeralError{Err: err}
	}

	// For the ‘wrong’ case we do not have a witness.
	// We must therefore simulate the transcript for this case. The standard
	// technique is to take (or generate) a challenge, randomly sample a response,
	// and compute the commitment using the challenge and response such that the
	// simulated transcript would be accepting.

	// Compute a random challenge for opposite value proof.
	c[1-M], err = group.RandomScalar(opts.Parameters().Group().Order())
	if err != nil {
		return nil, NewBinaryValueProofChallenge1MError{Err: err}
	}

	// Sample a random response for the opposite value proof.
	r[1-M], err = group.RandomScalar(opts.Parameters().Group().Order())
	if err != nil {
		return nil, NewBinaryValueProofResponse1MError{Err: err}
	}

	// Compute the simulated commitments.
	// a must satisfy: a = group^r / ct1^c
	// b must satisfy:
	// - b = y^r / ct2^c if M = 1, or
	// - b = y^r / (ct2^c / group)^c if M = 0
	// We can combine both with: b = y^r / (ct2 / (group^(1-M))^c
	// and simplify it to b = y^r / ct2^c * (group^(1-M))^c

	// a[1-M] = group^r[1-M] / (ct1^c[1-M])

	// - group^r[1-M]
	aSim, err := opts.Parameters().Group().Generator().Scale(r[1-M])
	if err != nil {
		return nil, NewBinaryValueProofScaleGeneratorByResponse1MError{Err: err}
	}

	// - 1 / (ct1^c[1-M])
	tmp, err := ct.A().Scale(c[1-M])
	if err != nil {
		return nil, NewBinaryValueProofScaleCipherEphemeralByChallenge1MError{Err: err}
	}
	tmp = tmp.Inverse()

	// - group^r[1-M] / (ct1^c[1-M])
	a[1-M], err = aSim.Op(tmp)
	if err != nil {
		return nil, NewBinaryValueProofASimOpError{Err: err}
	}

	// b[1-M] = y^r[1-M] / ct2^c[1-M] * group^((1-M) * c[1-M])

	// - y^r[1-M]
	bSim, err := opts.Public().Scale(r[1-M])
	if err != nil {
		return nil, NewBinaryValueProofScalePublicByResponse1MError{Err: err}
	}

	// - 1 / (ct2^c[1-M])
	tmp, err = ct.B().Scale(c[1-M])
	if err != nil {
		return nil, NewBinaryValueProofScaleCipherBlindedMsgByChallenge1MError{Err: err}
	}
	tmp = tmp.Inverse()

	// - y^r[1-m] / (ct2^c[1-M])
	bSim, err = bSim.Op(tmp)
	if err != nil {
		return nil, NewBinaryValueProofBSimOpError{Err: err}
	}

	// - (1-M) * c[1-M]
	tmp2 := group.OneScalar(opts.Parameters().Group().Order())
	tmp2, err = tmp2.Sub(bigM)
	if err != nil {
		return nil, NewBinaryValueProofSubMError{Err: err}
	}

	tmp2, err = tmp2.Mul(c[1-M])
	if err != nil {
		return nil, NewBinaryValueProofMulByChallenge1MError{Err: err}
	}

	// - group^((1-M) * c[1-M])
	tmp, err = opts.Parameters().Group().Generator().Scale(tmp2)
	if err != nil {
		return nil, NewBinaryValueProofScaleGeneratorByTmpError{Err: err}
	}

	// - y^r[1-M] / (ct2^c[1-M]) * (group^(1-M))^c[1-M]
	b[1-M], err = bSim.Op(tmp)
	if err != nil {
		return nil, NewBinaryValueProofBSimOp2Error{Err: err}
	}

	// Compute the challenge for the OR-proof using the Fiat-Shamir heuristic.
	C, err := proofChallenge(opts.Parameters().Group(), opts, ct, a, b)
	if err != nil {
		return nil, NewBinaryValueProofChallengeError{Err: err}
	}

	// Compute the challenge for the correct case s.t. challenge = c[M] + c[1-M].
	// c[M] = (challenge - c[1-M]) % Q
	c[M], err = C.Sub(c[1-M])
	if err != nil {
		return nil, NewBinaryValueProofSubChallenge1MFromChallengeError{Err: err}
	}

	// Compute the response for the correct case.
	// r[M] = (s + c[M] * t) % Q, with t the encryption randomness.
	r[M], err = c[M].Mul(ephemeral)
	if err != nil {
		return nil, NewBinaryValueProofMulChallengeMByEphemeralError{Err: err}
	}

	r[M], err = r[M].Add(s)
	if err != nil {
		return nil, NewBinaryValueProofAddResponseMError{Err: err}
	}

	return &BinaryValueProof{
		challenge0: c[0], response0: r[0],
		challenge1: c[1], response1: r[1],
	}, nil
}

// ASN1Marshal marshals b as
//
//	b ::= SEQUENCE {
//		INTEGER
//		INTEGER
//		INTEGER
//		INTEGER
//	}
func (bvp *BinaryValueProof) MarshalASN1() ([]byte, error) {
	challenge0Bytes, err := bvp.challenge0.Marshal()
	if err != nil {
		return nil, ASN1MarshalBinaryValueProofASN1MarshalChallenge0Error{Err: err}
	}

	response0Bytes, err := bvp.response0.Marshal()
	if err != nil {
		return nil, ASN1MarshalBinaryValueProofASN1MarshalResponse0Error{Err: err}
	}

	challenge1Bytes, err := bvp.challenge1.Marshal()
	if err != nil {
		return nil, ASN1MarshalBinaryValueProofASN1MarshalChallenge1Error{Err: err}
	}

	response1Bytes, err := bvp.response1.Marshal()
	if err != nil {
		return nil, ASN1MarshalBinaryValueProofASN1MarshalResponse1Error{Err: err}
	}

	var c cryptobyte.Builder

	c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(challenge0Bytes)
		c.AddBytes(response0Bytes)
		c.AddBytes(challenge1Bytes)
		c.AddBytes(response1Bytes)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalBinaryValueProofBytesError{Err: err}
	}

	return b, nil
}

// RangeProof is a proof to confirm that sum of all marks per choice == 1.
type RangeProof struct {
	// value is the sum of ephemeral random values in the ciphertexts.
	value *group.Scalar
}

func NewRangeProof(value *group.Scalar) *RangeProof {
	return &RangeProof{value: value}
}

// Result holds a Count and decryption proof for a single Mark.
type Result struct {
	// Count is amount of a given Mark has been chosen,
	// e.g. choice={2, 4}, then mark(2) count=2, and mark(4) count=4
	count uint64
	// DecryptionProof holds the decryption proof that the Mark is correctly
	// computed.
	proof []*elgamal.DecryptionProof
}

// Count returns count.
func (r *Result) Count() uint64 {
	return r.count
}

// Proofs returns DER marshalled proofs.
// Amount of proofs == amount of marks for a given choice.
func (r *Result) Proofs() ([][]byte, error) {
	var err error

	proofsBytes := make([][]byte, len(r.proof))
	for i, proof := range r.proof {
		proofsBytes[i], err = proof.MarshalASN1()
		if err != nil {
			return nil, ProofsResultError{Err: err}
		}
	}

	return proofsBytes, nil
}

// ASN1Marshal marshals r to
//
//	r ::= SEQUENCE {
//		SEQUENCE {
//			r.proof[0]
//			r.proof[1]
//			...
//		}
//		INTEGER
//	}
//
// Amount of proofs == amount of marks for a given choice.
func (r *Result) MarshalASN1() ([]byte, error) {
	var err error

	var c cryptobyte.Builder

	proofBytes := make([][]byte, len(r.proof))
	for i, proof := range r.proof {
		proofBytes[i], err = proof.MarshalASN1()
		if err != nil {
			return nil, ASN1MarshalResultASN1MarshalProofError{Err: err}
		}
	}

	c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
			for _, proof := range proofBytes {
				c.AddBytes(proof)
			}
		})
		c.AddASN1Int64(int64(r.count))
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalResultBytesError{Err: err}
	}

	return b, nil
}

// DistributedResult is a Result of homomorphic tally.
type DistributedResult struct {
	// count is amount of a given Mark has been chosen,
	// e.g. choice={2, 4}, then mark(2) count=2, and mark(4) count=4
	count uint64
	// DecryptionProof holds the decryption proof that the Mark is correctly
	// computed.
	proof [][]byte
}

// Count returns count.
func (r *DistributedResult) Count() uint64 {
	return r.count
}

// Proofs returns DER marshalled proofs.
// Amount of proofs == amount of marks for a given choice.
func (r *DistributedResult) Proofs() ([][]byte, error) {
	var err error

	proofsBytes := make([][]byte, len(r.proof))
	for i, proof := range r.proof {
		proofsBytes[i] = proof
		if err != nil {
			return nil, ProofsDistributedResultError{Err: err}
		}
	}

	return proofsBytes, nil
}

// ASN1Marshal marshals r to
//
//	r ::= SEQUENCE {
//		SEQUENCE {
//			r.proof[0]
//			r.proof[1]
//			...
//		}
//		INTEGER
//	}
//
// Amount of proofs == amount of marks for a given choice.
func (r *DistributedResult) Marshal() (asn_11.DER, error) {
	var err error

	var c cryptobyte.Builder

	proofBytes := make([][]byte, len(r.proof))
	for i, proof := range r.proof {
		proofBytes[i] = proof
		if err != nil {
			return nil, ASN1MarshalDistributedResultASN1MarshalProofError{Err: err}
		}
	}

	c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
			for _, proof := range proofBytes {
				c.AddBytes(proof)
			}
		})
		c.AddASN1Int64(int64(r.count))
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalDistributedResultBytesError{Err: err}
	}

	return b, nil
}

// Encrypt returns encrypted marks (candidates) as a Choice.
func Encrypt(pubKey *elgamal.PublicKey, marks []bool) (*Choice, error) {
	// Range proof for each choice, should be 1
	rangeProof := NewRangeProof(group.ZeroScalar(pubKey.Parameters().Group().Order()))

	// Create new choice from all marks per voter
	choice, err := NewChoice(make([]*Mark, len(marks)), rangeProof)
	if err != nil {
		return nil, EncryptNewChoiceError{Err: err}
	}

	// Convert from bool mark to Mark
	for i, mark := range marks {
		var ephemeral *group.Scalar

		// Create new encrypted mark, which shows that voter has (true) or
		//  hasn't (false) chosen a candidate, as well as attach proofs to the
		// mark, showing, that value could only be either true or false
		choice.marks[i], ephemeral, err = NewMark(pubKey, mark)
		if err != nil {
			return nil, EncryptNewMarkError{Err: err}
		}

		// Sum range proof, remember, the final value should only be 1
		choice.proof.value, err = choice.proof.value.Add(ephemeral)
		if err != nil {
			return nil, EncryptRangeProofAddError{Err: err}
		}
	}

	return choice, nil
}

// Aggregate will sum up all choices into a single choice using homomorphic
// property of an encrypted choices.
//
// NB! Aggregate is a long-running process, and passing all choices at once
// could cost a lost (imagine if len(choices) == 1 million).
//
// Therefore, it is worth to perform Aggregate in batches.
func Aggregate(pubkey *elgamal.PublicKey, withVerify bool, choices ...[]byte) (*Choice, error) {
	// Nothing to Aggregate
	if len(choices) == 0 {
		return nil, AggregateChoicesSizeIsZeroError{}
	}

	// Summed up choices will be stored here
	aggregated := &Choice{}

	// len(marks) within each choices[i] should be the same
	choice0, err := ASN1UnmarshalChoice(pubkey.Parameters().Group(), choices[0])
	if err != nil {
		return nil, AggregateASN1UnmarshalChoice0Error{Err: err}
	}
	marksCount := len(choice0.marks)

	// Zero initialization
	aggregated.marks = make([]*Mark, marksCount)
	//for i := range aggregated.marks {
	//	aggregated.marks[i] = new(Mark)
	//	aggregated.marks[i].ciphertext = new(elgamal.ciphertext)
	//	aggregated.marks[i].binaryProof = new(BinaryValueProof)
	//	aggregated.marks[i].binaryProof.challenge0 = group.ZeroScalar(pubkey.Group().Group().Order())
	//	aggregated.marks[i].binaryProof.challenge1 = group.ZeroScalar(pubkey.Group().Group().Order())
	//	aggregated.marks[i].binaryProof.response0 = group.ZeroScalar(pubkey.Group().Group().Order())
	//	aggregated.marks[i].binaryProof.response1 = group.ZeroScalar(pubkey.Group().Group().Order())
	//}
	aggregated.proof = new(RangeProof)
	aggregated.proof.value = group.ZeroScalar(pubkey.Parameters().Group().Order())

	for i := range aggregated.marks {
		ciphertext, err := newEncryptedMessage(pubkey, pubkey.Parameters().Group(), pubkey.Parameters().Group().Identity(), group.ZeroScalar(pubkey.Parameters().Group().Order()))
		if err != nil {
			return nil, AggregateNewEncryptedMessageError{Err: err}
		}
		aggregated.marks[i] = new(Mark)
		aggregated.marks[i].ciphertext = ciphertext
		aggregated.marks[i].binaryProof = new(BinaryValueProof)
		aggregated.marks[i].binaryProof.challenge0 = group.ZeroScalar(pubkey.Parameters().Group().Order())
		aggregated.marks[i].binaryProof.challenge1 = group.ZeroScalar(pubkey.Parameters().Group().Order())
		aggregated.marks[i].binaryProof.response0 = group.ZeroScalar(pubkey.Parameters().Group().Order())
		aggregated.marks[i].binaryProof.response1 = group.ZeroScalar(pubkey.Parameters().Group().Order())
	}

	for _, ch := range choices {
		choice, err := ASN1UnmarshalChoice(pubkey.Parameters().Group(), ch)
		if err != nil {
			return nil, AggregateASN1UnmarshalChoiceError{Err: err}
		}

		// All choices should have exactly the same amount of marks
		if len(choice.marks) != marksCount {
			return nil, AggregateIncompatibleMarksSizeError{
				Expected: marksCount,
				Got:      len(choice.marks),
			}
		}

		// If choices are already aggregated then withVerify should be false
		// otherwise that verification will fail, since range proof expects
		// to be exactly = 1 (sum of all marks for a given choice == 1, and
		// this can only be true if we have choice={0,0,1}, but not {1,0,1})
		if withVerify {
			err = choice.Verify(pubkey)
			if err != nil {
				return nil, AggregateChoiceVerifyError{Err: err}
			}
		}

		for i := range aggregated.marks {
			ct1 := aggregated.marks[i].ciphertext
			ct2 := choice.marks[i].ciphertext

			a, err := ct1.A().Op(ct2.A())
			if err != nil {
				return nil, AggregateEphemeralOpError{Err: err}
			}

			b, err := ct1.B().Op(ct2.B())
			if err != nil {
				return nil, AggregateBlindedMessageOpError{Err: err}
			}

			aggregated.marks[i].ciphertext = elgamal.NewCiphertext(a, b)
		}
	}

	return aggregated, nil
}

// DiscreteLog is a brute force method to find an exponent of a decrypted
// element by looping over all non-negative whole numbers Z+, starting from 0,
// but not more iterations than maxCount.
func DiscreteLog(g group.Group, maxCount int, decrypted group.Element) (exponent int, err error) {
	for element := g.Identity(); exponent <= maxCount; exponent++ {
		if element.Equal(decrypted) == nil {
			// Found an exponent
			return exponent, nil
		}

		// element * group, i.e. try next group element
		element, err = element.Op(g.Generator())
		if err != nil {
			return 0, DiscreteLogOpError{Err: err}
		}
	}

	return 0, DiscreteLogOverflowError{Max: strconv.Itoa(maxCount)}
}

func ASN1UnmarshalChoice(g group.Group, ch []byte) (*Choice, error) {
	chBytes := cryptobyte.String(ch)

	var outerSequence, innerSequence, rangeProofBytes, privBytes cryptobyte.String

	if !chBytes.ReadASN1(&outerSequence, asn_1.SEQUENCE) {
		return nil, ASN1UnmarshalChoiceNotAnOuterSequenceError{}
	}

	if !outerSequence.ReadASN1(&innerSequence, asn_1.SEQUENCE) {
		return nil, ASN1UnmarshalChoiceNotAnInnerSequenceError{}
	}

	marks := make([]*Mark, 0)

	// TODO is it vulnerable to read in infinite for loop here,
	//  what about infinite large DER SEQUENCE?
	for innerSequence.ReadAnyASN1Element(&privBytes, nil) {
		mark, err := ASN1UnmarshalMark(g, privBytes)
		if err != nil {
			return nil, ASN1UnmarshalChoiceASN1UnmarshalMarkError{Err: err}
		}

		marks = append(marks, mark)
	}

	if !outerSequence.ReadAnyASN1Element(&rangeProofBytes, nil) {
		return nil, ASN1UnmarshalChoiceNoRangeProofError{}
	}

	rangeProof, err := group.UnmarshalScalar(asn_11.DER(rangeProofBytes), g.Order())
	if err != nil {
		return nil, ASN1UnmarshalChoiceASN1UnmarshalRangeProofError{Err: err}
	}

	if !chBytes.Empty() || !outerSequence.Empty() || !innerSequence.Empty() {
		return nil, ASN1UnmarshalChoiceTrailingBytesError{Err: err}
	}

	return &Choice{marks: marks, proof: NewRangeProof(rangeProof)}, nil
}

func ASN1UnmarshalMark(g group.Group, mk []byte) (*Mark, error) {
	mkBytes := cryptobyte.String(mk)

	var outerSequence, cipherBytes, binaryProofBytes cryptobyte.String

	if !mkBytes.ReadASN1(&outerSequence, asn_1.SEQUENCE) {
		return nil, ASN1UnmarshalMarkNotASequenceError{}
	}

	if !outerSequence.ReadAnyASN1Element(&cipherBytes, nil) {
		return nil, ASN1UnmarshalMarkNoCiphertextError{}
	}

	if !outerSequence.ReadAnyASN1Element(&binaryProofBytes, nil) {
		return nil, ASN1UnmarshalMarkNoBinaryProofError{}
	}

	if !mkBytes.Empty() || !outerSequence.Empty() {
		return nil, ASN1UnmarshalMarkTrailingBytesError{}
	}

	ct, err := elgamal.ASN1UnmarshalCiphertext(g, cipherBytes)
	if err != nil {
		return nil, ASN1UnmarshalMarkASN1UnmarshalCiphertextError{}
	}

	bp, err := ASN1UnmarshalBinaryValueProof(g, binaryProofBytes)
	if err != nil {
		return nil, ASN1UnmarshalMarkASN1UnmarshalBinaryValueProofError{}
	}

	return &Mark{ciphertext: ct, binaryProof: bp}, nil
}

func ASN1UnmarshalBinaryValueProof(g group.Group, binProof []byte) (*BinaryValueProof, error) {
	binProofBytes := cryptobyte.String(binProof)

	var outerSequence, challengeBytes0, responseBytes0, challenge1Bytes, response1Bytes cryptobyte.String

	if !binProofBytes.ReadASN1(&outerSequence, asn_1.SEQUENCE) {
		return nil, ASN1UnmarshalBinaryValueProofNotASequenceError{}
	}

	if !outerSequence.ReadAnyASN1Element(&challengeBytes0, nil) {
		return nil, ASN1UnmarshalBinaryValueProofASN1ReadChallenge0Error{}
	}

	if !outerSequence.ReadAnyASN1Element(&responseBytes0, nil) {
		return nil, ASN1UnmarshalBinaryValueProofASN1ReadResponse0Error{}
	}

	if !outerSequence.ReadAnyASN1Element(&challenge1Bytes, nil) {
		return nil, ASN1UnmarshalBinaryValueProofASN1ReadChallenge1Error{}
	}

	if !outerSequence.ReadAnyASN1Element(&response1Bytes, nil) {
		return nil, ASN1UnmarshalBinaryValueProofASN1ReadResponse1Error{}
	}

	if !binProofBytes.Empty() || !outerSequence.Empty() {
		return nil, ASN1UnmarshalBinaryValueProofTrailingBytesError{}
	}

	challenge0, err := group.UnmarshalScalar(asn_11.DER(challengeBytes0), g.Order())
	if err != nil {
		return nil, ASN1UnmarshalBinaryValueProofChallenge0Error{}
	}

	response0, err := group.UnmarshalScalar(asn_11.DER(responseBytes0), g.Order())
	if err != nil {
		return nil, ASN1UnmarshalBinaryValueProofResponse0Error{}
	}

	challenge1, err := group.UnmarshalScalar(asn_11.DER(challenge1Bytes), g.Order())
	if err != nil {
		return nil, ASN1UnmarshalBinaryValueProofChallenge1Error{}
	}

	response1, err := group.UnmarshalScalar(asn_11.DER(response1Bytes), g.Order())
	if err != nil {
		return nil, ASN1UnmarshalBinaryValueProofResponse1Error{}
	}

	return &BinaryValueProof{
		challenge0: challenge0,
		response0:  response0,
		challenge1: challenge1,
		response1:  response1,
	}, nil
}

func ASN1UnmarshalResult(g group.Group, res []byte) (*Result, error) {
	resBytes := cryptobyte.String(res)

	var err error
	var count int

	var outerSequence, innerSequence, proofsBytes cryptobyte.String

	if !resBytes.ReadASN1(&outerSequence, asn_1.SEQUENCE) {
		return nil, ASN1UnmarshalResultNotAnOuterSequenceError{}
	}

	if !outerSequence.ReadASN1(&innerSequence, asn_1.SEQUENCE) {
		return nil, ASN1UnmarshalResultNotAnInnerSequenceError{}
	}

	proofs := make([]*elgamal.DecryptionProof, 0)

	// Read infinitely from DER SEQUENCE until all proofs are read
	// TODO vulnerability? What if infinite amount of proofs are inside DER SEQUENCE
	for innerSequence.ReadAnyASN1Element(&proofsBytes, nil) {
		proof, err := elgamal.ASN1UnmarshalDecryptionProof(g, proofsBytes)
		if err != nil {
			return nil, ASN1UnmarshalResultASN1UnmarshalDecryptionProofShareError{Err: err}
		}

		proofs = append(proofs, proof)
	}

	if !outerSequence.ReadASN1Integer(&count) {
		return nil, ASN1UnmarshalResultASN1ReadCountError{Err: err}
	}

	if !resBytes.Empty() || !outerSequence.Empty() || !innerSequence.Empty() {
		return nil, ASN1UnmarshalResultTrailingBytesError{Err: err}
	}

	return &Result{count: uint64(count), proof: proofs}, nil
}

func ASN1UnmarshalDistributedResult(res []byte) (*DistributedResult, error) {
	resBytes := cryptobyte.String(res)

	var err error
	var count int

	var outerSequence, innerSequence, proofsBytes cryptobyte.String

	if !resBytes.ReadASN1(&outerSequence, asn_1.SEQUENCE) {
		return nil, ASN1UnmarshalDistributedResultNotAnOuterSequenceError{}
	}

	if !outerSequence.ReadASN1(&innerSequence, asn_1.SEQUENCE) {
		return nil, ASN1UnmarshalDistributedResultNotAnInnerSequenceError{}
	}

	proofs := make([][]byte, 0)
	// Read infinitely from DER SEQUENCE until all proofs are read
	// TODO vulnerability? What if infinite amount of proofs are inside DER SEQUENCE
	for innerSequence.ReadAnyASN1Element(&proofsBytes, nil) {
		proofs = append(proofs, proofsBytes)
	}

	if !outerSequence.ReadASN1Integer(&count) {
		return nil, ASN1UnmarshalDistributedResultASN1ReadCountError{Err: err}
	}

	if !resBytes.Empty() || !outerSequence.Empty() || !innerSequence.Empty() {
		return nil, ASN1UnmarshalDistributedResultTrailingBytesError{Err: err}
	}

	return &DistributedResult{count: uint64(count), proof: proofs}, nil
}

func newEncryptedMessage(pub *elgamal.PublicKey, g group.Group, message group.Element, ephemeral *group.Scalar) (*elgamal.Ciphertext, error) {
	// group^ephemeral
	a, err := g.Generator().Scale(ephemeral)
	if err != nil {
		return nil, NewEncryptedMessageScaleGeneratorByEphemeralError{Err: err}
	}

	// y^ephemeral
	b, err := pub.Public().Scale(ephemeral)
	if err != nil {
		return nil, NewEncryptedMessageScalePublicByEphemeralError{Err: err}
	}

	// b * message
	b, err = b.Op(message)
	if err != nil {
		return nil, NewEncryptedMessageOpError{Err: err}
	}

	ct := elgamal.NewCiphertext(a, b)

	return ct, nil
}

func getPreCommitment(mark *Mark, isChosen bool, pub *elgamal.PublicKey) (group.Element, group.Element, error) {
	var challenge, response, bigValue *group.Scalar

	// If this mark (candidate) is not chosen by a voter
	if !isChosen {
		bigValue = group.ZeroScalar(pub.Parameters().Group().Order())
		challenge = mark.binaryProof.challenge0
		response = mark.binaryProof.response0
	} else {
		bigValue = group.OneScalar(pub.Parameters().Group().Order())
		challenge = mark.binaryProof.challenge1
		response = mark.binaryProof.response1
	}

	// Action 1.
	//	a = (group^response) / (ct1^challenge))

	// group^response
	a, err := pub.Parameters().Group().Generator().Scale(response)
	if err != nil {
		return nil, nil, GetPreCommitmentScaleGeneratorByResponseError{Err: err}
	}

	// ephemeral^challenge
	tmp, err := mark.ciphertext.A().Scale(challenge)
	if err != nil {
		return nil, nil, GetPreCommitmentScaleEphemeralByChallengeError{Err: err}
	}

	// -tmp
	tmp = tmp.Inverse()

	// a * tmp
	// Note that tmp is now inverse, which means, for example, with multiplication
	// a / b == a * (1/b)
	a, err = a.Op(tmp)
	if err != nil {
		return nil, nil, GetPreCommitmentOp1Error{Err: err}
	}

	// Action 2.
	//	b = (y^response) / (blindedMsg^challenge) * group^(value * challenge)

	// y^response
	b, err := pub.Public().Scale(response)
	if err != nil {
		return nil, nil, GetPreCommitmentScalePublicByResponseError{Err: err}
	}

	// blindedMsg^challenge
	tmp, err = mark.ciphertext.B().Scale(challenge)
	if err != nil {
		return nil, nil, GetPreCommitmentScaleBlindedMsgByChallengeError{Err: err}
	}

	// -tmp
	tmp = tmp.Inverse()

	// b / tmp
	// Note that tmp is inverse
	b, err = b.Op(tmp)
	if err != nil {
		return nil, nil, GetPreCommitmentOp2Error{Err: err}
	}

	// value * challenge
	tmp2, err := bigValue.Mul(challenge)
	if err != nil {
		return nil, nil, GetPreCommitmentMulValueByChallengeError{Err: err}
	}

	// group^tmp2
	tmp, err = pub.Parameters().Group().Generator().Scale(tmp2)
	if err != nil {
		return nil, nil, GetPreCommitmentScaleGeneratorByTmpValueError{Err: err}
	}

	// b * tmp
	b, err = b.Op(tmp)
	if err != nil {
		return nil, nil, GetPreCommitmentOp3Error{Err: err}
	}

	return a, b, nil
}

// proofChallenge creates a hash as
//
//	h = SHAKE256(
//		BinaryValue" ||
//		g ||
//		pub ||
//		ct.A() ||
//		ct.B() ||
//		commitmentsA[0] ||
//		commitmentsA[1] ||
//		commitmentsB[0] ||
//		commitmentsB[1]
//	)
//
// and then produces a scalar from h as
//
//	C = h mod [0, g.Order())
func proofChallenge(g group.Group, pub *elgamal.PublicKey, ct *elgamal.Ciphertext, commitmentsA, commitmentsB [2]group.Element) (C *group.Scalar, err error) {
	gBytes, err := group.Marshal(g)
	if err != nil {
		return nil, ProofChallengeASN1MarshalGroupError{Err: err}
	}

	pkBytes, err := pub.Public().Marshal()
	if err != nil {
		return nil, ProofChallengeASN1MarshalPublicError{Err: err}
	}

	ephemeralBytes, err := ct.A().Marshal()
	if err != nil {
		return nil, ProofChallengeASN1MarshalEphemeralError{Err: err}
	}

	blindedMsgBytes, err := ct.B().Marshal()
	if err != nil {
		return nil, ProofChallengeASN1MarshalBlindedMsgError{Err: err}
	}

	A0Bytes, err := commitmentsA[0].Marshal()
	if err != nil {
		return nil, ProofChallengeASN1MarshalChallenge0Error{Err: err}
	}

	A1Bytes, err := commitmentsA[1].Marshal()
	if err != nil {
		return nil, ProofChallengeASN1MarshalChallenge1Error{Err: err}
	}

	B0Bytes, err := commitmentsB[0].Marshal()
	if err != nil {
		return nil, ProofChallengeASN1MarshalResponse0Error{Err: err}
	}

	B1Bytes, err := commitmentsB[1].Marshal()
	if err != nil {
		return nil, ProofChallengeASN1MarshalResponse1Error{Err: err}
	}

	seed := struct {
		Domain     string
		Params     []byte
		PubElement []byte
		Ephemeral  []byte
		BlindedMsg []byte
		A0         []byte
		A1         []byte
		B0         []byte
		B1         []byte
	}{Domain: "BinaryValue",
		Params:     gBytes,
		PubElement: pkBytes,
		Ephemeral:  ephemeralBytes,
		BlindedMsg: blindedMsgBytes,
		A0:         A0Bytes,
		A1:         A1Bytes,
		B0:         B0Bytes,
		B1:         B1Bytes,
	}

	seedBytes, err := asn1.Marshal(seed)
	if err != nil {
		return nil, ProofChallengeASN1MarshalSeedError{Err: err}
	}

	// Hash(seed bytes)
	reader := sha3.NewShake256()
	reader.Write(seedBytes) //nolint:errcheck

	C, err = group.ScalarValueOfReader(reader, g.Order())
	if err != nil {
		return nil, ProofChallengeScalarValueOfReaderError{Err: err}
	}

	return C, nil
}
