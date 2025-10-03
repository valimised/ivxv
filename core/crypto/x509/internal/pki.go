package internal

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
)

const (
	Pkcs8Version = 0
)

// https://tools.ietf.org/html/rfc5208#section-5
type PrivateKeyInfo struct {
	Version             int
	PrivateKeyAlgorithm pkix.AlgorithmIdentifier
	PrivateKey          []byte
	// Attributes are not used for this PrivateKeyInfo
}

// https://tools.ietf.org/html/rfc5280#section-4.1.2.7
type SubjectPublicKeyInfo struct {
	Algorithm        pkix.AlgorithmIdentifier
	SubjectPublicKey asn1.BitString
}

// Pki is a generic public Key Infrastructure interface which wraps x509
// (SubjectPublicKeyInfo) and PKCS8 (PrivateKeyInfo) standards.
type Pki interface {
	SubjectPublicKeyInfo | PrivateKeyInfo
}

// Der2Pki unmarshalls DER into either SubjectPublicKeyInfo or
// PrivateKeyInfo depending on T type.
func Der2Pki[T Pki](der []byte) (pki T, err error) {
	_, err = asn1.Unmarshal(der, &pki)
	if err != nil {
		return pki, fmt.Errorf("failed to ASN.1 unmarshal PKI: %v", err)
	}

	return
}

// Pki2Der marshals either SubjectPublicKeyInfo or PrivateKeyInfo, depending on T type,
// into DER.
func Pki2Der[T Pki](pki T) (der []byte, err error) {
	der, err = asn1.Marshal(pki)
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 marshal PKI: %v", err)
	}

	return
}
