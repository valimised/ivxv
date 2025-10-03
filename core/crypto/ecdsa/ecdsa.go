package ecdsa

import (
	"bytes"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/binary"

	"golang.org/x/crypto/cryptobyte"
	asn_1 "golang.org/x/crypto/cryptobyte/asn1"
	"golang.org/x/crypto/sha3"

	crypto2 "tivi.io/core/crypto"
	"tivi.io/core/math/group"
)

// https://datatracker.ietf.org/doc/rfc8692/
var ecdsaWithSHAKE256 = asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 6, 33}

func init() {
	crypto2.RegisterPrivateKeyASN1Unmarshaller(ecdsaWithSHAKE256, func(g group.Group, key []byte) (crypto2.SigningPrivateKey, error) {
		return ASN1UnmarshalPrivateKey(g, key)
	})
	//crypto2.RegisterPublicKeyASN1Unmarshaller(ecdsaWithSHAKE256, func(g group.Group, key []byte) (crypto2.SigningPublicKey, error) {
	//	return ASN1UnmarshalPublicKey(g, key)
	//})
}

// New generates ECDSA private key.
func New(g group.Group) (PrivateKey, error) {
	sk, err := group.RandomScalar(g.Order())
	if err != nil {
		return PrivateKey{}, GenerateRandomScalarError{Err: err}
	}
	gen := g.Generator()
	pub, err := gen.Scale(sk.val)
	if err != nil {
		return PrivateKey{}, GenerateScaleGeneratorByValError{Err: err}
	}
	pr := PrivateKey{
		p: sk,
		pub: PublicKey{
			prm: g,
			pub: pub,
		},
	}
	return pr, nil
}

type PublicKey struct {
	prm group.Group
	pub group.Element
}

func (pk PublicKey) MarshalASN1() ([]byte, error) {
	pubBytes, err := pk.pub.Marshal()
	if err != nil {
		return nil, ASN1MarshalPublicKeyASN1MarshalPublicError{Err: err}
	}

	var c cryptobyte.Builder
	c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(pubBytes)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalPublicKeyBytesError{Err: err}
	}

	return b, nil
}

func ASN1UnmarshalPublicKey(g group.Group, data []byte) (PublicKey, error) {
	s := cryptobyte.String(data)

	var inner cryptobyte.String
	var pubBytes cryptobyte.String

	if !s.ReadASN1(&inner, asn_1.SEQUENCE) {
		return PublicKey{}, ASN1UnmarshalPublicKeyNotASequenceError{}
	}

	if !inner.ReadAnyASN1Element(&pubBytes, nil) {
		return PublicKey{}, ASN1UnmarshalPublicKeyNoPublicError{}
	}

	if !s.Empty() || !inner.Empty() {
		return PublicKey{}, ASN1UnmarshalPublicKeyTrailingBytesError{}
	}

	pub, err := g.ElementOf(pubBytes)
	if err != nil {
		return PublicKey{}, ASN1UnmarshalElementOfASN1Error{Err: err}
	}

	return PublicKey{g, pub}, nil
}

func (pk PublicKey) Parameters() crypto2.AlgorithmIdentifierParameters {
	return nil //pk.prm
}

func (pk PublicKey) Algorithm() asn1.ObjectIdentifier {
	return ecdsaWithSHAKE256
}

// Fingerprint returns first 24 bytes of SHA256(pk), or 0 if hashing fails.
func (pk PublicKey) Fingerprint() uint64 {
	pubBytes, err := pk.pub.Marshal()
	if err != nil {
		return 0
	}

	sum := sha256.Sum256(pubBytes)
	return binary.BigEndian.Uint64(sum[24:])
}

type PrivateKey struct {
	pub PublicKey
	p   group.Scalar
}

func (sk PrivateKey) Parameters() crypto2.AlgorithmIdentifierParameters {
	return nil //sk.pub.
}

func (sk PrivateKey) Algorithm() asn1.ObjectIdentifier {
	return ecdsaWithSHAKE256
}

// Fingerprint is unimplemented for sk.
func (sk PrivateKey) Fingerprint() uint64 {
	return 0
}

func (sk PrivateKey) MarshalASN1() ([]byte, error) {
	skBytes, err := sk.p.Marshal()
	if err != nil {
		return nil, ASN1MarshalPrivateKeyASN1MarshalPrivateError{Err: err}
	}

	var c cryptobyte.Builder
	c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(skBytes)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalPrivateKeyBytesError{Err: err}
	}

	return b, nil
}

func ASN1UnmarshalPrivateKey(g group.Group, data []byte) (PrivateKey, error) {
	s := cryptobyte.String(data)

	var inner cryptobyte.String
	var privBytes cryptobyte.String

	if !s.ReadASN1(&inner, asn_1.SEQUENCE) {
		return PrivateKey{}, ASN1UnmarshalPrivateKeyNotASequenceError{}
	}

	if !inner.ReadAnyASN1Element(&privBytes, nil) {
		return PrivateKey{}, ASN1UnmarshalPrivateKeyNoPrivateError{}
	}

	if !s.Empty() || !inner.Empty() {
		return PrivateKey{}, ASN1UnmarshalPrivateKeyTrailingBytesError{}
	}

	sk, err := group.UnmarshalScalar(privBytes, g.Order())
	if err != nil {
		return PrivateKey{}, ASN1UnmarshalASN1UnmarshalScalarError{Err: err}
	}

	pub, err := g.Generator().Scale(sk.val)
	if err != nil {
		return PrivateKey{}, ASN1UnmarshalPrivateKeyRestorePublicError{Err: err}
	}
	pr := PrivateKey{
		p: sk,
		pub: PublicKey{
			prm: g,
			pub: pub,
		},
	}
	return pr, nil
}

type M struct {
	R, S group.Scalar
}

func (m M) MarshalASN1() ([]byte, error) {
	rBytes, err := m.R.Marshal()
	if err != nil {
		return nil, ASN1MarshalMRError{Err: err}
	}

	sBytes, err := m.S.Marshal()
	if err != nil {
		return nil, ASN1MarshalMSError{Err: err}
	}
	var c cryptobyte.Builder
	c.AddASN1(asn_1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(rBytes)
		c.AddBytes(sBytes)
	})

	b, err := c.Bytes()
	if err != nil {
		return nil, ASN1MarshalMBytesError{Err: err}
	}

	return b, nil
}

// ASN1UnmarshalCiphertext unmarshalls ElGamal Cipher.
func ASN1UnmarshalM(g group.Group, data []byte) (M, error) {
	c := cryptobyte.String(data)

	var inner cryptobyte.String
	var r, s cryptobyte.String

	if !c.ReadASN1(&inner, asn_1.SEQUENCE) {
		return M{}, ASN1UnmarshalMNotASequenceError{}
	}

	if !inner.ReadAnyASN1Element(&r, nil) {
		return M{}, ASN1UnmarshalMNoRError{}
	}

	if !inner.ReadAnyASN1Element(&s, nil) {
		return M{}, ASN1UnmarshalMNoSError{}
	}

	if !inner.Empty() || !c.Empty() {
		return M{}, ASN1UnmarshalMTrailingBytesError{}
	}

	var err error

	R, err := group.UnmarshalScalar(r, g.Order())
	if err != nil {
		return M{}, ASN1UnmarshalMASN1UnmarshalRError{Err: err}
	}

	S, err := group.UnmarshalScalar(s, g.Order())
	if err != nil {
		return M{}, ASN1UnmarshalMASN1UnmarshalSError{Err: err}
	}

	return M{R, S}, nil
}

func (sk PrivateKey) PublicKey() crypto2.SigningPublicKey {
	return sk.pub
}

func (sk PrivateKey) Sign(data []byte) ([]byte, error) {
	// privkey
	sc := group.NewScalar(sk.p.val, sk.pub.prm.Order())

	// rand scalar k
	k, err := group.RandomScalar(sk.pub.prm.Order())
	if err != nil {
		return nil, SignPrivateKeyRandomScalarError{Err: err}
	}
	// x = Gk
	X, err := sk.pub.prm.Generator().Scale(k.val)
	if err != nil {
		return nil, SignPrivateKeyScaleGeneratorByValError{Err: err}
	}

	// r = Xx
	r1 := X.(ecp.NistElement).X
	if err != nil {
		return nil, SignPrivateKeyCastToNistElementError{Err: err}
	}
	r := group.NewScalar(r1, sk.pub.prm.Order())

	// SHAKE256(h)
	shake256 := sha3.NewShake256()
	shake256.Write(data) //nolint: errcheck

	// c = SHAKE56(H(text))
	c, err := group.ScalarValueOfReader(shake256, sk.pub.prm.Order())
	if err != nil {
		return nil, SignPrivateKeyScalarValueOfReaderError{Err: err}
	}

	// kN = k^-1
	kN := k.Inverse()

	// s = kN * (c + sc * r)
	sc2, err := sc.Mul(r)
	if err != nil {
		return nil, SignPrivateKeyMulError{Err: err}
	}
	sc3, err := c.Add(sc2)
	if err != nil {
		return nil, SignPrivateKeyAddError{Err: err}
	}
	kN, err = kN.Mul(sc3)
	if err != nil {
		return nil, SignPrivateKeyMul2Error{Err: err}
	}

	mm := M{r, kN}
	MBytes, err := mm.MarshalASN1()
	if err != nil {
		return nil, SignPrivateKeyASN1MarshalError{Err: err}
	}

	return MBytes, nil
}

func (pk PublicKey) Verify(signature, signed []byte) error {
	data, err := ASN1UnmarshalM(pk.prm, signature)
	if err != nil {
		return VerifyPublicKeyASN1UnmarshalMError{Err: err}
	}

	// SHAKE256(h)
	shake256 := sha3.NewShake256()
	shake256.Write(signed) //nolint: errcheck

	// c = SHAKE56(H(text))
	c, err := group.ScalarValueOfReader(shake256, pk.prm.Order())
	if err != nil {
		return VerifyPublicKeyScalarValueOfReaderError{Err: err}
	}

	// w = s^-1
	news := group.NewScalar(data.S.val, pk.prm.Order())
	w := news.Inverse()

	a, err := c.Mul(w)
	if err != nil {
		return VerifyPublicKeyNewScalarError{Err: err}
	}
	r := group.NewScalar(data.R.val, pk.prm.Order())

	b, err := r.Mul(w)
	if err != nil {
		return VerifyPublicKeyMulError{Err: err}
	}
	X1, err := pk.prm.Generator().Scale(a.val)
	if err != nil {
		return VerifyPublicKeyScaleGeneratorByValError{Err: err}
	}
	X2, err := pk.pub.Scale(b.val)
	if err != nil {
		return VerifyPublicKeyScalePublicByValError{Err: err}
	}

	X, err := X1.Op(X2)
	if err != nil {
		return VerifyPublicKeyOpError{Err: err}
	}
	xc1 := X.(ecp.NistElement).X
	if err != nil {
		return VerifyPublicKeyCastToNistElementError{Err: err}
	}
	xc := group.NewScalar(xc1, pk.prm.Order())
	sc1, err := xc.Marshal()
	if err != nil {
		return VerifyPublicKeyNewScalar2Error{Err: err}
	}

	sc2, err := data.R.Marshal()
	if err != nil {
		return VerifyPublicKeyASN1MarshalError{Err: err}
	}

	if !bytes.Equal(sc1, sc2) {
		return VerifyPublicKeyEqualError{}
	}

	return nil
}
