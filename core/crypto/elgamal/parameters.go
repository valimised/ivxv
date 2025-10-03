package elgamal

import (
	"crypto/x509/pkix"
	asn_1 "encoding/asn1"
	"fmt"

	"tivi.io/core/crypto"
	"tivi.io/core/math/group"
)

// Parameters are ElGamal key parameters.
type Parameters struct {
	// g is a group that ElGamal key belongs to.
	g group.Group
}

// NewParameters returns new ElGamal key parameters.
func NewParameters(g group.Group) *Parameters {
	return &Parameters{g: g}
}

// Group returns ElGamal group that this key belongs to.
func (p *Parameters) Group() group.Group {
	return p.g
}

// Algorithm returns ElGamal OID as defined in
// https://datatracker.ietf.org/doc/html/draft-rfced-info-pgutmann-00#section-2
func (p *Parameters) Algorithm() asn_1.ObjectIdentifier {
	return crypto.ElGamalEncryptionOID()
}

// ToPKIXAlgorithmIdentifier converts ElGamal parameters into PKIX parameters.
func (p *Parameters) ToPKIXAlgorithmIdentifier() (algorithm pkix.AlgorithmIdentifier, err error) {
	parameters, err := group.Marshal(p.g)
	if err != nil {
		return pkix.AlgorithmIdentifier{}, fmt.Errorf("failed to convert ElGamal parameters to PKIX parameters: %v", err)
	}

	algorithm = pkix.AlgorithmIdentifier{
		Algorithm: p.Algorithm(),
		Parameters: asn_1.RawValue{
			FullBytes: parameters,
		},
	}
	return
}

// OfPKIXAlgorithmIdentifier converts PKIX parameters into ElGamal parameters.
func OfPKIXAlgorithmIdentifier(algorithm pkix.AlgorithmIdentifier) (crypto.AlgorithmIdentifierParameters, error) {
	parameters, err := group.Unmarshal(algorithm.Parameters.FullBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to ASN.1 unmarshal group from PKIX parameters: %v", err)
	}

	return NewParameters(parameters), nil
}
