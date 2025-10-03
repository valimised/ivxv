// Package x509 is for internal usage of ElGamal implementations and contains common logic
// for any x509/PKCS8 operations.
package x509

import (
	"crypto/x509/pkix"

	"tivi.io/core/crypto"
	asn_1 "tivi.io/core/encoding/asn1"
)

// MarshalKeyParameters provides common logic for ElGamal key parameters PKIX marshalling.
func MarshalKeyParameters(params crypto.AlgorithmIdentifierParameters) (pkix.AlgorithmIdentifier, error) {
	return params.ToPKIXAlgorithmIdentifier()
}

// MarshalKeyElement provides common logic for ElGamal key marshalling.
func MarshalKeyElement(info crypto.KeyInfo) (asn_1.DER, error) {
	return info.Marshal()
}
