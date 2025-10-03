package distributed

import (
	"crypto/sha256"
	asn_1 "encoding/asn1"
	"encoding/binary"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
	"math/big"
	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	"tivi.io/core/crypto/elgamal/nizkp"
	asn_11 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group"
	"tivi.io/core/math/polynomial"
)

// PublicKeyShare is distributed ElGamal public key share.
type PublicKeyShare struct {
	params *Parameters
	// EncryptionKey is a generated ElGamal public key for this public key share.
	//EncryptionKey elgamal.EncryptionKey
	public group.Element
}

func (pk *PublicKeyShare) ProofVerify(proof []byte, salt []byte, opts []crypto.ProofOpts) (err error) {
	return nil
}

func NewPublicKeyShare(params *Parameters, pubkey group.Element) *PublicKeyShare {
	return &PublicKeyShare{
		params: params,
		public: pubkey,
	}
}

func (pk *PublicKeyShare) Parameters() crypto.AlgorithmIdentifierParameters {
	return pk.params
}

// Fingerprint returns first 24 bytes of SHA256(pk).
func (pk *PublicKeyShare) Fingerprint() (uint64, error) {
	pubBytes, err := pk.public.Marshal()
	if err != nil {
		return 0, FingerprintPublicKeyShareError{Err: err}
	}

	sum := sha256.Sum256(pubBytes)
	return binary.BigEndian.Uint64(sum[24:]), nil
}

// Encrypt is unimplemented for distributed ElGamal, use elgamal.EncryptionKey instead.
func (pk *PublicKeyShare) Encrypt(_ *group.Scalar, _ []byte) (ciphertext []byte, err error) {
	return nil, EncryptPublicKeyShareUnimplementedError{}
}

// Verify verifies DER marshalled DecryptionShareProof proof.
//
// Extra is the same as in ProvableDecrypt.
func (pk *PublicKeyShare) Verify(proof, extra []byte) error {
	p, err := ASN1UnmarshalDecryptionProofShare(pk.params.g, proof)
	if err != nil {
		return VerifyPublicKeyShareASN1UnmarshalDecryptionProofShareError{Err: err}
	}

	err = Verify(pk, p, extra)
	if err != nil {
		return VerifyPublicKeyShareError{Err: err}
	}

	return nil
}

// ASN1Marshal as
//
//	pk ::= SEQUENCE {
//		INTEGER
//		pk.EncryptionKey
//	}
func (pk *PublicKeyShare) Marshal() (asn_11.DER, error) {
	pubKeyBytes, err := pk.public.Marshal()
	if err != nil {
		return nil, ASN1MarshalPublicKeyShareASN1MarshalPublicKeyError{Err: err}
	}

	return pubKeyBytes, nil
}

// PrivateKeyShare is distributed ElGamal private key share.
type PrivateKeyShare struct {
	// Index is an identifier of a polynomial that this private key share belongs to.
	//Index group.Scalar
	params *Parameters
	// PrivateKey is a generated ElGamal private key for this private key share.
	//PrivateKey elgamal.PrivateKey
	priv *group.Scalar
	// pkey is a public key share derived from this private key share.
	pkey *PublicKeyShare
}

func (p *PrivateKeyShare) Prove(rand *group.Scalar, ciphertext []byte, dec crypto.Decryption, salt []byte) (proof []byte, err error) {
	skk, err := elgamal.NewPrivateKey(elgamal.NewParameters(p.Parameters().Group()), p.priv)

	proof, err = skk.Prove(rand, ciphertext, dec, salt)
	if err != nil {
		return nil, ProvableDecryptImplPrivateKeyShareProveError{Err: err}
	}
	return proof, nil
}

func (p *PrivateKeyShare) ProofProve(rand *group.Scalar, salt []byte, opts []crypto.ProofOpts) (proof []byte, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *PrivateKeyShare) Index() *group.Scalar {
	return p.params.Index()
}

func (p *PrivateKeyShare) Share() *group.Scalar {
	return p.priv
}

// NewPrivateKeyShares returns new distributed ElGamal private key shares and
// ElGamal public key. NB! ElGamal public key is not distributed!
func NewPrivateKeyShares(g group.Group, parties, threshold uint64) ([]*PrivateKeyShare, *elgamal.PublicKey, error) {
	if threshold == 0 {
		return nil, nil, NewPrivateKeySharesThresholdIsZeroError{}
	}

	// Quorum check
	quorum := (parties / 2) + 1
	if threshold < quorum {
		return nil, nil, NewPrivateKeySharesThresholdNotInQuorumError{
			Parties:   parties,
			Threshold: threshold,
			Quorum:    quorum,
		}
	}

	privShares := make([]*PrivateKeyShare, parties)

	// Holds polynomials, the amount of polynomials stored in this variable
	// equals to parties amount
	privPreShares := make([]*preshare, parties)

	// Index has two purposes:
	//	a) polynomial identifier
	//	b) x coordinate of a point on polynomial
	indices := make([]*group.Scalar, parties)

	// shareevs, map's key is indices[i] and value - is evs[i], iteration is done
	// over privPreShares
	shareevs := make(map[string][]*preshareEvaluation)

	// Each commitments[i] is a list of
	// commitments[i] = [for coefficient in coefficients {group^(coefficient)}],
	// where coefficients are the coefficients of the given polynomial and
	// group is a group generator
	commitments := make([]*preshareCommitment, parties)

	// evs holds (x,y) coordinates of each polynomial. Amount of coordinates
	// per polynomial == parties amount, but to restore a polynomial it is enough
	// to have only threshold amount of (x,y) coordinates
	evs := make([][]*preshareEvaluation, parties)

	var err error

	// Create polynomials, amount of polynomials created == parties amount
	for i := range privPreShares {
		// Random Index will be used for each polynomial, to make it unique
		indices[i], err = group.RandomScalar(g.Order())
		if err != nil {
			return nil, nil, NewPrivateKeySharesCreateShareIndexError{Err: err}
		}
	}

	// Loop over all polynomials
	for i := range privPreShares {
		// Generate new N-degree polynomial, where N=threshold and assign Index
		// to it (to distinguish between different polynomials)
		privPreShares[i], err = randomPreshare(g, threshold, indices[i])
		if err != nil {
			return nil, nil, NewPrivateKeySharesGeneratePolynomialError{Err: err}
		}

		// Each commitment hold an array of
		// commitments[i] = [for coefficient in coefficients {group^(coefficient)}]
		// where coefficient is a polynomial coefficient
		commitments[i], err = commit(privPreShares[i])
		if err != nil {
			return nil, nil, NewPrivateKeySharesCommitError{Err: err}
		}

		// Get all (x,y) coordinates of a polynomial.
		//
		// Note, that each evs[i] will have exactly parties amount of (x,y)
		// coordinates. To restore a polynomial we only need threshold amount
		// of points, but here we get parties amount of them - that is what we
		// call a key shares.
		//
		// Indices are x coordinates of a polynomial. And as you can see, that
		// each polynomial will have exactly the same set of x coordinates
		evs[i], err = evaluate(privPreShares[i], indices...)
		if err != nil {
			return nil, nil, NewPrivateKeySharesEvaluateError{Err: err}
		}

		// Loop over all (x,y) coordinates of a polynomial.
		//
		// What we are doing here is following:
		//	a) Take X1 coordinate of a polynomial
		//	b) Set that X1 coordinate as a key inside a shareevs map
		//	c) Take (X1,Y1) coordinates and append it to the array.
		//
		// Note that each "privPreShares" iteration we will, here (inside evs[i]
		// loop) go exactly over the same set of x coordinates.
		//
		// Let me clarify what shareevs[i] will hold inside,
		// Suppose you have 2 polynomials, P1={(X1,Y1),(X2,Y2)} and
		// P2={(X1,Y11),(X2,Y22)}. Note, that X1 and X2 are the same for
		// both polynomials. Alright, then shareevs[i] will hold
		// shareevs[X1]={(X1,Y1),(X1,Y11)} and shareevs[X2]={(X2,Y2),(X2,Y22)}.
		//
		// Hope now it is clear
		for j, ev := range evs[i] {
			// Take latest (x,y) coordinates
			point := shareevs[ev.point.X.Value().String()]

			// Append new (x,y) coordinates to the latest
			point = append(point, evs[i][j])
			shareevs[ev.point.X.Value().String()] = point
		}
	}

	// Create private key shares - which is sum of all y coordinates of a
	// given x coordinate. Recall, each polynomial has exactly the same set of
	// x coordinates, but evaluations (y coordinates) are different.
	// So, we take x coordinate in all polynomials and sum up all y coordinate
	// values, i.e. P1=(x1,y1), P2=(x1,y2), P3=(x1,y3), then sum1=y1+y2+y3,
	// then we take P1=(x2,y1), P2=(x2,y2), P3=(x2,y3), and sum2=y1+y2+y3
	for i, preshare := range privPreShares {
		// privShare[i] = sum of all y coordinates
		privShares[i], err = privateKeyShareFromPreshare(g, preshare, commitments, shareevs[preshare.index.Value().String()])
		if err != nil {
			return nil, nil, NewPrivateKeySharesCreatePrivateKeyShareError{Err: err}
		}
	}

	// Generate ElGamal public key
	zeroIndex := group.ZeroScalar(g.Order())
	pk, err := generationPublicKeyShare(g, threshold, commitments, zeroIndex)
	if err != nil {
		return nil, nil, NewPrivateKeySharesCreatePublicKeyError{Err: err}
	}

	params := NewParameters(g, zeroIndex)
	pubkeyShare := NewPublicKeyShare(params, pk)

	elGamalParams := elgamal.NewParameters(g)
	pubkey := elgamal.NewPublicKey(elGamalParams, pubkeyShare.public)

	return privShares, pubkey, nil
}

// NewPrivateKeyShareWithSecret returns new PrivateKeyShare from provided secret.
func NewPrivateKeyShareWithSecret(params *Parameters, secret *group.Scalar) (*PrivateKeyShare, error) {
	elgamalParams := elgamal.NewParameters(params.Group())
	sk, err := elgamal.NewPrivateKey(elgamalParams, secret)
	if err != nil {
		return nil, NewPrivateKeyShareFromSecretError{Err: err}
	}

	pubShare := NewPublicKeyShare(params, sk.Public().Public())

	return &PrivateKeyShare{
		params: params,
		priv:   secret,
		pkey:   pubShare,
	}, nil
}

func (sk *PrivateKeyShare) Parameters() crypto.AlgorithmIdentifierParameters {
	return sk.params
}

// Fingerprint is unimplemented for sk.
func (sk *PrivateKeyShare) Fingerprint() (uint64, error) {
	return 0, nil
}

// Decrypt partially decrypts a ciphertext. Use Combiner for combining
// decrypted parts.
func (sk *PrivateKeyShare) Decrypt(ciphertext []byte, checkDecodable bool) (crypto.Decryption, error) {
	ct, err := elgamal.ASN1UnmarshalCiphertext(sk.params.g, ciphertext)
	if err != nil {
		return nil, DecryptPrivateKeyShareASN1UnmarshalCiphertextError{Err: err}
	}

	decryptedShare, err := Decrypt(sk, ct, checkDecodable)
	if err != nil {
		return nil, DecryptPrivateKeyShareError{Err: err}
	}

	return decryptedShare, nil
}

// ProvableDecrypt partially decrypts a ciphertext and provides a proof for
// that partial decryption.
//
// Extra is used as "salt" for proof challenge.
func (sk *PrivateKeyShare) ProvableDecrypt(y *group.Scalar, ciphertext, extra []byte, checkDecodable bool) (decryptedPart crypto.Decryption, proofForPart []byte, err error) {
	//ct, err := elgamal.ASN1UnmarshalCiphertext(sk.params.g, ciphertext)
	//if err != nil {
	//	return nil, nil, ProvableDecryptPrivateKeyShareASN1UnmarshalCiphertextError{Err: err}
	//}

	decryptedShare, proof, err := ProvableDecrypt(y, sk, ciphertext, extra, checkDecodable)
	if err != nil {
		return nil, nil, ProvableDecryptPrivateKeyShareError{Err: err}
	}

	proofBytes, err := proof.MarshalASN1()
	if err != nil {
		return nil, nil, ProvableDecryptPrivateKeyShareASN1MarshalProofError{Err: err}
	}

	return decryptedShare, proofBytes, nil
}

// PublicKey returns PublicKeyShare of a given PrivateKeyShare.
func (sk *PrivateKeyShare) EncryptionKey() crypto.EncryptionKey {
	return sk.pkey
}

// ASN1Marshal as
//
//	sk ::= SEQUENCE {
//		INTEGER
//		INTEGER
//	}
func (sk *PrivateKeyShare) Marshal() (asn_11.DER, error) {
	//indexBytes, err := sk.Index.Marshal()
	//if err != nil {
	//	return nil, ASN1MarshalPrivateKeyShareASN1MarshalIndexError{Err: err}
	//}

	//privBytes, err := sk.PrivateKey.Marshal()
	//if err != nil {
	//	return nil, ASN1MarshalPrivateKeyShareASN1MarshalPrivateKeyError{Err: err}
	//}
	//
	//var c cryptobyte.Builder
	//c.AddASN1(asn1.SEQUENCE, func(c *cryptobyte.Builder) {
	//	c.AddBytes(indexBytes)
	//	c.AddBytes(privBytes)
	//})

	//b, err := c.Bytes()
	//if err != nil {
	//	return nil, ASN1MarshalPrivateKeyShareError{Err: err}
	//}

	return sk.priv.Marshal()
}

// DecryptionShare is a partially decrypted value.
//
// DecryptionShare is not intended to be initialized as
//
//	DecryptionShare{}
//
// instead use a constructor
//
//	NewDecryptionShare()
//
// if you do so, you will prevent code panics, when some fields are uninitialized.
type DecryptionShare struct {
	index     *group.Scalar
	decrypted *elgamal.Decryption
}

func NewDecryptionShare(index *group.Scalar, decryption *elgamal.Decryption) *DecryptionShare {
	return &DecryptionShare{
		index:     index,
		decrypted: decryption,
	}
}

func (s DecryptionShare) Value() group.Element {
	return s.decrypted.Value()
}

func (s DecryptionShare) Plaintext() ([]byte, error) {
	return s.decrypted.Plaintext()
}

// ASN1Marshal as
//
//	s ::= SEQUENCE {
//		INTEGER
//		s.decrypted
//	}
func (s DecryptionShare) Marshal() (asn_11.DER, error) {
	indexBytes, err := s.index.Marshal()
	if err != nil {
		return nil, DecryptionShareIndexASN1MarshalError{Err: err}
	}

	proofBytes, err := s.decrypted.Value().Marshal()
	if err != nil {
		return nil, DecryptionShareProofASN1MarshalError{Err: err}
	}

	var c cryptobyte.Builder
	c.AddASN1(asn1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(indexBytes)
		c.AddBytes(proofBytes)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, DecryptionShareBytesError{Err: err}
	}

	return b, nil
}

// DecryptionShareProof is not intended to be initialized as
//
//	DecryptionShareProof{}
//
// instead use a constructor
//
//	NewDecryptionProofShare()
//
// if you do so, you will prevent code panics, when some fields are uninitialized.
type DecryptionShareProof struct {
	index *group.Scalar
	proof *elgamal.DecryptionProof
}

func NewDecryptionProofShare(index *group.Scalar, proof *elgamal.DecryptionProof) *DecryptionShareProof {
	return &DecryptionShareProof{
		index: index,
		proof: proof,
	}
}

// ASN1Marshal as
//
//	d ::= SEQUENCE {
//		INTEGER
//		d.proof
//	}
func (d *DecryptionShareProof) MarshalASN1() ([]byte, error) {
	indexBytes, err := d.index.Marshal()
	if err != nil {
		return nil, ASN1MarshalDecryptionProofShareASN1MarshalError{Err: err}
	}

	proofBytes, err := d.proof.MarshalASN1()
	if err != nil {
		return nil, ASN1MarshalDecryptionProofShareASN1MarshalProofError{Err: err}
	}

	var c cryptobyte.Builder
	c.AddASN1(asn1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(indexBytes)
		c.AddBytes(proofBytes)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalDecryptionProofShareBytesError{Err: err}
	}

	return b, nil
}

// Combiner combines all DecryptionShare into a single decrypted value.
//
// Combiner is not intended to be initialized as
//
//	Combiner{}
//
// instead use a constructor
//
//	NewDecryptionShareCombiner()
//
// if you do so, you will prevent code panics, when some fields are uninitialized.
type Combiner struct {
	g         group.Group
	threshold uint64
}

func NewDecryptionShareCombiner(g group.Group, threshold uint64) *Combiner {
	if threshold < 2 {
		threshold = 1
	}
	return &Combiner{
		g:         g,
		threshold: threshold,
	}
}

// DecryptionSharesCombine combines DER marshalled DecryptionShare
// sharesBytes into a single decrypted value.
func (c *Combiner) DecryptionSharesCombine(ciphertext []byte, sharesBytes ...[]byte) (crypto.Decryption, error) {
	// Skip if non-distributed encryption
	if len(sharesBytes) == 1 || c.threshold < 2 {
		decrypted, err := elgamal.ASN1UnmarshalDecryption(c.g, sharesBytes[0])
		if err != nil {
			return nil, DecryptionSharesCombineCombinerASN1UnmarshalDecryptionError{Err: err}
		}

		return decrypted, nil
	}

	ct, err := elgamal.ASN1UnmarshalCiphertext(c.g, ciphertext)
	if err != nil {
		return nil, DecryptionSharesCombineCombinerASN1UnmarshalCiphertextError{Err: err}
	}

	if len(sharesBytes) < int(c.threshold) {
		return nil, CombineDecryptedSharesSharesLessThanThresholdError{
			SharesCount: len(sharesBytes),
			Threshold:   int(c.threshold),
		}
	}

	shares := make([]*DecryptionShare, len(sharesBytes))
	indices := make([]*group.Scalar, len(shares))
	for i, share := range sharesBytes {
		shares[i], err = ASN1UnmarshalDecryptionShare(c.g, share)
		if err != nil {
			return nil, DecryptionSharesCombineCombinerASN1UnmarshalDecryptionShareError{Err: err}
		}
		indices[i] = shares[i].index
	}

	zero := group.ZeroScalar(c.g.Order())
	denom := c.g.Identity()
	var tmp group.Element

	for _, share := range shares {
		s, err := polynomial.LagrangeBasisPolynomial(c.g, zero, indices, share.index)
		if err != nil {
			return nil, DecryptionSharesCombineCombinerLagrangeBasisPolynomialError{Err: err}
		}

		tmp, err = share.Value().Scale(s)
		if err != nil {
			return nil, DecryptionSharesCombineCombinerShareScaleError{Err: err}
		}

		denom, err = denom.Op(tmp)
		if err != nil {
			return nil, DecryptionSharesCombineCombinerOpError{Err: err}
		}

	}

	tmp2, err := ct.B().Op(denom.Inverse())
	if err != nil {
		return nil, DecryptionSharesCombineCombinerBlindedMessageOpError{Err: err}
	}

	decrypted := elgamal.NewDecryption(c.g, tmp2)

	return decrypted, nil
}

// Regenerator regenerates a private key share by only share Index given.
type Regenerator struct {
	g group.Group
	//group     group.group
	threshold uint64
}

func NewRegenerator(g group.Group, threshold uint64) *Regenerator {
	return &Regenerator{g: g, threshold: threshold}
}

func (r *Regenerator) RegeneratePrivateKeyShare(index *group.Scalar, privKeySharesBytes []crypto.EncryptionPrivateKeyShare) (crypto.EncryptionPrivateKeyShare, error) {
	var err error

	// DER unmarshalled private key shares
	privKeyShares := make([]*PrivateKeyShare, len(privKeySharesBytes))

	for i, privKeyShareBytes := range privKeySharesBytes {

		//ska, err := UnmarshalPrivateElementShare(r.g, privKeyShareBytes.Share())
		eparams := elgamal.NewParameters(r.g)
		params := NewParameters(r.g, privKeyShareBytes.Index())
		skz, err := elgamal.NewPrivateKey(eparams, privKeyShareBytes.Share())
		privKeyShares[i] = &PrivateKeyShare{
			params: params,
			priv:   privKeyShareBytes.Share(),
			pkey:   NewPublicKeyShare(params, skz.Public().Public()),
		}
		if err != nil {
			return nil, RegeneratePrivateKeyShareByIndexASN1UnmarshalPrivateKeyShareError{Err: err}
		}
	}

	regblinds := make([][]*regenerationBlindShare, r.threshold)

	// Loop exactly over #threshold amount of shares
	for i := uint64(0); i < r.threshold; i++ {
		regblinds[i], err = randomRegenerationBlindShares(r.g, r.threshold, privKeyShares[i].params.Index(), privKeyShares[:r.threshold])
		if err != nil {
			return nil, RegeneratePrivateKeyShareByIndexImplRandomRegenerateBlindSharesError{Err: err}
		}
	}

	recvregblinds := make([][]*regenerationBlindShare, r.threshold)
	for i := uint64(0); i < r.threshold; i++ {
		recvregblinds[i] = make([]*regenerationBlindShare, r.threshold)
	}

	for i := uint64(0); i < r.threshold; i++ {
		for j := uint64(0); j < r.threshold; j++ {
			recvregblinds[j][i] = regblinds[i][j]
		}
	}

	regshares := make([]*regenerationShare, r.threshold)

	for i := uint64(0); i < r.threshold; i++ {
		regshares[i], err = regenerationFromShare(r.g, r.threshold, privKeyShares[i], index, recvregblinds[i])
		if err != nil {
			return nil, RegeneratePrivateKeyShareByIndexImplRegenerationFromShareError{Err: err}
		}
	}

	regeneratedShare, err := combineRegShares(r.g, r.threshold, regshares)
	if err != nil {
		return nil, RegeneratePrivateKeyShareByIndexImplCombineRegSharesError{Err: err}
	}

	return regeneratedShare, nil
}

// preshare is the private input of a party taking part in a distributed Elgamal
// key generation.
//
// Or simply put, preshare is a polynomial with identifier shareIndex.
type preshare struct {
	index *group.Scalar
	//params Parameters
	poly *polynomial.Polynomial
}

// preshareCoefficientCommitment is a commitment to a coefficient of a preshare.
type preshareCoefficientCommitment struct {
	// commitmentIndex is a value from [0, polynomial degree+1).
	commitmentIndex uint64
	// commitment is a value of
	//	group {^,*} (polynomial coefficient)
	// where group is a group generator
	commitment group.Element
}

// preshareCommitment is a commitment to a preshare.
type preshareCommitment struct {
	// shareIndex has two purposes:
	//	a) serve as identifier for a polynomial (distinguish)
	//	b) x coordinate of all polynomials. So, shareIndex will present in
	// each polynomial as x coordinate
	shareIndex *group.Scalar
	// coefficientCommitments are all preshareCoefficientCommitment for
	// a given polynomial
	coefficientCommitments []*preshareCoefficientCommitment
}

// preshareEvaluation is an evaluation of the polynomial representing the preshare.
type preshareEvaluation struct {
	// shareIndex has two purposes:
	//	a) serve as identifier for a polynomial (distinguish)
	//	b) x coordinate of all polynomials. So, shareIndex will present in
	// each polynomial as x coordinate
	shareIndex *group.Scalar
	// point is polynomial point (x,y) coordinates, where x == shareIndex.
	point *polynomial.Point
}

// regenerationShare is the regeneration share for recovering a lost private keyshare.
type regenerationShare struct {
	// shareIndex is the Index of the party which generated this share
	shareIndex *group.Scalar
	// regenerationIndex is the Index of the party recovering the private keyshare
	regenerationIndex *group.Scalar
	// share is the value of the regeneration share.
	share *group.Scalar
}

// regenerationBlindShare is the share of the additive secret share used to
// blind the regeneration share.
type regenerationBlindShare struct {
	// shareIndex is the Index of the party which generated this share
	shareIndex *group.Scalar
	// targetIndex is the Index of the intended recipient of this share
	targetIndex *group.Scalar
	// blindShare is the value of the this share.
	blindShare *group.Scalar
}

// commit constructs a commitment to this preshare.
func commit(s *preshare) (*preshareCommitment, error) {
	coms := make([]*preshareCoefficientCommitment, s.poly.Degree()+1)

	// Loop over all polynomial coefficients
	// and do coms[i] = group^(polynomial coefficient)
	for i := uint64(0); i < s.poly.Degree()+1; i++ {
		gen := s.poly.Group.Generator()

		// com = group^(polynomial coefficient)
		com, err := gen.Scale(s.poly.Coefficient(i))
		if err != nil {
			return nil, CommitScaleCoefficientError{Err: err}
		}

		// i is a polynomial coefficient Index, i.e. poly coefficient {a, b, C},
		// then a Index is 0, b Index is 1, C Index is 2 (just like in an array)
		coms[i] = &preshareCoefficientCommitment{
			commitmentIndex: i,
			commitment:      com,
		}
	}

	// s is a random scalar that we use to distinguish polynomials between each
	// other
	return &preshareCommitment{
		shareIndex:             s.index,
		coefficientCommitments: coms,
	}, nil
}

// evaluate evaluates (gets (x,y) coordinates) an s polynomial (preshare).
//
// Indicies is a list of x coordinates that we evaluate s polynomial for.
func evaluate(s *preshare, indices ...*group.Scalar) ([]*preshareEvaluation, error) {
	evs := make([]*preshareEvaluation, len(indices))

	// Index is a polynomial identifier to distinguish between each other
	for i := range indices {
		// Get (Index,y) coordinate of a point on a polynomial, where x coordinate
		// is given x=Index.
		point, err := s.poly.Evaluate(indices[i])
		if err != nil {
			return nil, EvaluatePolynomialError{Err: err}
		}

		// From now on, Index is used both to:
		//	a) distinguish between different polynomials
		//	b) as x coordinate of a point on a polynomial
		evs[i] = &preshareEvaluation{
			shareIndex: s.index,
			point:      point,
		}
	}

	return evs, nil
}

// privateKeyShareFromPreshare verifies that the preshare commitments correspond to preshare
// evaluations and returns the private key share.
func privateKeyShareFromPreshare(g group.Group, s *preshare, commits []*preshareCommitment, evs []*preshareEvaluation) (*PrivateKeyShare, error) {
	err := shareVerify(g, s, commits, evs)
	if err != nil {
		return nil, PrivateKeyShareFromPreshareShareVerifyError{Err: err}
	}

	y := group.ZeroScalar(g.Order())

	// Loop over all polynomials points (note that x coordinates are the same
	// for each polynomial)
	for _, e := range evs {
		// Sum all y coordinates of all polynomials
		y, err = y.Add(e.point.Y)
		if err != nil {
			return nil, PrivateKeyShareFromPreshareShareAddError{Err: err}
		}
	}

	// Return ElGamal private key share by given secret
	params := NewParameters(g, s.index)
	privkey, err := NewPrivateKeyShareWithSecret(params, y)
	if err != nil {
		return nil, PrivateKeyShareFromPreshareNewPrivateKeyShareFromSecretError{Err: err}
	}

	return privkey, nil
}

func shareVerify(g group.Group, preshare *preshare, commits []*preshareCommitment, shareevs []*preshareEvaluation) error {
	// shareevs[i] holds all (x,y) coordinates of all polynomials (preshares),
	// where x == preshare.shareIndex,
	// So, e.g. suppose X1 == preshare.shareIndex, then
	// shareevs[X1] = {(X1,Y1), (X1,Y2), (X1,Y3)}
	for _, shareev := range shareevs {
		// There should not be any other than preshare.shareIndex x coordinates
		// of a polynomial
		err := shareev.point.X.Equal(preshare.index)
		if err != nil {
			return ShareVerifyPreshareEqualError{Err: err}
		}
	}

	for _, cms := range commits {
		// Amount of commitments should be the same as polynomial degree+1
		if len(cms.coefficientCommitments) != int(preshare.poly.Degree())+1 {
			return ShareVerifyCommitmentsAmountNotEqualToPolynomialDegreeError{
				Expected: int(preshare.poly.Degree()) + 1,
				Got:      len(cms.coefficientCommitments),
			}
		}
	}

	for _, ev := range shareevs {
		gen := g.Generator()

		// group {*,^} point.y
		lhs, err := gen.Scale(ev.point.Y)
		if err != nil {
			return ShareVerifyScaleGeneratorByPolynomialPointYValueError{Err: err}
		}

		cmap := make(map[uint64]*preshareCoefficientCommitment)
		// Copy coefficient commitments by shareIndex into a tmp "cmap" map
		for _, cms := range commits {
			if cms.shareIndex.Equal(ev.shareIndex) == nil {
				for _, cm := range cms.coefficientCommitments {
					cmap[cm.commitmentIndex] = cm
				}
			}
		}

		// Amount of cmap keys should be the same as polynomial degree+1
		if len(cmap) != int(preshare.poly.Degree())+1 {
			return ShareVerifyCommitmentsAmountError{
				Expected: int(preshare.poly.Degree()) + 1,
				Got:      len(cmap),
			}
		}

		rhs := g.Identity()

		exponent := group.ZeroScalar(g.Order())

		var tmp group.Element

		for _, cm := range cmap {
			// exponent = preshare.shareIndex^commitmentIndex, where
			// preshare.shareIndex is polynomial x coordinate, and
			// commitmentIndex is number from [0, polynomial degree+1)
			exponent = preshare.index.Exp(new(big.Int).SetUint64(cm.commitmentIndex))

			// tmp is a group element
			tmp, err = cm.commitment.Scale(exponent)
			if err != nil {
				return ShareVerifyScaleCommitmentError{Err: err}
			}

			// rhs is next group element (right after the tmp)
			rhs, err = rhs.Op(tmp)
			if err != nil {
				return ShareVerifyOpCommitmentError{Err: err}
			}
		}

		err = lhs.Equal(rhs)
		if err != nil {
			return ShareVerifyEqualError{Err: err}
		}
	}

	return nil
}

// randomPreshare generates new threshold degree random polynomial.
func randomPreshare(g group.Group, threshold uint64, index *group.Scalar) (*preshare, error) {
	// Random polynomial of N degree, where N=threshold
	pol, err := polynomial.NewRandom(g, threshold-1)
	if err != nil {
		return nil, RandomPreshareNewPolynomialError{Err: err}
	}

	return &preshare{index: index, poly: pol}, nil
}

// generationPublicKeyShare returns ElGamal public key that is generated from
// commitments.
func generationPublicKeyShare(g group.Group, threshold uint64, commits []*preshareCommitment, index *group.Scalar) (group.Element, error) {
	res := g.Identity()

	for _, cms := range commits {
		found := make(map[uint64]struct{})

		exp := group.ZeroScalar(g.Order())

		var tmp group.Element

		var err error

		// Loop over all polynomial group^coefficient values
		for _, cm := range cms.coefficientCommitments {
			found[cm.commitmentIndex] = struct{}{}

			// exp is number from [0...Polynomial degree+1)
			exp = index.Exp(new(big.Int).SetUint64(cm.commitmentIndex))

			// tmp = (group^coefficient)^exp
			tmp, err = cm.commitment.Scale(exp)
			if err != nil {
				return nil, GenerationPublicKeyShareCommitmentScaleError{Err: err}
			}

			// Each iteration do res = res {*,+} tmp
			res, err = res.Op(tmp)
			if err != nil {
				return nil, GenerationPublicKeyShareCommitmentOpError{Err: err}
			}
		}

		// Check that iterations amount == threshold
		if len(found) != int(threshold) {
			return nil, GenerationPublicKeyShareNotEnoughCommitmentsError{
				Expected: int(threshold),
				Got:      len(found),
			}
		}
	}

	// Create new ElGamal public key, where public element is res
	//pubkey := elgamal.NewPublicKey(g, res)

	return res, nil
}

// RandomRegenerationBlindShares returns the shares for additive secret shares.
// The argument Index is the Index of the party calling the method and
// targets is a list of ShareIndices for whom to generate the shares.
//
// The size of argument target must be at least threshold and the elements must
// be unique. For statelessness, Index must be included in targets. The method
// ensures that the sum of the shares is 0.
//
// Every returned shares must be transmitted to the party indicated by
// targetIndex in an authenticated and encrypted channel.
func randomRegenerationBlindShares(g group.Group, threshold uint64, index *group.Scalar, shares []*PrivateKeyShare) ([]*regenerationBlindShare, error) {
	// Shares should be >= threshold
	if len(shares) < int(threshold) {
		return nil, RandomRegenerationBlindSharesSharesLessThanThresholdError{
			Shares:    len(shares),
			Threshold: int(threshold),
		}
	}

	var seen bool

	// Index, i.e. x coordinate of polynomial, must present in a share,
	// otherwise that Index is outside the polynomial
	//
	// Loop over all polynomials and seek for the x coordinate (Index)
	for _, target := range shares {
		if target.params.Index().Equal(index) == nil {
			seen = true
		}
	}

	// Index is not found in any polynomial
	if !seen {
		return nil, RandomRegenerationBlindSharesIndexNotInSharesError{}
	}

	blindShares := make([]*regenerationBlindShare, len(shares))

	blind1 := group.ZeroScalar(g.Order())

	for i := 0; i < len(shares)-1; i++ {
		blind2, err := group.RandomScalar(g.Order())
		if err != nil {
			return nil, RandomRegenerationBlindSharesRandomScalarError{Err: err}
		}

		blindShares[i] = &regenerationBlindShare{
			shareIndex:  index,
			targetIndex: shares[i].params.Index(),
			blindShare:  blind2,
		}

		// Sum up all random scalars
		blind1, err = blind1.Add(blind2)
		if err != nil {
			return nil, RandomRegenerationBlindSharesAddError{Err: err}
		}
	}

	blindShares[len(shares)-1] = &regenerationBlindShare{
		shareIndex:  index,
		targetIndex: shares[len(shares)-1].params.Index(),
		blindShare:  blind1.Negate(),
	}

	return blindShares, nil
}

func combineRegenerationBlindShares(g group.Group, threshold uint64, index *group.Scalar, shares []*regenerationBlindShare) (*group.Scalar, []*group.Scalar, error) {
	// Shares must be >= threshold
	if len(shares) < int(threshold) {
		return nil, nil, CombineRegenerationBlindSharesSharesLessThanThresholdError{
			Expected: int(threshold),
			Got:      len(shares),
		}
	}

	var seenIndices []*group.Scalar

	for _, share := range shares {
		for _, seen := range seenIndices {
			err := share.shareIndex.Equal(seen)
			if err == nil {
				// Multiple shares from same Index
				return nil, nil, CombineRegenerationBlindSharesShareAlreadySeenError{}
			}
		}

		seenIndices = append(seenIndices, share.shareIndex)
	}

	blind := group.ZeroScalar(g.Order())

	var err error
	for _, share := range shares {
		// Each share should be exactly of the Index, i.e.
		// x coordinate of share == Index
		err = share.targetIndex.Equal(index)
		if err != nil {
			return nil, nil, CombineRegenerationBlindSharesIndexAlreadyInsideSharesError{}
		}

		// Sum up all blinds
		blind, err = blind.Add(share.blindShare)
		if err != nil {
			return nil, nil, CombineRegenerationBlindSharesAddBlindShareError{Err: err}
		}
	}

	return blind, seenIndices, nil
}

// regenerationFromShare uses the private keyshare sk of the party and received
// shares to generate the regeneration share for the party indicated by
// newIndex. The returned regeneration share must be transmitted to the party
// with newIndex over an authenticated and encrypted channel.
func regenerationFromShare(g group.Group, threshold uint64, sk *PrivateKeyShare, newIndex *group.Scalar, shares []*regenerationBlindShare) (*regenerationShare, error) {
	blind, indices, err := combineRegenerationBlindShares(g, threshold, sk.params.Index(), shares)
	if err != nil {
		return nil, RegenerationFromShareCombineRegenerationBlindSharesError{Err: err}
	}

	priv := sk.priv //group.NewScalar(sk.private, sk.params.group.Order())

	eval, err := polynomial.LagrangeBasisPolynomial(g, newIndex, indices, sk.params.Index())
	if err != nil {
		return nil, RegenerationFromShareLagrangeBasisPolynomialError{Err: err}
	}

	priv, err = priv.Mul(eval)
	if err != nil {
		return nil, RegenerationFromShareMulError{Err: err}
	}

	priv, err = priv.Add(blind)
	if err != nil {
		return nil, RegenerationFromShareAddError{Err: err}
	}

	return &regenerationShare{
		shareIndex:        sk.params.Index(),
		regenerationIndex: newIndex,
		share:             priv,
	}, nil
}

// combineRegShares uses the received regeneration shares to generate a new private keyshare.
func combineRegShares(g group.Group, threshold uint64, shares []*regenerationShare) (*PrivateKeyShare, error) {
	if len(shares) < int(threshold) {
		return nil, CombineRegSharesSharesLessThanThresholdError{
			Expected: int(threshold),
			Got:      len(shares),
		}
	}

	regenIndex := shares[0].regenerationIndex
	var seenIndices []*group.Scalar
	for _, share := range shares {
		if share.regenerationIndex.Equal(regenIndex) != nil {
			return nil, CombineRegSharesInconsistentRegIndexError{}
		}

		for _, seen := range seenIndices {
			err := share.shareIndex.Equal(seen)
			if err == nil {
				return nil, CombineRegSharesMultipleSharesFromSingleIndex{}
			}
		}

		seenIndices = append(seenIndices, share.shareIndex)
	}

	sk := group.ZeroScalar(g.Order())

	var err error
	for _, share := range shares {
		sk, err = sk.Add(share.share)
		if err != nil {
			return nil, CombineRegSharesAddError{Err: err}
		}
	}

	privkey, err := NewPrivateKeyShareWithSecret(NewParameters(g, regenIndex), sk)
	if err != nil {
		return nil, CombineRegSharesNewPrivateKeyShareFromSecretError{Err: err}
	}

	return privkey, nil
}

func Verify(pk *PublicKeyShare, proof *DecryptionShareProof, extra []byte) error {
	err := pk.params.Index().Equal(proof.index)
	if err != nil {
		return VerifyImplPublicKeyShareEqualIndexError{Err: err}
	}

	extraBytes, err := asn1MarshalShareExtraBytes(pk.params.Index(), extra)
	if err != nil {
		return VerifyImplPublicKeyShareASN1MarshalShareExtraBytesError{Err: err}
	}

	keyCommitment, err := nizkp.Verify(pk.params.g.Generator(), pk.public, proof.proof.Proof().Challenge(), proof.proof.Proof().Response())
	if err != nil {
		return VerifyImplPublicKeyShareKeyCommittmentVerifyLogProofError{Err: err}
	}

	msgCommitment, err := nizkp.Verify(proof.proof.Ciphertext().A(), proof.proof.Decryption().Value(), proof.proof.Proof().Challenge(), proof.proof.Proof().Response())
	if err != nil {
		return VerifyImplPublicKeyShareMsgCommittmentVerifyLogProofError{Err: err}
	}

	pkey := elgamal.NewPublicKey(elgamal.NewParameters(pk.Parameters().Group()), pk.public)
	challenge, err := pkey.ProofChallenge(proof.proof.Ciphertext(), proof.proof.Decryption().Value(), msgCommitment, keyCommitment, extraBytes)
	if err != nil {
		return VerifyImplPublicKeyShareProofChallengeError{Err: err}
	}

	err = challenge.Equal(proof.proof.Proof().Challenge())
	if err != nil {
		return VerifyImplPublicKeyShareEqualChallengeError{Err: err}
	}

	return nil
}

func Decrypt(sk *PrivateKeyShare, ciphertext *elgamal.Ciphertext, _ bool) (*DecryptionShare, error) {
	E, err := ciphertext.A().Scale(sk.priv)
	if err != nil {
		return nil, DecryptImplPrivateKeyShareEphemeralScaleError{Err: err}
	}

	decrypted := elgamal.NewDecryption(sk.params.g, E)

	decryptedShare := NewDecryptionShare(sk.params.Index(), decrypted)

	return decryptedShare, nil
}
func ProvableDecrypt(y *group.Scalar, sk *PrivateKeyShare, ciphertext []byte, extra []byte, checkDecodable bool) (decryptedPart *DecryptionShare, proofForPart *DecryptionShareProof, err error) {
	decrypted, err := sk.Decrypt(ciphertext, checkDecodable) //Decrypt(sk, ct)
	if err != nil {
		return nil, nil, ProvableDecryptPrivateKeyShareDecryptImplError{Err: err}
	}

	extraBytes, err := asn1MarshalShareExtraBytes(sk.params.Index(), extra)
	if err != nil {
		return nil, nil, ProvableDecryptPrivateKeyShareASN1MarshalShareExtraBytesError{Err: err}
	}

	skk, err := elgamal.NewPrivateKey(elgamal.NewParameters(sk.Parameters().Group()), sk.priv)

	proof, err := skk.Prove(y, ciphertext, decrypted, extraBytes)
	if err != nil {
		return nil, nil, ProvableDecryptImplPrivateKeyShareProveYError{Err: err}
	}

	proof2, err := elgamal.ASN1UnmarshalDecryptionProof(sk.params.g, proof)
	if err != nil {
		return nil, nil, ProvableDecryptImplASN1UnmarshalProof{Err: err}
	}

	decProof := NewDecryptionProofShare(sk.params.Index(), proof2)

	return decrypted.(*DecryptionShare), decProof, nil
}

func UnmarshalPublicElementShare(oid asn_1.ObjectIdentifier, g group.Group, pk []byte) (group.Element, error) { //nolint:dupl
	//pkBytes := cryptobyte.String(pk)
	//
	//var outerSequence, pubBytes cryptobyte.String
	//
	//if !pkBytes.ReadASN1(&outerSequence, asn1.SEQUENCE) {
	//	return PublicKeyShare{}, ASN1UnmarshalPublicKeyShareNotASequenceError{}
	//}
	//
	//if !outerSequence.ReadAnyASN1Element(&indexBytes, nil) {
	//	return PublicKeyShare{}, ASN1UnmarshalPublicKeyShareNoIndexError{}
	//}

	//if !outerSequence.ReadAnyASN1Element(&pubBytes, nil) {
	//	return PublicKeyShare{}, ASN1UnmarshalPublicKeyShareNoPublicError{}
	//}
	//
	//if !pkBytes.Empty() || !outerSequence.Empty() {
	//	return PublicKeyShare{}, ASN1UnmarshalPublicKeyShareTrailingBytesError{}
	//}

	//Index, err := group.UnmarshalScalar(indexBytes, g.Order())
	//if err != nil {
	//	return PublicKeyShare{}, ASN1UnmarshalPublicKeyShareASN1UnmarshalIndexError{Err: err}
	//}

	pub, err := elgamal.ASN1UnmarshalPublicElement(g, pk)
	if err != nil {
		return nil, ASN1UnmarshalPublicKeyShareASN1UnmarshalPublicKeyError{Err: err}
	}
	//
	//pkey := elgamal.NewPublicKey(g, pub)

	//pubkey := NewPublicKeyShare(Group{group: g, Index: Index}, pkey)

	return pub, nil
}

func UnmarshalPrivateElementShare(g group.Group, sk []byte) (*group.Scalar, error) {
	//skBytes := cryptobyte.String(sk)
	//
	//var outerSequence, indexBytes, privBytes cryptobyte.String
	//
	//if !skBytes.ReadASN1(&outerSequence, asn1.SEQUENCE) {
	//	return nil, ASN1UnmarshalPrivateKeyShareNotASN1SequenceError{}
	//}
	//
	//if !outerSequence.ReadAnyASN1Element(&indexBytes, nil) {
	//	return nil, ASN1UnmarshalPrivateKeyShareNoIndexError{}
	//}
	//
	//if !outerSequence.ReadAnyASN1Element(&privBytes, nil) {
	//	return nil, ASN1UnmarshalPrivateKeyShareNoPrivError{}
	//}
	//
	//if !skBytes.Empty() || !outerSequence.Empty() {
	//	return nil, ASN1UnmarshalPrivateKeyShareTrailingBytesError{}
	//}
	//
	//Index, err := group.UnmarshalScalar(indexBytes, g.Order())
	//if err != nil {
	//	return nil, ASN1UnmarshalPrivateKeyShareASN1UnmarshalIndexError{Err: err}
	//}
	//
	//priv, err := group.UnmarshalScalar(privBytes, g.Order())
	//if err != nil {
	//	return nil, ASN1UnmarshalPrivateKeyShareASN1UnmarshalPrivError{Err: err}
	//}
	//
	//privkey, err := NewPrivateKeyShareWithSecret(g, Index, priv)
	//if err != nil {
	//	return nil, ASN1UnmarshalPrivateKeyShareNewPrivateKeyShareFromSecretError{Err: err}
	//}

	ska, err := group.UnmarshalScalar(sk, g.Order())
	if err != nil {
		panic(err)
	}

	//privkey, err := NewPrivateKeyShareWithSecret(params, ska)
	//if err != nil {
	//	return nil, ASN1UnmarshalPrivateKeyShareNewPrivateKeyShareFromSecretError{Err: err}
	//}

	return ska, nil
}

func ASN1UnmarshalDecryptionShare(g group.Group, data []byte) (*DecryptionShare, error) { //nolint:dupl
	dataBytes := cryptobyte.String(data)

	var outerSequence, shareBytes, pubBytes cryptobyte.String

	if !dataBytes.ReadASN1(&outerSequence, asn1.SEQUENCE) {
		return nil, ASN1UnmarshalDecryptionShareNotASequenceError{}
	}

	if !outerSequence.ReadAnyASN1Element(&shareBytes, nil) {
		return nil, ASN1UnmarshalDecryptionShareNoIndexError{}
	}

	if !outerSequence.ReadAnyASN1Element(&pubBytes, nil) {
		return nil, ASN1UnmarshalDecryptionShareNoPublicError{}
	}

	if !dataBytes.Empty() || !outerSequence.Empty() {
		return nil, ASN1UnmarshalDecryptionShareTrailingBytesError{}
	}

	index, err := group.UnmarshalScalar(asn_11.DER(shareBytes), g.Order())
	if err != nil {
		return nil, ASN1UnmarshalDecryptionShareIndexError{Err: err}
	}

	pub, err := elgamal.ASN1UnmarshalDecryption(g, pubBytes)
	if err != nil {
		return nil, ASN1UnmarshalDecryptionShareASN1UnmarshalDecryptionError{Err: err}
	}

	proof := NewDecryptionShare(index, pub)

	return proof, nil
}

func ASN1UnmarshalDecryptionProofShare(g group.Group, data []byte) (*DecryptionShareProof, error) { //nolint:dupl
	dataBytes := cryptobyte.String(data)

	var outerSequence, shareBytes, pubBytes cryptobyte.String

	if !dataBytes.ReadASN1(&outerSequence, asn1.SEQUENCE) {
		return nil, ASN1UnmarshalDecryptionProofShareNotASequenceError{}
	}

	if !outerSequence.ReadAnyASN1Element(&shareBytes, nil) {
		return nil, ASN1UnmarshalDecryptionProofShareNoIndexError{}
	}

	if !outerSequence.ReadAnyASN1Element(&pubBytes, nil) {
		return nil, ASN1UnmarshalDecryptionProofShareNoPublicError{}
	}

	if !dataBytes.Empty() || !outerSequence.Empty() {
		return nil, ASN1UnmarshalDecryptionProofShareTrailingBytesError{}
	}

	index, err := group.UnmarshalScalar(asn_11.DER(shareBytes), g.Order())
	if err != nil {
		return nil, ASN1UnmarshalDecryptionProofShareIndexError{Err: err}
	}

	pub, err := elgamal.ASN1UnmarshalDecryptionProof(g, pubBytes)
	if err != nil {
		return nil, ASN1UnmarshalDecryptionProofShareASN1UnmarshalDecryptionProofError{Err: err}
	}

	proof := NewDecryptionProofShare(index, pub)

	return proof, nil
}

// asn1MarshalShareExtraBytes as
//
//	extra ::= SEQUENCE {
//		INTEGER
//		OCTET STRING
//		OCTET STRING
//		...
//	}
func asn1MarshalShareExtraBytes(shareIndex *group.Scalar, extra ...[]byte) ([]byte, error) {
	shareIndexBytes, err := shareIndex.Marshal()
	if err != nil {
		return nil, ASN1MarshalShareExtraBytesShareIndexASN1Marshal{Err: err}
	}

	var c cryptobyte.Builder

	c.AddASN1(asn1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(shareIndexBytes)
		if len(extra) > 0 {
			c.AddASN1(asn1.SEQUENCE, func(c *cryptobyte.Builder) {
				for _, ex := range extra {
					c.AddASN1OctetString(ex)
				}
			})
		}
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalShareExtraBytesShareBytes{Err: err}
	}

	return b, nil
}
