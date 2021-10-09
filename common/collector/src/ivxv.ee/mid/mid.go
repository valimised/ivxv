/*
Package mid provides client for the MID-REST service.

https://github.com/SK-EID/MID
*/
package mid

import (
	"crypto"
	"crypto/x509"
	"strings"

	"ivxv.ee/cryptoutil"
	"ivxv.ee/ocsp"
)

var (
	// InputError wraps errors which are caused by bad input to mid functions.
	_ = InputError{Err: nil}

	// NotMIDUserError is the error returned if the submitted phone number
	// does not belong to a Mobile-ID user.
	_ = NotMIDUserError{}

	// SIMError is the error returned if there is a problem with user's
	// phone SIM card and the user should contact the service provider.
	_ = SIMError{Result: ""}

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

	// allowedAuthHashFunctions is list of hash functions allowed for
	// authentication. Since the hash function is determined by it's hash
	// size (Conf.AuthChallengeSize), the functions should have hash sizes
	// unique in the list. If Conf.AuthChallengeSize is zero, then the
	// first one from this list is used.
	allowedAuthHashFunctions = []crypto.Hash{crypto.SHA256, crypto.SHA384, crypto.SHA512}
)

// Conf contains the configurable options for the MID REST API client. It
// only contains serialized values such that it can easily be unmarshaled from
// a file.
type Conf struct {
	URL              string // URL of MID REST API.
	RelyingPartyUUID string // The UUID of the relying party, i.e service consumer.
	RelyingPartyName string // The name of the relying party, i.e service consumer.

	Language    string // Language for user dialog in mobile phone.
	AuthMessage string // Message to display during authentication.
	SignMessage string // Message to display during signing.

	// MessageFormat is the message format name sent to MID service.
	// Possible values are "GSM-7" (default by MID) and "UCS-2”. GSM-7
	// allows AuthMessage and SignMessage to contain up to 40 characters
	// from standard GSM 7-bit alpabet including up to 5 characters from
	// extension table (€[]^|{}\). UCS-2 allows up to 20 characters from
	// UCS-2 alphabet. More info in MID documentation:
	// https://github.com/SK-EID/MID#323-request-parameters.
	// More info about encoding: https://en.wikipedia.org/wiki/GSM_03.38.
	MessageFormat     string
	AuthChallengeSize int64 // The authentication challenge size, either 32, 48 or 64 bytes.
	StatusTimeoutMS   int64 // The long-polling timeout for authentication/signing status request.

	Roots         []string  // PEM-encoded authentication certificate verification roots.
	Intermediates []string  // PEM-encoded authentication certificate verification intermediates.
	OCSP          ocsp.Conf // OCSP configuration for checking authentication certificate revocation.
}

// Client implements MID REST API authentication and signing.
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

// New returns a new MID REST API client with the provided configuration.
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
