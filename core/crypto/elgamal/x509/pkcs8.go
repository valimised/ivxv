package x509

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	internal "tivi.io/core/crypto/elgamal/internal/x509"
	x_509 "tivi.io/core/crypto/x509/marshal"
	asn_1 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group"
)

type UnmarshallerPKCS8 struct{}

func NewUnmarshallerPKCS8() x_509.KeyUnmarshaller[*group.Scalar] {
	return UnmarshallerPKCS8{}
}

func (pkcs8 UnmarshallerPKCS8) UnmarshalKeyParameters(algoID pkix.AlgorithmIdentifier) (oid asn1.ObjectIdentifier, params crypto.AlgorithmIdentifierParameters, err error) {
	params, err = elgamal.OfPKIXAlgorithmIdentifier(algoID)
	oid = algoID.Algorithm
	return
}

func (pkcs8 UnmarshallerPKCS8) UnmarshalKeyElement(params group.Group, der asn_1.DER) (*group.Scalar, error) {
	return elgamal.ASN1UnmarshalPrivateElement(params, der)
}

type MarshallerPKCS8 struct{}

func NewMarshallerPKCS8() x_509.KeyMarshaller {
	return MarshallerPKCS8{}
}

func (pkcs8 MarshallerPKCS8) MarshalKeyParameters(params crypto.AlgorithmIdentifierParameters) (pkix.AlgorithmIdentifier, error) {
	return internal.MarshalKeyParameters(params)
}

func (pkcs8 MarshallerPKCS8) MarshalKeyElement(info crypto.KeyInfo) (asn_1.DER, error) {
	return internal.MarshalKeyElement(info)
}
