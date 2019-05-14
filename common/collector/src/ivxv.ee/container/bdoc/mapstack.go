package bdoc

// mapstack is a stack of string to string maps. Modifications to the map are
// made in layers and can be removedd as one. Lookup is performed recursively
// through all layers.
type mapstack []map[string]string

// push adds a new empty layer to the stack.
func (s *mapstack) push() {
	*s = append(*s, nil)
}

// set sets a key-value in the top-most layer. Panics if there are no layers.
func (s *mapstack) set(key, value string) {
	ds := *s // Dereferenced copy to simplify following code.
	m := ds[len(ds)-1]
	if m == nil {
		m = make(map[string]string)
		(*s)[len(ds)-1] = m
	}
	m[key] = value
}

// pop removes the top-most layer. Panics if there are no layers.
func (s *mapstack) pop() {
	*s = (*s)[:len(*s)-1]
}

// get gets the value of key, looking from the top-most layer down.
func (s *mapstack) get(key string) (string, bool) {
	for i := len(*s) - 1; i >= 0; i-- {
		if value, ok := (*s)[i][key]; ok {
			return value, ok
		}
	}
	return "", false
}

// flatten returns a flat map with all key-values currently in the stack.
func (s *mapstack) flatten() map[string]string {
	flat := make(map[string]string)
	for _, m := range *s {
		for key, value := range m {
			flat[key] = value
		}
	}
	return flat
}
