/*
Package q11n provides common code for requesting qualifying properties for
signature containers.

Qualifying properties can be certificate statuses, timemarks and -stamps,
registration confirmations, etc.
*/
package q11n // import "ivxv.ee/q11n"

import (
	"context"
	"sync"

	"ivxv.ee/container"
	"ivxv.ee/yaml"
)

// Protocol identifies a signature container qualification protocol. The actual
// qualifier implementations are in other packages.
type Protocol string

// Enumeration of qualification protocols.
const (
	OCSP   Protocol = "ocsp"   // import "ivxv.ee/q11n/ocsp"
	OCSPTM Protocol = "ocsptm" // import "ivxv.ee/q11n/ocsp"
	TSP    Protocol = "tsp"    // import "ivxv.ee/q11n/tsp"
	TSPREG Protocol = "tspreg" // import "ivxv.ee/q11n/tsp"
)

// Qualifier is used for requesting qualifying properties for signature
// containers.
type Qualifier interface {
	// Qualify requests a qualifying property for a signed container and
	// returns it encoded as a byte slice.
	//
	// Implementations are free to require a specific container type, e.g.,
	// for querying extra properties not available through the Container
	// interface, but must check for it themselves.
	//
	// Implementations must obey cancellation signals from ctx.Done().
	Qualify(context.Context, container.Container) ([]byte, error)
}

// Here we "declare" error types, but instead of defining them ourselves, we
// want them to be generated so that they implement all the extra interfaces of
// generated errors.
//
// Although this is an error, it still nests an error to uniquely specify where
// the nesting error came from. So you would use these like
//
//	return q11n.BadCertificateStatusError{Err: UnderlyingError{}}
//
// where UnderlyingError will specify the package that returned the error.
var (
	// BadCertificateStatusError wraps errors which are caused by a
	// container signing certificate with bad status, e.g., revoked. This
	// can be returned by any qualifier which checks the status of the
	// signing certificate.
	_ = BadCertificateStatusError{Err: nil}
)

// NewFunc is the type of functions that create a signature container qualifier
// with the specified configuration and service directory. The latter can be
// used to pass private keys and other sensitive information to the qualifier.
type NewFunc func(yaml.Node, string) (Qualifier, error)

var (
	reglock  sync.RWMutex
	registry = make(map[Protocol]NewFunc)
)

// Register registers a signature container qualifier implementation. It is
// intended to be called from init functions of packages that implement
// qualifiers.
func Register(p Protocol, n NewFunc) {
	reglock.Lock()
	defer reglock.Unlock()
	registry[p] = n
}

// Conf is the qualifier set configuration. It contains an ordered list of
// qualifier protocols to use and their configurations. The latter is an
// unspecified YAML Node, which will be applied to the corresponding qualifier
// protocol's configuration structure.
type Conf []struct {
	Protocol Protocol
	Conf     yaml.Node
}

// Qualifiers is a list of qualifiers in the same order they were presented in
// the configuration. This is also the order in which the qualification
// requests should be made.
type Qualifiers []struct {
	Protocol  Protocol
	Qualifier Qualifier
}

// Configure configures a list of qualifier implementations specified in the
// configuration. sensitive is the path to the service instance directory which
// can contain sensitive information, e.g., request signing keys.
func Configure(c Conf, sensitive string) (qs Qualifiers, err error) {
	qs = make(Qualifiers, len(c))

	// For each configured implementation, ...
	reglock.RLock()
	defer reglock.RUnlock()
	for i, p := range c {
		// ...check if it is linked ...
		n, ok := registry[p.Protocol]
		if !ok {
			return nil, UnlinkedProtocolError{Protocol: p.Protocol}
		}
		qs[i].Protocol = p.Protocol

		// ...and if creating the qualifier succeeds.
		qs[i].Qualifier, err = n(p.Conf, sensitive)
		if err != nil {
			return nil, ConfigureProtocolError{Protocol: p.Protocol, Err: err}
		}
	}
	return
}

// Properties is a map from qualifier protocols to qualifying properties. It is
// a convenience type to be used outside of q11n to store the results of
// qualification.
type Properties map[Protocol][]byte
