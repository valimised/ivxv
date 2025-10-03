package distributed

import (
	"crypto/x509/pkix"
	asn_1 "encoding/asn1"
	asn_11 "tivi.io/core/encoding/asn1"

	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"

	"tivi.io/core/crypto"
	"tivi.io/core/math/group"
)

// Parameters are ElGamal key parameters.
type Parameters struct {
	// g is a group that ElGamal key belongs to.
	g     group.Group
	index *group.Scalar
}

func NewParameters(g group.Group, index *group.Scalar) *Parameters {
	return &Parameters{
		g:     g,
		index: index,
	}
}

func (p *Parameters) Group() group.Group {
	return p.g
}

func (p *Parameters) Index() *group.Scalar {
	return p.index
}

func (p *Parameters) Algorithm() asn_1.ObjectIdentifier {
	return crypto.DistributedElGamalEncryptionOID
}

func (p *Parameters) ToPKIXAlgorithmIdentifier() (algorithm pkix.AlgorithmIdentifier, err error) {
	parameters, err := group.Marshal(p.g)
	if err != nil {
		return pkix.AlgorithmIdentifier{}, ToPKIXAlgorithmIdentifierGroupError{Err: err}
	}

	index, err := p.index.Marshal()
	if err != nil {
		panic(err)
	}

	var builder cryptobyte.Builder

	builder.AddASN1(asn1.SEQUENCE, func(c *cryptobyte.Builder) {
		c.AddBytes(parameters)
		c.AddBytes(index)
	})

	bb, err := builder.Bytes()
	if err != nil {
		panic(err)
	}

	algorithm = pkix.AlgorithmIdentifier{
		Algorithm: p.Algorithm(),
		Parameters: asn_1.RawValue{
			FullBytes: bb,
		},
	}

	return
}

func OfPKIXAlgorithmIdentifier(algorithm pkix.AlgorithmIdentifier) (crypto.AlgorithmIdentifierParameters, error) {
	c := cryptobyte.String(algorithm.Parameters.FullBytes)

	var cc cryptobyte.String
	var cc1 cryptobyte.String
	var cc2 cryptobyte.String
	if !c.ReadASN1(&cc, asn1.SEQUENCE) {
		panic("SEQ")
	}
	if !cc.ReadAnyASN1Element(&cc1, nil) {
		panic("1111")
	}
	if !cc.ReadAnyASN1Element(&cc2, nil) {
		panic("2222")
	}

	g, err := group.Unmarshal(asn_11.DER(cc1))
	if err != nil {
		return nil, OfPKIXAlgorithmIdentifierGroupError{Err: err}
	}

	index, err := group.UnmarshalScalar(asn_11.DER(cc2), g.Order())
	if err != nil {
		return nil, err
	}

	return NewParameters(g, index), nil
}
