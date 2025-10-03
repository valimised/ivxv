// Package marshal contains marshalling and unmarshalling interfaces that client
// should implement in order to be able to marshal/unmarshal custom x509/PKCS8
// keys from/into crypto key object.
//
// These interfaces allow caller to have custom formatted x509/PKCS8 key's:
//
// a) parameters
//
// b) key material (element)
//
// Crypto key object is a central unit that allows caller to fully utilize crypto
// library interfaces.
package marshal

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/math/group"
)

// KeyMarshaller is an interface that allows a caller to marshal crypto key into
// its own, custom formatted, x509/PKCS8 key.
type KeyMarshaller interface {
	// MarshalKeyParameters marshals crypto key parameters into PKIX parameters.
	MarshalKeyParameters(crypto.AlgorithmIdentifierParameters) (pkix.AlgorithmIdentifier, error)

	// MarshalKeyElement marshals crypto key element into a SubjectPublicKey (x509)
	// or PrivateKey (PKCS8).
	MarshalKeyElement(crypto.KeyInfo) (asn_1.DER, error)
}

// KeyUnmarshaller is an interface that allows a caller to unmarshal its own,
// custom formatted, X509/PKCS8 key into a crypto key object.
//
// T is either group.Element for x509 or *group.Scalar for PKCS8, any other types
// will result in error, as are not supported by crypto library.
type KeyUnmarshaller[T any] interface {
	// UnmarshalKeyParameters unmarshalls PKIX parameters into crypto key parameters.
	UnmarshalKeyParameters(pkix.AlgorithmIdentifier) (asn1.ObjectIdentifier, crypto.AlgorithmIdentifierParameters, error)

	// UnmarshalKeyElement unmarshalls SubjectPublicKey (x509) or PrivateKey (PKCS8)
	// into a crypto key element.
	UnmarshalKeyElement(group.Group, asn_1.DER) (T, error)
}
