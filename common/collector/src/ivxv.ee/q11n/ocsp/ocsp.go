/*
Package ocsp contains qualification protocols which perform OCSP requests.

ocsp registers the following qualifiers:

    - ocsp, which checks the status of the signing certificate using OCSP
            (there must be exactly one signature on the signed container) and
            returns the OCSP response as the qualifying property, and

    - ocsptm, which does the same as ocsp, but uses the DigestInfo of the
              signature as the OCSP nonce. This allows us to use the OCSP
              response as a timemark, proving that the signature existed when
              the certificate status was checked. Containers passed to the
              ocsptm qualifier must implement SignatureValuer.
*/
package ocsp // import "ivxv.ee/q11n/ocsp"

import (
	"context"
	"crypto"

	"ivxv.ee/container"
	"ivxv.ee/cryptoutil"
	"ivxv.ee/ocsp"
	"ivxv.ee/q11n"
	"ivxv.ee/yaml"

	// SHA1 is used for the ocsptm DigestInfo.
	_ "crypto/sha1"
)

func init() {
	q11n.Register(q11n.OCSP, newreg(false))
	q11n.Register(q11n.OCSPTM, newreg(true))
}

type client struct {
	ocsp *ocsp.Client
	tm   bool
}

func newreg(tm bool) func(yaml.Node, string) (q11n.Qualifier, error) {
	return func(n yaml.Node, _ string) (q q11n.Qualifier, err error) {
		var conf ocsp.Conf
		if err = yaml.Apply(n, &conf); err != nil {
			return nil, YAMLApplyError{Err: err}
		}

		c := &client{tm: tm}
		c.ocsp, err = ocsp.New(&conf)
		if err != nil {
			return nil, ClientError{Err: err}
		}

		return c, nil
	}
}

func (c *client) Qualify(ctx context.Context, container container.Container) ([]byte, error) {
	sigs := container.Signatures()
	if len(sigs) != 1 {
		return nil, NoSingleSignatureError{Count: len(sigs)}
	}
	cert := sigs[0].Signer
	issuer := sigs[0].Issuer

	var nonce []byte
	if c.tm {
		valuer, ok := container.(SignatureValuer)
		if !ok {
			return nil, ContainerNotSignatureValuerError{}
		}
		value, err := valuer.SignatureValue(sigs[0].ID)
		if err != nil {
			return nil, SignatureValueError{ID: sigs[0].ID, Err: err}
		}
		nonce = cryptoutil.DigestInfo(crypto.SHA1, value)
	}

	status, err := c.ocsp.Check(ctx, cert, issuer, nonce)
	if err != nil {
		return nil, CheckOCSPError{Err: err}
	}
	if !status.Good {
		if status.Unknown {
			return nil, q11n.BadCertificateStatusError{
				Err: StatusUnknownError{},
			}
		}
		return nil, q11n.BadCertificateStatusError{
			Err: RevokedError{Reason: status.RevocationReason},
		}
	}
	return status.RawResponse, nil
}

// SignatureValuer returns the raw signature value of the signature with the
// given ID.
type SignatureValuer interface {
	SignatureValue(id string) ([]byte, error)
}
