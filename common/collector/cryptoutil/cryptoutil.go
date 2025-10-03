/*
Package cryptoutil provides utility functions that are used as building blocks
when implementing cryptographic formats and algorithms.
*/
package cryptoutil

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"math/big"
)

const entropy256bit uint64 = 32

// hp is a map from hash function identifier to DigestInfo ASN.1 prefix.
var hp = map[crypto.Hash][]byte{
	crypto.SHA224: {
		0x30, 0x2d, 0x30, 0x0d,
		0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x04,
		0x04, 0x1c,
	},
	crypto.SHA256: {
		0x30, 0x31, 0x30, 0x0d,
		0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01,
		0x04, 0x20,
	},
	crypto.SHA384: {
		0x30, 0x41, 0x30, 0x0d,
		0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x02,
		0x04, 0x30,
	},
	crypto.SHA512: {
		0x30, 0x51, 0x30, 0x0d,
		0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x03,
		0x04, 0x40,
	},
}

// DigestInfo hashes data using the hash function h and returns the result in a
// DER-encoded PKCS #1 DigestInfo structure. DigestInfo panics if h is not
// available or a prefix for it has not been defined in x509util: this reduces
// error checking code, as these problems can be statically detected.
func DigestInfo(h crypto.Hash, data []byte) []byte {
	// Create a copy of the prefix, so that Sum can append to it directly.
	// Allocate space for both the prefix and the hash.
	p := hp[h]
	prefix := make([]byte, len(p), len(p)+h.Size())
	copy(prefix, p)

	hash := h.New()
	hash.Write(data)
	return hash.Sum(prefix)
}

// PEMDecode performs strict PEM decoding, i.e., the block type must match
// exactly, no headers are allowed, and there cannot be any trailing data.
func PEMDecode(encoded, blockType string) (decoded []byte, err error) {
	block, rest := pem.Decode([]byte(encoded))
	if block == nil {
		return nil, NotPEMEncodingError{
			Description: _CUTIL_PEM,
		}
	}
	if len(rest) > 0 {
		return nil, PEMTrailingDataError{Trailing: rest,
			Description: _CUTIL_PEM}
	}
	if block.Type != blockType {
		return nil, PEMBlockTypeError{Type: block.Type,
			Expected: blockType, Description: _CUTIL_PEM_H}
	}
	if len(block.Headers) > 0 {
		return nil, PEMHeadersError{Headers: block.Headers,
			Description: _CUTIL_PEM_H_BAD}
	}
	return block.Bytes, nil
}

// PEMCertificates parses one or more PEM type certificates into a slice and
// returns them.
func PEMCertificates(pems ...string) (certs []*x509.Certificate, err error) {
	for i, pem := range pems {
		cert, err := PEMCertificate(pem)
		if err != nil {
			return nil, PEMCertificateError{
				Idx:         i,
				Err:         err,
				Description: _CUTIL_CERT_PEM,
			}
		}
		certs = append(certs, cert)
	}
	return
}

// PEMCertificate parses a PEM block of type CERTIFICATE and the X.509
// certificate within.
func PEMCertificate(p string) (cert *x509.Certificate, err error) {
	der, err := PEMDecode(p, "CERTIFICATE")
	if err != nil {
		return nil, PEMCertificateDecodeError{Err: err,
			Description: _CUTIL_PEM}
	}
	cert, err = x509.ParseCertificate(der)
	if err != nil {
		return nil, PEMCertParseX509Error{Err: err,
			Description: _CUTIL_CERT_PEM}
	}
	return
}

// Base64Certificate parses a X.509 certificate from base64-encoded DER data.
func Base64Certificate(b string) (cert *x509.Certificate, err error) {
	certDER, err := base64.StdEncoding.DecodeString(b)
	if err != nil {
		return nil, Base64CertificateDecodeError{Err: err,
			Description: _CUTIL_CERT_B64}
	}
	cert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return nil, Base64CertParseX509Error{Err: err,
			Description: _CUTIL_CERT_PARSE}
	}
	return
}

// PEMCertificatePool parses a slice of PEM-encoded certificates and puts them
// in a CertPool.
func PEMCertificatePool(pems ...string) (pool *x509.CertPool, err error) {
	pool = x509.NewCertPool()
	for i, pem := range pems {
		c, err := PEMCertificate(pem)
		if err != nil {
			return nil, PEMCertError{Index: i, Err: err,
				Description: _CUTIL_CERT_PEM}
		}
		pool.AddCert(c)
	}
	return pool, nil
}

// CertificatePool puts all provided certificates in a CertPool. This is a
// simple helper function to replace a for loop with a single expression.
func CertificatePool(certs ...*x509.Certificate) (pool *x509.CertPool) {
	pool = x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}
	return pool
}

// ReEncodeECDSASignature re-encodes an ECDSA XML digital signature to an ECDSA
// X.509 PKI digital signature. The XML signature format is specified in
// https://tools.ietf.org/html/rfc4050#section-3.3 and the X.509 PKI signature
// format is specified in https://tools.ietf.org/html/rfc3279#section-2.2.3.
func ReEncodeECDSASignature(signature []byte) (recode []byte, err error) {
	var r, s *big.Int
	if r, s, err = ParseECDSAXMLSignature(signature); err != nil {
		return nil, ECDSASignatureParseError{Signature: signature,
			Err: err, Description: _CUTIL_ECDSA2XML}
	}
	if recode, err = asn1.Marshal(struct {
		R *big.Int
		S *big.Int
	}{r, s}); err != nil {
		return nil, ECDSASignatureASN1MarshalError{Err: err,
			Description: _CUTIL_ASN1ECDSA}
	}
	return
}

func IsECDSAASN1EncodedSignature(signature []byte) error {
	type ecdsaRawSig struct {
		R *big.Int
		S *big.Int
	}
	sig := new(ecdsaRawSig)

	unmarshal, err := asn1.Unmarshal(signature, sig)
	if err != nil {
		return SignatureIsNotASN1EncodedECDSAError{Err: err,
			Description: _CUTIL_UASN1ECDSA}
	}
	if len(unmarshal) != 0 {
		return SignatureIsASN1EncodedECDSAButHasRestAmountOfBytesError{
			Amount:      len(unmarshal),
			Description: _CUTIL_UASN1ECDSA,
		}
	}
	return nil
}

// ParseECDSAXMLSignature parses an ECDSA XML digital signature. The XML
// signature is a concatenation of the R and S outputs of the ECDSA algorithm
// as specified in https://tools.ietf.org/html/rfc4050#section-3.3.
//
// Note! This format differs from the one used in X.509 certificates specified
// here https://tools.ietf.org/html/rfc3279#section-2.2.3.
func ParseECDSAXMLSignature(signature []byte) (r, s *big.Int, err error) {
	if len(signature)%2 != 0 {
		return nil, nil, ECDSASignatureLengthError{Description: _CUTIL_ECDSA_SIZE}
	}
	n := len(signature) / 2
	r = big.NewInt(0).SetBytes(signature[:n])
	s = big.NewInt(0).SetBytes(signature[n:])
	return
}

// ParseECDSAASN1Signature parses an ECDSA ASN.1 digital signature. The ASN.1
// signature is a DER-encoded SEQUENCE of the R and S INTEGERS of the ECDSA
// algorithm as specified in https://tools.ietf.org/html/rfc3279#section-2.2.3.
//
// Note! This format differs from the one used in XML digital signatures
// specified here https://tools.ietf.org/html/rfc4050#section-3.3.
func ParseECDSAASN1Signature(der []byte) (r, s *big.Int, err error) {
	var parsed struct {
		R *big.Int
		S *big.Int
	}
	if rest, err := asn1.Unmarshal(der, &parsed); err != nil {
		return nil, nil, ECDSASignatureASN1UnmarshalError{Err: err,
			Description: _CUTIL_UASN1ECDSA}
	} else if len(rest) > 0 {
		return nil, nil, ECDSASignatureTrailingDataError{Err: err,
			Description: _CUTIL_UASN1ECDSA}
	}
	return parsed.R, parsed.S, nil
}

// AlgorithmIdentifierCmp compares two pkix.AlgorithmIdentifier structures
// and returns true if they are equal, false otherwise.
func AlgorithmIdentifierCmp(a, b pkix.AlgorithmIdentifier) bool {
	if !a.Algorithm.Equal(b.Algorithm) {
		return false
	}

	aParams := a.Parameters.FullBytes
	if len(aParams) == 0 {
		// If the optional Parameters field is empty encode it as a NULL tag with length 0
		aParams = []byte{5, 0}
	}
	bParams := b.Parameters.FullBytes
	if len(bParams) == 0 {
		// If the optional Parameters field is empty encode it as a NULL tag with length 0
		bParams = []byte{5, 0}
	}

	return bytes.Equal(aParams, bParams)
}

// Nonce44Bytes returns a base64(cryptographic nonce) with a length of 44 bytes.
func Nonce44Bytes() (string, error) {
	nonce := make([]byte, entropy256bit)
	_, err := rand.Read(nonce)
	if err != nil {
		return "", NonceGenerationError{Err: err,
			Description: _CUTIL_44B_RAND}
	}
	return base64.StdEncoding.EncodeToString(nonce), nil
}
