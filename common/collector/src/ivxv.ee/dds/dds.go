/*
Package dds implements Mobile-ID authentication and signing using
DigiDocService.

https://sk-eid.github.io/dds-documentation/
*/
package dds

import (
	"crypto/x509"

	"ivxv.ee/cryptoutil"
	"ivxv.ee/ocsp"
)

var (
	// InputError wraps errors which are caused by bad input to dds functions.
	_ = InputError{Err: nil}

	// NotMIDUserError is the error returned if the submitted phone number
	// does not belong to a Mobile-ID user.
	_ = NotMIDUserError{}

	// AbsentError is the error returned if the voter's phone is absent
	// (switched off or out of coverage).
	_ = AbsentError{}

	// CanceledError is the error returned if the voter canceled the
	// operation.
	_ = CanceledError{}

	// ExpiredError is the error returned if the session expired
	// before the voter entered their PIN.
	_ = ExpiredError{}

	// CertificateError wraps errors which are caused by errors with the
	// voter's certificate (revoked, suspended, not activated, etc).
	_ = CertificateError{Err: nil}

	// StatusError wraps errors which are caused by an unexpected session
	// status: this is a catch-all for other types of Mobile-ID problems.
	_ = StatusError{Err: nil}
)

// Conf contains the configurable options for the DigiDocService client. It
// only contains serialized values such that it can easily be unmarshaled from
// a file.
type Conf struct {
	URL         string // URL of DigiDocService.
	Language    string // Language for user dialog in mobile phone.
	ServiceName string // Service name agreed with DigiDocService.
	AuthMessage string // Message to display during authentication.
	SignMessage string // Message to display during signing.

	Roots         []string  // PEM-encoded authentication certificate verification roots.
	Intermediates []string  // PEM-encoded authentication certificate verification intermediates.
	OCSP          ocsp.Conf // OCSP configuration for checking authentication certificate revocation.
}

// Client implements DigiDocService authentication and signing.
type Client struct {
	conf  Conf
	rpool *x509.CertPool
	ipool *x509.CertPool
	ocsp  *ocsp.Client
}

// New returns a new DigiDocService client with the provided configuration.
func New(conf *Conf) (c *Client, err error) {
	if len(conf.Roots) == 0 {
		return nil, UnconfiguredRootsError{}
	}

	c = &Client{conf: *conf} // Save a copy of conf so it cannot be changed.
	if c.rpool, err = cryptoutil.PEMCertificatePool(c.conf.Roots...); err != nil {
		return nil, RootsParsingError{Err: err}
	}
	if c.ipool, err = cryptoutil.PEMCertificatePool(c.conf.Intermediates...); err != nil {
		return nil, IntermediatesParsingError{Err: err}
	}
	if c.ocsp, err = ocsp.New(&c.conf.OCSP); err != nil {
		return nil, OCSPClientError{Err: err}
	}
	return
}
