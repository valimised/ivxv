/*
Package identity provides an abstraction for extracting unique identifiers from
clients' Distinguished Name.
*/
package identity // import "ivxv.ee/identity"

import (
	"crypto/x509/pkix"
	"strings"
	"sync"
)

// Type identifies an identity extraction method.
type Type string

// Enumeration of identity types.
const (
	CommonName   Type = "commonname"   // Built-in.
	SerialNumber      = "serialnumber" // Built-in.
	PNOEE             = "pnoee"        // Built-in.
)

// Identifier is the type of functions that extract unique identifiers from
// Distinguished Names. Identifier must be safe for concurrent use.
type Identifier func(*pkix.Name) (string, error)

var (
	reglock  sync.RWMutex
	registry = map[Type]Identifier{
		CommonName:   commonName,
		SerialNumber: serialNumber,
		PNOEE:        pnoee,
	}
)

// Register registers an identity extraction method. It is intended to be
// called from init functions of packages that implement identity types.
func Register(t Type, i Identifier) {
	reglock.Lock()
	defer reglock.Unlock()
	registry[t] = i
}

// Get returns the Identifier corresponding to the requested Type.
func Get(t Type) (i Identifier, err error) {
	reglock.RLock()
	defer reglock.RUnlock()
	i, ok := registry[t]
	if !ok {
		return nil, UnlinkedTypeError{Type: t}
	}
	return i, nil
}

// commonName returns the CommonName from a Distinguished Name.
func commonName(name *pkix.Name) (id string, err error) {
	if name == nil {
		return "", CNEmptyDNError{}
	}
	id = name.CommonName
	if len(id) == 0 {
		err = EmptyCNError{}
	}
	return
}

// serialNumber returns the SerialNumber from a Distinguished Name.
func serialNumber(name *pkix.Name) (id string, err error) {
	if name == nil {
		return "", SerialEmptyDNError{}
	}
	id = name.SerialNumber
	if len(id) == 0 {
		err = EmptySerialError{}
	}
	return
}

// pnoee returns the SerialNumber from a Distinguished Name after removing an
// optional PNOEE- prefix. The prefix denotes the Estonian national personal
// number in accordance with ETSI EN 319 412-1 section 5.1.3.
func pnoee(name *pkix.Name) (id string, err error) {
	if name == nil {
		return "", PNOEEEmptyDNError{}
	}
	id, err = serialNumber(name)
	return strings.TrimPrefix(id, "PNOEE-"), err
}
