package util

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"

	crypto2 "tivi.io/core/crypto"
	asn_1 "tivi.io/core/crypto/asn1"
	"tivi.io/core/math/group"
)

// GetPEMCertificate decodes and returns a single certificate from the given PEM encoded data.
func GetPEMCertificate(data []byte) (*x509.Certificate, error) {
	decodedCert, err := decodePEMCertificate(data)
	if err != nil {
		return nil, fmt.Errorf("decode PEM certificate: %w", err)
	}

	var cert *x509.Certificate
	cert, err = x509.ParseCertificate(decodedCert)
	if err != nil {
		return nil, fmt.Errorf("x509 parse certificate: %w", err)
	}

	return cert, nil
}

func decodePEMCertificate(encodedCert []byte) ([]byte, error) {
	p, rest := pem.Decode(encodedCert)
	if p == nil {
		return nil, fmt.Errorf("no PEM certificate")
	}

	if len(rest) > 0 {
		return nil, fmt.Errorf("PEM trailing error")
	}

	if p.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("wrong PEM block type")
	}

	if len(p.Headers) > 0 {
		return nil, fmt.Errorf("PEM headers error")
	}

	return p.Bytes, nil
}

func DefaultUint64(value, def uint64) uint64 {
	if value == 0 {
		return def
	}

	return value
}

func AlgorithmIdentifierCmp(aid1, aid2 pkix.AlgorithmIdentifier) bool {
	if !aid1.Algorithm.Equal(aid2.Algorithm) {
		return false
	}

	aid1Params := aid1.Parameters.FullBytes
	if len(aid1Params) == 0 {
		// If the optional Group field is empty encode it as a NULL tag with length 0
		aid1Params = []byte{5, 0}
	}

	aid2Params := aid2.Parameters.FullBytes
	if len(aid2Params) == 0 {
		// If the optional Group field is empty encode it as a NULL tag with length 0
		aid2Params = []byte{5, 0}
	}

	return bytes.Equal(aid1Params, aid2Params)
}

func IsRDNSequenceEqual(seq1, seq2 pkix.RDNSequence) bool {
	if len(seq1) != len(seq2) {
		return false
	}

	for i, a := range seq1 {
		b := seq2[i]
		if len(a) != len(b) {
			return false
		}

		c := make(pkix.RelativeDistinguishedNameSET, len(b))
		copy(c, b)
		// If panic was caused, it means a and b are not equal
		// because of non-comparable attribute values.
		defer func() {
			recover() //nolint:errcheck
		}()
		for len(c) > 0 {
			for _, aatv := range a {
				found := false
				// Assuming an unordered sequence
				for j, catv := range c {
					if aatv.Type.Equal(catv.Type) && aatv.Value == catv.Value {
						// Delete value matched and keep searching
						c[j] = c[len(c)-1]
						c = c[:len(c)-1]
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}

	return true
}

// EncryptAll uses encryption public key to encrypt plaintexts.
//
// Identifier is optional, but if not nil, then is used to identify a set of
// plaintexts, normally you leave it nil and identifier is set to be a
// public key algorithm, however, when you have a distributed encryption scheme
// then public and private key algorithm differ. That is because public key is
// of regular encryption scheme and private key, which is in form of private key
// shares, is of distributed encryption scheme.
func EncryptAll(rand *group.Scalar, pkey crypto2.Encrypter, identifier asn1.ObjectIdentifier, plaintexts ...[]byte) ([]byte, error) {
	var err error
	ciphertextsDer := make([][]byte, len(plaintexts))

	// Encrypt all plaintexts, we will then ASN.1 compact them together
	for i, plaintext := range plaintexts {
		// Encrypt using pkey implementation
		ciphertextsDer[i], err = pkey.Encrypt(rand, plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt a plaintext: %v", err)
		}
	}

	// ASN.1 compact ciphertexts
	ciphertextsDerCompacted, err := asn_1.Concat(ciphertextsDer...)
	if err != nil {
		return nil, fmt.Errorf("failed to compact ciphertexts into ASN.1: %v", err)
	}

	// If nil, then use public key algorithm identifier, otherwise use provided one
	if identifier == nil {
		identifier = pkey.Parameters().Algorithm()
	}

	identified, err := asn_1.AddIdentifier(identifier, ciphertextsDerCompacted)
	if err != nil {
		return nil, fmt.Errorf("failed to add OID to compacted ciphertexts: %v", err)
	}

	return identified, nil
}

// DecryptAll uses encryption private key to decrypt a ciphertext. Returns ASN.1
// unpacked ciphertexts (split ciphertext to many) and ASN.1 decrypted values.
//
// Identifier is optional, but if not nil, then is used to identify a set of plaintexts,
// normally you leave it nil and identifier is set to be a private key algorithm,
// however, when you have a distributed encryption scheme then public and
// private key algorithm differ. That is because public key is of regular encryption
// scheme and private key, which is in form of private key shares, is of distributed
// encryption scheme.
func DecryptAll(key crypto2.Decrypter, identifier asn1.ObjectIdentifier, ciphertext []byte, checkDecodable bool) ([][]byte, []crypto2.Decryption, error) {
	// ASN.1 split ciphertext to many
	ciphertexts, err := ParseCiphertext(key, identifier, ciphertext)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to split ASN.1 ciphertext to many: %v", err)
	}

	decryptions := make([]crypto2.Decryption, len(ciphertexts))

	// Loop over all ciphertexts and decrypt then individually
	for i, cipher := range ciphertexts {
		decryptions[i], err = key.Decrypt(cipher, checkDecodable)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decrypt a ciphertext: %v", err)
		}
	}

	return ciphertexts, decryptions, nil
}

// ProvableDecryptAll uses encryption private key to provably decrypt a ciphertext.
// Returns ASN.1 unpacked ciphertexts (split ciphertext to many),
// ASN.1 decrypted values and ASN.1 proofs.
//
// Identifier is optional, but if not nil, then is used to identify a set of plaintexts,
// normally you leave it nil and identifier is set to be a private key algorithm,
// however, when you have a distributed encryption scheme then public and
// private key algorithm differ. That is because public key is of regular encryption
// scheme and private key, which is in form of private key shares, is of distributed
// encryption scheme.
func ProvableDecryptAll(rand *group.Scalar, key crypto2.DecryptionKey, identifier asn1.ObjectIdentifier, ciphertext, salt []byte, checkDecodable bool) ([][]byte, []crypto2.Decryption, [][]byte, error) {
	// ASN.1 split ciphertext to many
	ciphertexts, err := ParseCiphertext(key, identifier, ciphertext)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to split ASN.1 ciphertext to many: %v", err)
	}

	decryptions := make([]crypto2.Decryption, len(ciphertexts))
	proofs := make([][]byte, len(ciphertexts))

	// Loop over all ciphertexts and provably decrypt then individually
	for i, cipher := range ciphertexts {
		decryptions[i], err = key.Decrypt(cipher, checkDecodable)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decrypt a ciphertext: %v", err)
		}
		proofs[i], err = key.Prove(rand, cipher, decryptions[i], salt)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to prove decrypt a ciphertext: %v", err)
		}
	}

	return ciphertexts, decryptions, proofs, nil
}

// SignAll uses signing private key to sign a set of data. The set of data is
// ASN.1 compacted, then signed.
// Signature is ASN.1 marshalled as well. Format of a signature depends on
// key implementation.
func SignAll(rand *group.Scalar, key crypto2.Signer, data ...[]byte) ([]byte, error) {
	// ASN.1 pack data
	packed, err := asn_1.Pack(data...)
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 compact data: %v", err)
	}

	signature, err := key.Sign(rand, packed)
	if err != nil {
		return nil, fmt.Errorf("failed to sign ASN.1 compacted data: %v", err)
	}

	// ASN.1 signature as {OID, signature}
	identified, err := asn_1.AddIdentifier(key.Parameters().Algorithm(), signature)
	if err != nil {
		return nil, fmt.Errorf("failed to add ASN.1 identifier to signed data: %v", err)
	}

	return identified, nil
}

// VerifyAll uses signing public key to verify a signature.
// Data is the data that was used during SignAll.
func VerifyAll(rand *group.Scalar, pkey crypto2.SignatureVerifier, signature []byte, data ...[]byte) error {
	// Remove OID identifier from a signature
	algorithm, signatureNoIdentifier, err := asn_1.ParseIdentifier(signature)
	if err != nil {
		return fmt.Errorf("failed to parse ASN.1 identifier from a signature: %v", err)
	}

	// Algorithm of a public key doesn't match parsed algorithm identifier
	// from a signature
	if !pkey.Parameters().Algorithm().Equal(algorithm) {
		return fmt.Errorf("public key algorithm %v mismatches signature's one %v", pkey.Parameters().Algorithm(), algorithm)
	}

	// ASN.1 pack data (just the same way as in SignAll)
	packed, err := asn_1.Pack(data...)
	if err != nil {
		return fmt.Errorf("failed to ASN.1 compact data: %v", err)
	}

	err = pkey.Verify(rand, signatureNoIdentifier, packed)
	if err != nil {
		return fmt.Errorf("failed to verify a signature: %v", err)
	}

	return nil
}

// ParseCiphertext returns ASN.1 split of a ciphertext after ASN.1 unpacking.
func ParseCiphertext(key crypto2.KeyInfo, identifier asn1.ObjectIdentifier, ciphertext []byte) ([][]byte, error) {
	// Get identifier from an ASN.1 compacted ciphertext and remove identifier
	// from ciphertext
	algorithm, ciphertextNoIdentifier, err := asn_1.ParseIdentifier(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ASN.1 identifier from a ciphertext: %v", err)
	}

	// If nil then use private key algorithm, otherwise provided one
	if identifier == nil {
		identifier = key.Parameters().Algorithm()
	}

	// Identifiers of private key and ciphertext should match
	if !identifier.Equal(algorithm) {
		return nil, fmt.Errorf("key algorithm %v mismatches signature's one %v", key.Parameters().Algorithm(), algorithm)
	}

	// ASN.1 split ciphertext to many
	ciphertexts, err := asn_1.Split(ciphertextNoIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 split ciphertext: %v", err)
	}

	return ciphertexts, nil
}
