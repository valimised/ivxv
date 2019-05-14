package storage

import (
	"context"
	"sync"

	"ivxv.ee/yaml"
)

// Protocol identifies a storage protocol. The actual protocol client
// implementations are in other packages.
type Protocol string

// Enumeration of storage protocols.
const (
	Memory Protocol = "memory" // import "ivxv.ee/storage/memory"
	File   Protocol = "file"   // import "ivxv.ee/storage/file"
	Etcd   Protocol = "etcd"   // import "ivxv.ee/storage/etcd"
)

// Here we "declare" error types, but instead of defining them ourselves, we
// want them to be generated so that they implement all the extra interfaces of
// generated errors.
//
// Although these are errors in themselves, they still all nest an error to
// uniquely specify where the nesting error came from. So you would use these
// like
//
//	return storage.ExistError{Key: key, Err: ExistingKeyError{}}
//
// where ExistingKeyError will specify the package that returned the error.
var (
	// ExistError wraps errors which are caused by a key already existing.
	_ = ExistError{Key: "", Err: nil}

	// NotExistError wraps errors which are caused by a key not existing.
	_ = NotExistError{Key: "", Err: nil}

	// UnexpectedValueError wraps errors which are caused by
	// compare-and-swap encountering an unexpected old value.
	_ = UnexpectedValueError{Key: "", Err: nil}
)

// GetWithPrefixResult is a single result from PutGetter.GetWithPrefix.
type GetWithPrefixResult struct {
	Key   string
	Value []byte
}

// PutGetter is the interface that must be implemented by storage protocol
// clients and is used by Client as the underlying protocol. By design, it does
// not have methods for deleting or overwriting values (except for CASAndGet),
// because we want to keep everything we store.
type PutGetter interface {
	// Put puts value into the storage service with the given key. It is an
	// error for the key to already exist. Implementations must obey
	// cancellation signals from ctx.Done().
	Put(ctx context.Context, key string, value []byte) error

	// Get returns the value stored with the given key. It is an error for
	// the key to not exist. Implementations must obey cancellation signals
	// from ctx.Done().
	Get(ctx context.Context, key string) ([]byte, error)

	// GetWithPrefix returns the stored keys and values where the key has
	// the provided prefix.
	//
	// The result channel should be read until closed after which the error
	// channel should be read to determine if the first one was closed due
	// to completion or an error. If reading from the result channel is
	// stopped early, then the context should be cancelled to stop
	// GetWithPrefix from sending more results.
	//
	// Implementations must ensure that at least one (possibly nil) error
	// can be read from the error channel after the result channel is
	// closed. This can be done by simply closing the channel.
	//
	// Implementations must obey cancellation signals from ctx.Done().
	GetWithPrefix(ctx context.Context, prefix string) (<-chan GetWithPrefixResult, <-chan error)

	// CAS compare-and-swaps the value stored with the key cas. This is the
	// only method which can overwrite values and is only meant for read
	// counters.
	//
	// Implementations must obey cancellation signals from ctx.Done().
	CAS(ctx context.Context, cas string, old, new []byte) error
}

// PutAllRequest is a single key-value pair to put in a Batcher.PutAll request.
type PutAllRequest struct {
	Key   string
	Value []byte
}

// Batcher is an optional interface that PutGetters can implement for batch
// processing. Storage operations will prefer methods in this interface when
// working on multiple keys at once.
type Batcher interface {
	PutGetter

	// BatchSize returns the maximum number of operations that can be
	// batched by this protocol. Other Batcher methods return errors if
	// more than BatchSize operations are given.
	BatchSize() int

	// GetAll returns all the values stored with the given keys. If a key
	// is missing, then it will not be included in the returned map: it is
	// the caller's responsibility to ensure that all required keys are
	// returned.
	//
	// Implementations must obey cancellation signals from ctx.Done().
	GetAll(ctx context.Context, keys ...string) (map[string][]byte, error)

	// PutAll puts all the key-value pairs into the storage service. It is
	// an error for ANY key to exist. On error, no values will be stored.
	// Implementations must obey cancellation signals from ctx.Done().
	PutAll(ctx context.Context, reqs ...PutAllRequest) error
}

// NewFunc is the type of functions that create a storage protocol client with
// a provided configuration and service instance information.
type NewFunc func(yaml.Node, *Services) (PutGetter, error)

var (
	reglock  sync.RWMutex
	registry = make(map[Protocol]NewFunc)
)

// Register registers a storage protocol client implementation. It is intended
// to be called from init functions of packages that implement storage protocol
// clients.
//
// n is a constructor function used to create a new storage protocol client
// with provided configuration.
func Register(p Protocol, n NewFunc) {
	reglock.Lock()
	defer reglock.Unlock()
	registry[p] = n
}
