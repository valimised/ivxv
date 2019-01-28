/*
Package cryptoutil provides utility functions that are used as building blocks
when implementing cryptographic formats and algorithms.
*/
package cryptoutil // import "ivxv.ee/cryptoutil"

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
)

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
		return nil, NotPEMEncodingError{}
	}
	if len(rest) > 0 {
		return nil, PEMTrailingDataError{Trailing: rest}
	}
	if block.Type != blockType {
		return nil, PEMBlockTypeError{Type: block.Type, Expected: blockType}
	}
	if len(block.Headers) > 0 {
		return nil, PEMHeadersError{Headers: block.Headers}
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
				Idx: i,
				Err: err,
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
		return nil, PEMCertificateDecodeError{Err: err}
	}
	cert, err = x509.ParseCertificate(der)
	if err != nil {
		return nil, PEMCertParseX509Error{Err: err}
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
			return nil, PEMCertError{Index: i, Err: err}
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

// ParseECDSAXMLSignature parses an ECDSA XML digital signature. The XML
// signature is a concatenation of the R and S outputs of the ECDSA algorithm
// as specified in https://tools.ietf.org/html/rfc4050#section-3.3.
//
// Note! This format differs from the one used in X.509 certificates specified
// here https://tools.ietf.org/html/rfc3279#section-2.2.3.
func ParseECDSAXMLSignature(signature []byte) (r, s *big.Int, err error) {
	if len(signature)%2 != 0 {
		return nil, nil, ECDSASignatureLengthError{}
	}
	n := len(signature) / 2
	r = big.NewInt(0).SetBytes(signature[:n])
	s = big.NewInt(0).SetBytes(signature[n:])
	return
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
