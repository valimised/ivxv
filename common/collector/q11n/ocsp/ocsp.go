//nolint:fmt
/*
Package ocsp contains qualification protocols which perform OCSP requests.

ocsp registers the following qualifiers:

    - ocsp, which checks the status of the signing certificate using OCSP
            (there must be exactly one signature on the signed container) and
            returns the OCSP response as the qualifying property, and
*/
package ocsp

import (
	"context"

	"ivxv.ee/common/collector/container"
	"ivxv.ee/common/collector/ocsp"
	"ivxv.ee/common/collector/q11n"
	"ivxv.ee/common/collector/yaml"
)

func init() {
	q11n.Register(q11n.OCSP, newreg(), ocsp.ExtractProducedAtTimeFromRawOcspResponse)
}

type client struct {
	ocsp *ocsp.Client
}

func newreg() func(yaml.Node, string) (q11n.Qualifier, error) {
	return func(n yaml.Node, _ string) (q q11n.Qualifier, err error) {
		var conf ocsp.Conf
		if err = yaml.Apply(n, &conf); err != nil {
			return nil, YAMLApplyError{Err: err,
				Description: _OCSP_CFG}
		}

		c := new(client)
		c.ocsp, err = ocsp.New(&conf)
		if err != nil {
			return nil, ClientError{Err: err,
				Description: _OCPS_NEW}
		}

		return c, nil
	}
}

func (c *client) Qualify(ctx context.Context, container container.Container) ([]byte, error) {
	sigs := container.Signatures()
	if len(sigs) != 1 {
		return nil, NoSingleSignatureError{Count: len(sigs),
			Description: _OCSP_NO_SIG}
	}
	cert := sigs[0].Signer
	issuer := sigs[0].Issuer

	var nonce []byte

	status, err := c.ocsp.Check(ctx, cert, issuer, nonce)
	if err != nil {
		return nil, CheckOCSPError{Err: err,
			Description: _OCSP_VERIFY}
	}
	if !status.Good {
		if status.Unknown {
			return nil, q11n.BadCertificateStatusError{
				Err: StatusUnknownError{Description: _OCSP_UNKNOWN},
			}
		}
		return nil, q11n.BadCertificateStatusError{
			Err: RevokedError{Reason: status.RevocationReason, Description: _OCSP_REVOKED},
		}
	}
	return status.RawResponse, nil
}

// SignatureValuer returns the raw signature value of the signature with the
// given ID.
type SignatureValuer interface {
	SignatureValue(id string) ([]byte, error)
}
