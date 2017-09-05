//ivxv:development

/*
Package memory implements a storage protocol which stores everything in memory,
used for testing.

The contents of the memory storage protocol can either be initialized via the
New method or through the configuration when using the storage protocol
registry interface.

The configuration is a simple YAML mapping with string values which will be
used as the initial contents. As an exception, when a string value is wrapped
with "base64(" and ")", then the contents will be Base64 decoded during
initialization. This way we can use raw literals in most cases, which is useful
for testing, but still have a way for specifying raw byte values. For example:

	/some/string/value: foobar
	/some/byte/value:   base64(Zm9vYmFy)

*/
package memory // import "ivxv.ee/storage/memory"

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"

	"ivxv.ee/storage"
	"ivxv.ee/yaml"
)

func init() {
	storage.Register(storage.Memory, func(n yaml.Node, _ *storage.Services) (
		s storage.PutGetter, err error) {

		m := New(nil)

		// Only apply if n is non-nil, otherwise the map we just made
		// in New will be replaced by nil.
		if n != nil {
			if err = yaml.Apply(n, &m.db); err != nil {
				return nil, ConfigurationApplicationError{Err: err}
			}
		}
		for k, v := range m.db {
			if m.db[k], err = decode(v); err != nil {
				return nil, DecodeValueError{Key: k, Err: err}
			}
		}
		return m, nil
	})
}

// decode checks if s has a "base64(" prefix and a ")" suffix and Base64
// decodes the inner content. If the prefix or suffix is not found, then s is
// returned unmodified.
func decode(s string) (string, error) {
	const prefix, suffix = "base64(", ")"
	t := strings.TrimSpace(s)
	if strings.HasPrefix(t, prefix) && strings.HasSuffix(t, suffix) {
		b, err := base64.StdEncoding.DecodeString(t[len(prefix) : len(t)-len(suffix)])
		return string(b), err
	}
	return s, nil
}

// M maps keys to values in the memory storage protocol.
type M struct {
	db   map[string]string
	lock sync.Mutex
}

// New creates a new M with the provided initial contents.
func New(initial map[string]string) *M {
	m := &M{db: make(map[string]string)}
	for k, v := range initial {
		m.db[k] = v
	}
	return m
}

// Put implements the storage.PutGetter interface.
func (m *M) Put(ctx context.Context, key string, value []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.db[key]; ok {
		return storage.ExistError{Key: key, Err: PutExistingKeyError{}}
	}
	m.db[key] = string(value)
	return nil
}

// Get implements the storage.PutGetter interface.
func (m *M) Get(ctx context.Context, key string) ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	value, ok := m.db[key]
	if !ok {
		return nil, storage.NotExistError{Key: key, Err: GetMissingKeyError{}}
	}
	return []byte(value), nil
}

// GetWithPrefix implements the storage.PutGetter interface. GetWithPrefix
// blocks all other operations on M until the last result is read.
func (m *M) GetWithPrefix(ctx context.Context, prefix string) (
	<-chan storage.GetWithPrefixResult, <-chan error) {

	m.lock.Lock()
	c := make(chan storage.GetWithPrefixResult)
	errc := make(chan error, 1)
	go func() {
		defer m.lock.Unlock()
		defer close(errc)
		defer close(c)
		for k, v := range m.db {
			if strings.HasPrefix(k, prefix) {
				select {
				case c <- storage.GetWithPrefixResult{Key: k, Value: []byte(v)}:
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
			}
		}
	}()
	return c, errc
}

// CASAndGet implements the storage.PutGetter interface.
func (m *M) CASAndGet(ctx context.Context, cas string, old, new []byte, keys ...string) (
	map[string][]byte, error) {

	m.lock.Lock()
	defer m.lock.Unlock()

	// Compare the old value.
	value, ok := m.db[cas]
	if !ok {
		return nil, storage.NotExistError{Key: cas, Err: CASMissingCASKeyError{}}
	}
	if value != string(old) {
		return nil, storage.UnexpectedValueError{
			Key: cas,
			Err: CASValueMismatchError{
				Have:     value,
				Expected: string(old),
			}}
	}

	// Get the requested values.
	ret := make(map[string][]byte)
	for _, key := range keys {
		value, ok = m.db[key]
		if !ok {
			return nil, storage.NotExistError{Key: key, Err: CASMissingKeyError{}}
		}
		ret[key] = []byte(value)
	}

	// Set the compared value.
	m.db[cas] = string(new)
	return ret, nil
}
