/*
Package container provides common code for handling signature containers.
*/
package container // import "ivxv.ee/container"

import (
	"crypto/x509"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ivxv.ee/yaml"
)

// Signature contains metadata about a signature.
type Signature struct {
	ID          string // ID uniquely identifies this signature in the container.
	Signer      *x509.Certificate
	Issuer      *x509.Certificate
	SigningTime time.Time
}

// Container is a container that protects data with one or multiple signatures.
// It is only necessary for Container to be able to read containers and not
// create them.
type Container interface {
	// Close closes and frees any resources held by this Container. Any
	// byte slices returned by Data can no longer be used and can be reused
	// by Container implementations.
	io.Closer

	// Signatures returns metadata about the signatures.
	Signatures() []Signature

	// Data returns the signed data protected by this Container. The data
	// is expected to be a set of byte slices with unique string keys.
	//
	// Note! It might seem to make more sense to not return all the
	// documents as in-memory byte slices, but rather as file references or
	// open readers to them. However, since they need to be completely read
	// in order to check container signatures, then it does not make sense
	// to have to read them twice and it is easier to just remember the
	// read data.
	//
	// The returned data must no longer be referenced after the container
	// is closed: implementations are free to reclaim the byte slices for
	// reuse.
	Data() map[string][]byte
}

// Conf is the container set configuration. It maps enabled container types to
// their configurations. The latter is listed as an unspecified YAML Node,
// which will be applied to the corresponding container type's configuration
// structure.
type Conf map[Type]yaml.Node

// Opener contains a configured set of container implementations.
type Opener map[Type]OpenFunc

// Configure configures a set of container implementations specified in the
// configuration.
func Configure(c Conf) (o Opener, err error) {
	o = make(Opener)

	// For each configured implementation, ...
	reglock.RLock()
	defer reglock.RUnlock()
	for t, y := range c {
		// ...check if it is linked ...
		re, ok := registry[t]
		if !ok {
			return nil, UnlinkedTypeError{Type: t}
		}

		// ...and if creating an opening function succeeds.
		f, err := re.n(y)
		if err != nil {
			return nil, ConfigureTypeError{Type: t, Err: err}
		}
		o[t] = f
	}

	return
}

// Open dispatches the encoded container to the correct opening function and
// returns the verified Container.
func (o Opener) Open(t Type, container io.Reader) (c Container, err error) {
	f, ok := o[t]
	if !ok {
		return nil, UnconfiguredTypeError{Type: t}
	}
	return f(container)
}

// OpenFile opens the file at path, and passes its contents to Open. The path
// must have an extension corresponding to the container type to use.
func (o Opener) OpenFile(path string) (c Container, err error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, OpenFileError{Path: path, Err: err}
	}
	defer fp.Close() // nolint: errcheck, ignore close failure of read-only fd.

	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	if len(ext) == 0 {
		return nil, MissingExtensionError{Path: path}
	}

	return o.Open(Type(ext), fp)
}

// UnverifiedOpen dispatches the encoded container to the correct unverified
// opening function and returns the unverified Container.
//
// Note! This is a dangerous function, which provides access to signed data
// without verifying the signatures and should only be used during
// bootstrapping. Even then, the signatures should be retroactively verified.
func UnverifiedOpen(t Type, container io.Reader) (c Container, err error) {
	reglock.RLock()
	defer reglock.RUnlock()
	re, ok := registry[t]
	if !ok {
		return nil, OpenUnlinkedTypeError{Type: t}
	}
	return re.o(container)
}
