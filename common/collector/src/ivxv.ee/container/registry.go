package container

import (
	"io"
	"sync"

	"ivxv.ee/yaml"
)

// Type identifies a signature container type. The actual implementations are
// in other packages.
type Type string

// Enumeration of container types.
const (
	Dummy Type = "dummy" // import "ivxv.ee/container/dummy"
	BDOC       = "bdoc"  // import "ivxv.ee/container/bdoc"
)

// OpenFunc is the type of functions that parse and verify an encoded container
// into a Container.
type OpenFunc func(io.Reader) (Container, error)

// NewFunc is the type of functions that create a container opening function
// with a specified configuration.
type NewFunc func(yaml.Node) (OpenFunc, error)

// UnverifiedOpenFunc is the type of functions that open a container without
// needing any configuration and verifying any signatures.
type UnverifiedOpenFunc func(io.Reader) (Container, error)

type regentry struct {
	n NewFunc
	o UnverifiedOpenFunc
}

var (
	reglock  sync.RWMutex
	registry = make(map[Type]regentry)
)

// Register registers a signature container implementation. It is intended to
// be called from init functions of packages that implement container types.
//
// n is a constructor function used to create container opening functions with
// a specified configuration. o is a container opening function which is used
// during bootstrapping to open a container without needing any configuration
// and verifying its signatures.
func Register(t Type, n NewFunc, o UnverifiedOpenFunc) {
	reglock.Lock()
	defer reglock.Unlock()
	registry[t] = regentry{n, o}
}
