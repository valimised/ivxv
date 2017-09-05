/*
Package tls implements voter authentication using TLS client authentication.

The server package requests any client certificate during the TLS handshake,
but does not require or verify it. This package can be used to require it and
set the parameters used for verification.

Note that only the configured roots and intermediates will be used for
verification: any intermediate certificates provided by the client will be
ignored.
*/
package tls // import "ivxv.ee/auth/tls"

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"

	"ivxv.ee/auth"
	"ivxv.ee/cryptoutil"
	"ivxv.ee/log"
	"ivxv.ee/ocsp"
	"ivxv.ee/server"
	"ivxv.ee/yaml"
)

func init() {
	auth.Register(auth.TLS, func(n yaml.Node) (auth.Verifier, error) {
		c := new(Conf)
		if err := yaml.Apply(n, c); err != nil {
			return nil, ConfigurationError{Err: err}
		}
		v, err := New(c)
		if err != nil {
			return nil, CertificateParsingError{Err: err}
		}
		return v, nil
	})
}

// Conf is the TLS authentication verifier configuration.
type Conf struct {
	Roots         []string   // PEM-encoded client certificate verification roots.
	Intermediates []string   // PEM-encoded client certificate verification intermediates.
	OCSP          *ocsp.Conf // Optional OCSP-checking configuration.
}

// V is an auth.Verifier which checks TLS authentication.
type V struct {
	rpool *x509.CertPool
	ipool *x509.CertPool
	ocsp  *ocsp.Client
}

// New returns a new TLS authentication verifier with the provided configuration.
func New(c *Conf) (v *V, err error) {
	if len(c.Roots) == 0 {
		return nil, UnconfiguredRootsError{}
	}

	v = new(V)
	if v.rpool, err = cryptoutil.PEMCertificatePool(c.Roots...); err != nil {
		return nil, RootsParsingError{Err: err}
	}
	if v.ipool, err = cryptoutil.PEMCertificatePool(c.Intermediates...); err != nil {
		return nil, IntermediatesParsingError{Err: err}
	}
	if c.OCSP != nil {
		if v.ocsp, err = ocsp.New(c.OCSP); err != nil {
			return nil, OCSPClientError{Err: err}
		}
	}
	return
}

// Verify implements the auth.Verifier interface. The token is unused and only
// the client certificate in context is verified.
func (v *V) Verify(ctx context.Context, token []byte) (*pkix.Name, error) {
	if len(token) > 0 {
		return nil, auth.MalformedTokenError{
			Err: NonEmptyTokenError{Token: log.Sensitive(token)},
		}
	}

	// Client certificates were added to the context by ivxv.ee/server.
	certs := server.TLSClient(ctx)
	if len(certs) == 0 {
		return nil, auth.CertificateError{Err: NoClientCertificateError{}}
	}
	cert := certs[0] // The client certificate must be first.

	opts := x509.VerifyOptions{
		Roots:         v.rpool,
		Intermediates: v.ipool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	chains, err := cert.Verify(opts)
	if err != nil {
		return nil, auth.CertificateError{
			Err: CertificateVerificationError{
				Certificate: cert.Raw, // Log entire cert for diagnostics.
				Err:         err,
			},
		}
	}

	if v.ocsp != nil {
		issuer := cert
		if len(chains[0]) > 1 { // At least one chain is guaranteed.
			issuer = chains[0][1]
		}
		status, err := v.ocsp.Check(ctx, cert, issuer, nil)
		if err != nil {
			return nil, CheckCertificateStatusError{
				Certificate: cert,
				Err:         err,
			}
		}
		if !status.Good {
			return nil, auth.CertificateError{
				Err: CertificateStatusError{
					Certificate:      cert,
					Unknown:          status.Unknown,
					RevocationReason: status.RevocationReason,
				},
			}
		}
	}

	return &cert.Subject, nil
}
