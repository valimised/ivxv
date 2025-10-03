package elgamal

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"

	"golang.org/x/crypto/cryptobyte"
	asn_1 "golang.org/x/crypto/cryptobyte/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/crypto/elgamal"
	asn_11 "tivi.io/core/encoding/asn1"
	"tivi.io/core/math/group"
)

type ECpGroupParameters struct {
	CurveID    string
	ElectionID string
}

type ModPGroupParameters struct {
	P          *big.Int
	G          *big.Int
	ElectionID string
}

type X509Unmarshaller struct {
	ElectionID string
}

func NewX509Unmarshaller() *X509Unmarshaller {
	return &X509Unmarshaller{}
}

// UnmarshalKeyParameters unmarshalls IVXV x509 public key parameters, but returns
// crypto library compatible ElGamal oid.
func (x509 *X509Unmarshaller) UnmarshalKeyParameters(algorithmIdentifier pkix.AlgorithmIdentifier) (oid asn1.ObjectIdentifier, parameters crypto.AlgorithmIdentifierParameters, err error) { //nolint:lll
	var G group.Group

	switch algorithmIdentifier.Algorithm.String() {
	case crypto.ElGamalEncryptionOID().String():
		var params ModPGroupParameters

		_, err = asn1.Unmarshal(algorithmIdentifier.Parameters.FullBytes, &params)
		if err != nil {
			return nil, nil, UnmarshalKeyParametersX509UnmarshallerModPKeyParametersError{Err: err,
				Description: _PUB_MODP_PARAMS}
		}

		// Crypto library compatible oid for ElGamal encryption
		oid = crypto.ElGamalEncryptionOID()
		x509.ElectionID = params.ElectionID
		G, err = params.IVXVModPGroupParser()
	case ecElGamalEncryptionOID().String():
		var params ECpGroupParameters

		_, err = asn1.Unmarshal(algorithmIdentifier.Parameters.FullBytes, &params)
		if err != nil {
			return nil, nil, UnmarshalKeyParametersX509UnmarshallerECpKeyParametersError{Err: err,
				Description: _PUB_ECP_PARAMS}
		}

		// Crypto library compatible oid for ElGamal encryption
		oid = crypto.ElGamalEncryptionOID()
		x509.ElectionID = params.ElectionID
		G, err = params.IVXVECpGroupParser()
	default:
		return nil, nil, UnmarshalKeyParametersX509UnmarshallerUnknownAlgorithmError{
			OID:         algorithmIdentifier.Algorithm.String(),
			Description: _PUB_ALGO}
	}
	if err != nil {
		return nil, nil, UnmarshalKeyParametersX509UnmarshallerError{
			Err: err, Description: _PUB_PARAMS}
	}

	return oid, elgamal.NewParameters(G), nil
}

// UnmarshalKeyElement unmarshalls IVXV x509 public key element.
func (x509 *X509Unmarshaller) UnmarshalKeyElement(g group.Group, der asn_11.DER) (element group.Element, err error) {
	sequence := cryptobyte.String(der)
	var data cryptobyte.String

	if !sequence.ReadASN1(&data, asn_1.SEQUENCE) {
		return nil, UnmarshalKeyElementX509UnmarshallerNotSequenceError{
			Description: _PUB_DATA}
	}

	if !sequence.Empty() {
		return nil, UnmarshalKeyElementX509UnmarshallerTrailingBytesError{
			Description: _PUB_TRAIL}
	}

	element, err = elgamal.ASN1UnmarshalPublicElement(g, asn_11.DER(data))
	if err != nil {
		return nil, UnmarshalKeyElementX509UnmarshallerElGamalPublicError{
			Err: err, Description: _PUB_EL}
	}

	return
}

func (p *ModPGroupParameters) IVXVModPGroupParser() (G group.Group, err error) { //nolint:gocritic
	switch p.P.BitLen() {
	case 2048:
		G, err = group.Get("RFC3526ModPGroup2048")
	case 3072:
		G, err = group.Get("RFC3526ModPGroup3072")
	case 4096:
		G, err = group.Get("RFC3526ModPGroup4096")
	case 8192:
		G, err = group.Get("RFC3526ModPGroup8192")
	default:
		return nil, IVXVModPGroupParserModPGroupParametersUnsupportedModPElementError{Err: err,
			Description: _MODP_LEN}
	}
	if err != nil {
		return nil, IVXVModPGroupParserModPGroupParametersUnknownModPGroupError{Err: err,
			Description: _MODP_G}
	}

	return
}

func (p *ECpGroupParameters) IVXVECpGroupParser() (G group.Group, err error) { //nolint:gocritic
	switch p.CurveID {
	case "P-256":
		G, err = group.Get("NIST-P256")
	case "P-384":
		return group.Get("NIST-P384")
	case "P-521":
		return group.Get("NIST-P521")
	default:
		return nil, IVXVECpGroupParserECpGroupParametersUnsupportedECpElementError{Err: err,
			Description: _ECP_LEN}
	}
	if err != nil {
		return nil, IVXVECpGroupParserECpGroupParametersUnknownGroupError{Err: err,
			Description: _ECP_G}
	}

	return
}
