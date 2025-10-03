package x509

import (
	"crypto/x509/pkix"
	"encoding/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	"tivi.io/core/crypto/elgamal/distributed"
	x_509 "tivi.io/core/crypto/x509/marshal"
	asn_1 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group"
)

type UnmarshallerPKCS8 struct{}

func NewUnmarshallerPKCS8() x_509.KeyUnmarshaller[*group.Scalar] {
	return UnmarshallerPKCS8{}
}

func (pkcs8 UnmarshallerPKCS8) UnmarshalKeyParameters(algoID pkix.AlgorithmIdentifier) (oid asn1.ObjectIdentifier, params crypto.AlgorithmIdentifierParameters, err error) {
	params, err = distributed.OfPKIXAlgorithmIdentifier(algoID)
	oid = algoID.Algorithm
	return
}

func (pkcs8 UnmarshallerPKCS8) UnmarshalKeyElement(g group.Group, der asn_1.DER) (*group.Scalar, error) {
	return elgamal.ASN1UnmarshalPrivateElement(g, der)
}
