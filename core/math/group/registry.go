package group

import (
	"fmt"
	"sync"
)

// registry is a map with sync.RWMutex. And its purpose is to hold all group
// implementations.
//
// sync.Map itself is optimized for get operations. It is assumed that
// Register will be infrequent, otherwise no advantage of using sync.Map.
var registry = new(sync.Map)

// Register registers a group implementation in a registry.
func Register(G Group) {
	registry.Store(G.Name(), G)
}

// Get gets a group implementation by its name.
// NB! Caller should have 'all' package imported as:
//
//	_ import "tivi.io/core/math/group/all
func Get(name string) (Group, error) {
	G, ok := registry.Load(name)
	if !ok {
		return nil, fmt.Errorf("math/group: unknown group: %s", name)
	}

	// Safe to cast, g is checked at compile time
	return G.(Group), nil
}

// All returns all registered group implementations.
func All() []Group {
	groups := make([]Group, 0)

	registry.Range(func(_, G any) bool {
		groups = append(groups, G.(Group))
		// If you `return false`, then Range will stop and exit
		return true
	})

	return groups
}
