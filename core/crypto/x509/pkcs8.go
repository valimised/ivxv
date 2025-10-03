package x509

import (
	"fmt"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/x509/internal"
	"tivi.io/core/crypto/x509/marshal"
	"tivi.io/core/math/group"
)

// UnmarshallerPKCS8 is an PKCS8 key decoding crypto instance. Any PKCS8 decoding
// from PKCS8 to crypto private key are done via this instance.
type UnmarshallerPKCS8[T any] struct {
	// unmarshaller is the implementation of marshal.KeyUnmarshaller interface,
	// i.e. custom PKCS8 parameters' and private key material parser.
	unmarshaller marshal.KeyUnmarshaller[*group.Scalar]
}

// NewUnmarshallerPKCS8 creates new PKCS8 key decoder to crypto private key instance.
func NewUnmarshallerPKCS8[T any](unmarshaller marshal.KeyUnmarshaller[*group.Scalar]) *UnmarshallerPKCS8[T] {
	return &UnmarshallerPKCS8[T]{unmarshaller: unmarshaller}
}

// Unmarshal unmarshalls PEM/DER to crypto private key.
func (pkcs8 *UnmarshallerPKCS8[T]) Unmarshal(data []byte) (params crypto.AlgorithmIdentifierParameters, key T, err error) {
	// Unmarshal PEM or DER to DER
	der, err := internal.Pem2Der(data, internal.PKCS8Header)
	if err != nil {
		return nil, key, fmt.Errorf("cannot PEM/DER decode PKCS8 key to DER: %v", err)
	}

	// Unmarshal DER to PrivateKeyInfo
	pki, err := internal.Der2Pki[internal.PrivateKeyInfo](der)
	if err != nil {
		return nil, key, fmt.Errorf("cannot decode DER to PKCS8 PrivateKeyInfo: %v", err)
	}

	// Unmarshal PKCS8 AlgorithmIdentifier parameters
	oid, params, err := pkcs8.unmarshaller.UnmarshalKeyParameters(pki.PrivateKeyAlgorithm)
	if err != nil {
		return nil, key, fmt.Errorf("cannot unmarshal PrivateKeyAlgorithm parameters: %v", err)
	}

	// Unmarshal PKCS8 PrivateKey, i.e. PKCS8 key material
	privateKey, err := pkcs8.unmarshaller.UnmarshalKeyElement(params.Group(), pki.PrivateKey)
	if err != nil {
		return nil, key, fmt.Errorf("cannot unmarshal PrivateKey material: %v", err)
	}

	// Fetch crypto private key unmarshaller by OID returned by PKCS8 parameters parser
	privateKeyUnmarshaller, err := internal.NewPrivateKey(oid)
	if err != nil {
		return nil, key, fmt.Errorf("failed to fetch crypto private key unmarshaller: %v", err)
	}

	newPrivateKey, err := privateKeyUnmarshaller(params, privateKey)
	if err != nil {
		return nil, key, fmt.Errorf("failed to create new crypto private key: %v", err)
	}

	// Check that created crypto private key can be cast to caller provided T type
	var ok bool
	key, ok = newPrivateKey.(T)
	if !ok {
		return nil, key, fmt.Errorf("failed to cast crypto private key type to caller's provided T type")
	}

	return
}

// MarshallerPKCS8 is an PKCS8 key encoding crypto instance. Any PKCS8 encoding
// to PKCS8 from crypto private key are done via this instance.
type MarshallerPKCS8 struct {
	// marshaller is the implementation of marshal.KeyMarshaller interface,
	// i.e. custom PKCS8 parameters' and private key material encoder.
	marshaller marshal.KeyMarshaller
}

// NewMarshallerPKCS8 creates new PKCS8 key encoder from crypto private key instance
// to PKCS8.
func NewMarshallerPKCS8(marshaller marshal.KeyMarshaller) MarshallerPKCS8 {
	return MarshallerPKCS8{marshaller: marshaller}
}

// Marshal marshals crypto private key to PEM.
func (pkcs8 MarshallerPKCS8) Marshal(key crypto.KeyInfo) (pem PEM, err error) {
	der, err := pkcs8.MarshalDER(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PKCS8 to DER: %v", err)
	}

	pem, err = internal.Der2Pem(der, internal.PKCS8Header)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PKCS8 DER key to PEM: %v", err)
	}

	return
}

// MarshalDER marshals crypto private key to DER.
func (pkcs8 MarshallerPKCS8) MarshalDER(key crypto.KeyInfo) (der asn_1.DER, err error) {
	// Marshal crypto private key parameters
	algorithmIdentifier, err := pkcs8.marshaller.MarshalKeyParameters(key.Parameters())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal crypto private key parameters to PKCS8 parameters: %v", err)
	}

	// Marshal crypto private key element
	priv, err := pkcs8.marshaller.MarshalKeyElement(key) //keyInfo.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal crypto private key material: %v", err)
	}

	// Crypto private key to PrivateKeyInfo
	pki := internal.PrivateKeyInfo{
		Version:             internal.Pkcs8Version,
		PrivateKeyAlgorithm: algorithmIdentifier,
		PrivateKey:          priv,
	}

	der, err = internal.Pki2Der[internal.PrivateKeyInfo](pki)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PKCS8 to DER: %v", err)
	}

	return
}
