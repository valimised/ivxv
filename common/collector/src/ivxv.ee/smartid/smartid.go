/*
Package smartid provides client for the Smart-ID REST service.

https://github.com/SK-EID/smart-id-documentation
*/
package smartid

import (
	"crypto"
	"crypto/x509"
	"strings"

	"ivxv.ee/cryptoutil"
	"ivxv.ee/ocsp"
)

var (
	// InputError wraps errors which are caused by bad input to smartid functions.
	_ = InputError{Err: nil}

	// VerificationError is the error returned if the voter 3 different code
	// was displayed in app and selected wrong code.
	_ = VerificationError{}

	// AccountError is the error returned that are caused by user account configuration.
	_ = AccountError{}

	// CanceledError is the error returned if the voter canceled the
	// operation.
	_ = CanceledError{}

	// ExpiredError is the error returned if the session expired
	// before the voter did any action.
	_ = ExpiredError{}

	// CertificateError wraps errors which are caused by errors with the
	// voter's certificate (revoked, suspended, not activated, etc).
	_ = CertificateError{Err: nil}

	// StatusError wraps errors which are caused by an unexpected session
	// status: this is a catch-all for other types of Smart-ID problems.
	_ = StatusError{Err: nil}

	// allowedAuthHashFunctions is list of hash functions allowed for
	// authentication. Since the hash function is determined by it's hash
	// size (Conf.AuthChallengeSize), the functions should have hash sizes
	// unique in the list. If Conf.AuthChallengeSize is zero, then the
	// first one from this list is used.
	allowedAuthHashFunctions = []crypto.Hash{crypto.SHA256, crypto.SHA384, crypto.SHA512}
)

// Conf contains the configurable options for the Smart-ID REST API client. It
// only contains serialized values such that it can easily be unmarshaled from
// a file.
type Conf struct {
	URL              string // URL of Smart-ID REST API.
	RelyingPartyUUID string // The UUID of the relying party, i.e service consumer.
	RelyingPartyName string // The name of the relying party, i.e service consumer.

	CertificateLevel      string                     // Certificate level for requests
	AuthInteractionsOrder []allowedInteractionsOrder // Interactions to display during authentication.
	SignInteractionsOrder []allowedInteractionsOrder // Interactions to display during signing.

	AuthChallengeSize int64 // The authentication challenge size.
	StatusTimeoutMS   int64 // The long-polling timeout for authentication/signing status request.

	Roots         []string  // PEM-encoded authentication certificate verification roots.
	Intermediates []string  // PEM-encoded authentication certificate verification intermediates.
	OCSP          ocsp.Conf // OCSP configuration for checking authentication certificate revocation.
}

// Client implements Smart-ID REST API authentication and signing.
type Client struct {
	conf  Conf
	rpool *x509.CertPool
	ipool *x509.CertPool
	ocsp  *ocsp.Client
	// authHashFunction is the hash function to use for authentication.
	// Determined by Conf.AuthChallengeSize.
	authHashFunction crypto.Hash
	// url is the same as 'conf.URL', but guaranteed to end with slash.
	url string
}

// New returns a new Smart-ID REST API client with the provided configuration.
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
	if c.authHashFunction, err = findAuthHashFunction(c.conf.AuthChallengeSize); err != nil {
		return nil, err
	}
	c.url = conf.URL
	if !strings.HasSuffix(c.url, "/") {
		c.url += "/"
	}
	return
}

// findAuthHashFunction is helper function to find allowed authentication hash
// algorithm by it's hash size. If the size is not configured, the first value
// in the allowedAuthHashFunctions is used.
func findAuthHashFunction(size int64) (crypto.Hash, error) {
	if size == 0 {
		return allowedAuthHashFunctions[0], nil
	}
	var sizes []int
	for _, hf := range allowedAuthHashFunctions {
		if hf.Size() == int(size) {
			return hf, nil
		}
		sizes = append(sizes, hf.Size())
	}
	return 0, AuthChallengeSizeError{Size: size, AllowedSizes: sizes}
}
