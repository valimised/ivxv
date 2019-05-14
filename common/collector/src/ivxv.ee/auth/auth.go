/*
Package auth provides common code for authenticating clients.
*/
package auth // import "ivxv.ee/auth"

import (
	"context"
	"crypto/x509/pkix"
	"sync"

	"ivxv.ee/yaml"
)

// Type identifies an authentication verifier. The actual implementations are
// in other packages.
type Type string

// Enumeration of authentication verifiers.
const (
	Dummy  Type = "dummy"  // import "ivxv.ee/auth/dummy"
	TLS    Type = "tls"    // import "ivxv.ee/auth/tls"
	Ticket Type = "ticket" // import "ivxv.ee/auth/ticket"
)

// Here we "declare" errors that authentication modules should use for wrapping
// errors so they can be classified.
var (
	// MalformedTokenError wraps errors which are cause by malformed
	// authentication tokens.
	_ = MalformedTokenError{Err: nil}

	// CertificateError wraps errors which are caused by invalid
	// authentication certificates.
	_ = CertificateError{Err: nil}

	// UnauthorizedError wraps errors where authentication succeeded, but
	// the client is not authorized to use any services.
	_ = UnauthorizedError{Err: nil}
)

// Verifier is a client authentication token verifier. An authentication token
// can be anything from a X.509 certificate used during TLS client
// authentication to an identity proof constructed by a trusted external
// service.
type Verifier interface {
	// Verify verifies the authentication token and returns the
	// authenticated client's PKIX name. The content of token depends on
	// the implementation. The returned error can only be nil if
	// authentication succeeded.
	//
	// Context is provided since some authentication verifiers might need
	// to contact external services.
	Verify(ctx context.Context, token []byte) (*pkix.Name, error)
}

// VoteIdentifier is an optional additional interface that Verifiers can
// implement. If a Verifier is also a VoteIdentifier, then it should be used to
// retrieve the vote identifier for storing the vote submitted using the
// authentication token.
//
// This is necessary for single-use tokens to ensure that they cannot be
// used for multiple votes. Therefore an authentication token must always
// return the same vote identifier.
type VoteIdentifier interface {
	VoteIdentifier(token []byte) (voteID []byte, err error)
}

// NewFunc is the type of functions that an authentication verifier with a
// specified configuration.
type NewFunc func(yaml.Node) (Verifier, error)

var (
	reglock  sync.RWMutex
	registry = make(map[Type]NewFunc)
)

// Register registers an authentication verifier implementation. It is intended
// to be called from init functions of packages that implement authentication
// verifiers.
func Register(t Type, n NewFunc) {
	reglock.Lock()
	defer reglock.Unlock()
	registry[t] = n
}

// Conf is the authentication verifier set configuration. It maps enabled
// authentication verifiers to their configurations. The latter is listed as an
// unspecified YAML Node, which will be applied to the corresponding
// authentication verifier's configuration structure.
type Conf map[Type]yaml.Node

// Auther contains a configured set of authentication verifiers.
type Auther map[Type]Verifier

// Configure configures a set of authentication verifiers specified in the
// configuration.
func Configure(c Conf) (a Auther, err error) {
	a = make(Auther)

	// For each configured implementation, ...
	reglock.RLock()
	defer reglock.RUnlock()
	for t, y := range c {
		// ..check if it is linked...
		n, ok := registry[t]
		if !ok {
			return nil, UnlinkedTypeError{Type: t}
		}

		// ..and if creating a verifier succeeds.
		v, err := n(y)
		if err != nil {
			return nil, ConfigureTypeError{Type: t, Err: err}
		}
		a[t] = v
	}

	return
}

// Verify dispatches the authentication token to the verifier of type t and
// returns the authenticated client's name.
//
// If the type also implements VoteIdentifier, then it returns the vote
// identifier to store the vote with.
func (a Auther) Verify(ctx context.Context, t Type, token []byte) (
	name *pkix.Name, voteID []byte, err error) {

	v, ok := a[t]
	if !ok {
		return nil, nil, UnconfiguredTypeError{Type: t}
	}
	if name, err = v.Verify(ctx, token); err != nil {
		return nil, nil, err
	}
	if vid, ok := v.(VoteIdentifier); ok {
		if voteID, err = vid.VoteIdentifier(token); err != nil {
			return nil, nil, err
		}
	}
	return
}
