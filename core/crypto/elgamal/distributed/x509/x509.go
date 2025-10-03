package x509

import (
	"crypto/x509/pkix"
	"encoding/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	"tivi.io/core/crypto/elgamal/distributed"
	asn_1 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group"
)

type Unmarshaller struct{}

func NewUnmarshaller() Unmarshaller {
	return Unmarshaller{}
}

func (x509 Unmarshaller) UnmarshalKeyParameters(algoID pkix.AlgorithmIdentifier) (oid asn1.ObjectIdentifier, params crypto.AlgorithmIdentifierParameters, err error) {
	params, err = distributed.OfPKIXAlgorithmIdentifier(algoID)
	oid = algoID.Algorithm
	return
}

func (x509 Unmarshaller) UnmarshalKeyElement(g group.Group, der asn_1.DER) (group.Element, error) {
	return elgamal.ASN1UnmarshalPublicElement(g, der)
}
