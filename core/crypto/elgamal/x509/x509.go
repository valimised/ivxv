// Package x509 provides adapters and other ElGamal specific implementations for crypto/x509.
package x509

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"tivi.io/core/crypto/elgamal"
	asn_1 "tivi.io/core/encoding/asn1"

	"tivi.io/core/crypto"
	x509e "tivi.io/core/crypto/elgamal/internal/x509"
	x_509 "tivi.io/core/crypto/x509/marshal"
	"tivi.io/core/math/group"
)

type Unmarshaller struct{}

func NewUnmarshaller() x_509.KeyUnmarshaller[group.Element] {
	return Unmarshaller{}
}

func (x509 Unmarshaller) UnmarshalKeyParameters(algoID pkix.AlgorithmIdentifier) (oid asn1.ObjectIdentifier, params crypto.AlgorithmIdentifierParameters, err error) {
	params, err = elgamal.OfPKIXAlgorithmIdentifier(algoID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ASN.1 unmarshal PKIX parameters to ElGamal public key parameters: %v", err)
	}
	oid = algoID.Algorithm
	return
}

func (x509 Unmarshaller) UnmarshalKeyElement(g group.Group, der asn_1.DER) (group.Element, error) {
	return elgamal.ASN1UnmarshalPublicElement(g, der)
}

type Marshaller struct{}

func NewMarshaller() x_509.KeyMarshaller {
	return Marshaller{}
}

func (x509 Marshaller) MarshalKeyParameters(params crypto.AlgorithmIdentifierParameters) (pkix.AlgorithmIdentifier, error) {
	return x509e.MarshalKeyParameters(params)
}

func (x509 Marshaller) MarshalKeyElement(info crypto.KeyInfo) (asn_1.DER, error) {
	return x509e.MarshalKeyElement(info)
}
