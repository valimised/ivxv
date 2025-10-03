// Package x509 implements x509 and PKCS8 parsing.
package x509

import (
	"encoding/asn1"
	"fmt"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/x509/internal"
	"tivi.io/core/crypto/x509/marshal"
	"tivi.io/core/math/group"
)

// PEM is a format of base-64 serialized x509/PKCS8 keys with included headers.
type PEM []byte

// Unmarshaller is a x509 key decoding crypto instance. Any x509 decoding
// from x509 to crypto public key are done via this instance.
type Unmarshaller[T any] struct {
	// unmarshaller is the implementation of marshal.KeyUnmarshaller interface,
	// i.e. custom x509 parameters' and public key material parser.
	unmarshaller marshal.KeyUnmarshaller[group.Element]
}

// NewUnmarshaller creates new x509 key decoder to crypto public key instance.
func NewUnmarshaller[T any](unmarshaller marshal.KeyUnmarshaller[group.Element]) *Unmarshaller[T] {
	return &Unmarshaller[T]{unmarshaller: unmarshaller}
}

// Unmarshal unmarshalls PEM/DER to crypto public key.
func (x509 *Unmarshaller[T]) Unmarshal(data []byte) (params crypto.AlgorithmIdentifierParameters, pkey T, err error) {
	// Unmarshal PEM or DER to DER
	der, err := internal.Pem2Der(data, internal.X509Header)
	if err != nil {
		return nil, pkey, fmt.Errorf("cannot PEM/DER decode x509 key to DER: %v", err)
	}

	// Unmarshal DER to SubjectPublicKeyInfo
	spki, err := internal.Der2Pki[internal.SubjectPublicKeyInfo](der)
	if err != nil {
		return nil, pkey, fmt.Errorf("cannot decode DER x509 to SubjectPublicKeyInfo format: %v", err)
	}

	// Unmarshal x509 AlgorithmIdentifier parameters
	oid, params, err := x509.unmarshaller.UnmarshalKeyParameters(spki.Algorithm)
	if err != nil {
		return nil, pkey, fmt.Errorf("cannot unmarshal Algorithm parameters: %v", err)
	}

	// Unmarshal x509 SubjectPublicKey, i.e. x509 key material
	pub, err := x509.unmarshaller.UnmarshalKeyElement(params.Group(), spki.SubjectPublicKey.Bytes)
	if err != nil {
		return nil, pkey, fmt.Errorf("cannot unmarshal SubjectPublicKey material: %v", err)
	}

	// Create new crypto public key
	publicKeyUnmarshaller, err := internal.NewPublicKey(oid)
	if err != nil {
		return nil, pkey, fmt.Errorf("failed to fetch crypto public key unmarshaller: %v", err)
	}

	newPublicKey, err := publicKeyUnmarshaller(params, pub)
	if err != nil {
		return nil, pkey, fmt.Errorf("failed to create new crypto public key: %v", err)
	}

	// Check that created crypto public key can be cast to caller provided T type
	var ok bool
	pkey, ok = newPublicKey.(T)
	if !ok {
		return nil, pkey, fmt.Errorf("failed to cast crypto public key type to caller's provided T type")
	}

	return
}

// Marshaller is a x509 key encoding crypto instance. Any x509 encoding
// to x509 from crypto private key are done via this instance.
type Marshaller struct {
	// marshaller is the implementation of marshal.KeyMarshaller interface,
	// i.e. custom x509 parameters' and public key material encoder.
	marshaller marshal.KeyMarshaller
}

// NewMarshaller creates new x509 key encoder from crypto public key instance
// to x509.
func NewMarshaller(marshaller marshal.KeyMarshaller) Marshaller {
	return Marshaller{marshaller: marshaller}
}

// Marshal marshals crypto public key to PEM.
func (x509 Marshaller) Marshal(keyInfo crypto.KeyInfo) (pem PEM, err error) {
	der, err := x509.MarshalDER(keyInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal x509 to DER: %v", err)
	}

	pem, err = internal.Der2Pem(der, internal.X509Header)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal x509 DER key to PEM: %v", err)
	}

	return
}

// MarshalDER marshals crypto public key to DER.
func (x509 Marshaller) MarshalDER(key crypto.KeyInfo) (der asn_1.DER, err error) {
	// Marshal crypto public key parameters
	algorithmIdentifier, err := x509.marshaller.MarshalKeyParameters(key.Parameters())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal crypto public key parameters to PKCS8 parameters: %v", err)
	}

	// Marshal crypto public key element
	pub, err := x509.marshaller.MarshalKeyElement(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal crypto public key element: %v", err)
	}

	// Crypto public key to SubjectPublicKeyInfo
	spki := internal.SubjectPublicKeyInfo{
		Algorithm: algorithmIdentifier,
		SubjectPublicKey: asn1.BitString{
			Bytes:     pub,
			BitLength: len(pub) * 8,
		},
	}

	der, err = internal.Pki2Der[internal.SubjectPublicKeyInfo](spki)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal x509 to DER: %v", err)
	}

	return
}
