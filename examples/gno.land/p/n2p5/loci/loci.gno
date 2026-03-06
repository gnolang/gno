// loci is a single purpose datastore keyed by the caller's address. It has two
// functions: Set and Get. loci is plural for locus, which is a central or core
// place where something is found or from which it originates. In this case,
// it's a simple key-value store where an address (the key) can store exactly
// one value (in the form of a byte slice). Only the caller can set the value
// for their address, but anyone can retrieve the value for any address.
package loci

import (
	"chain/runtime"

	"gno.land/p/nt/avl/v0"
)

// LociStore is a simple key-value store that uses
// an AVL tree to store the data.
type LociStore struct {
	internal *avl.Tree
}

// New creates a reference to a new LociStore.
func New() *LociStore {
	return &LociStore{
		internal: avl.NewTree(),
	}
}

// Set stores a byte slice in the AVL tree using the `std.PreviousRealm().Address()`
// string as the key.
func (s *LociStore) Set(value []byte) {
	key := string(runtime.PreviousRealm().Address())
	s.internal.Set(key, value)
}

// Get retrieves a byte slice from the AVL tree using the provided address.
// The return values are the byte slice value and a boolean indicating
// whether the value exists.
func (s *LociStore) Get(addr address) []byte {
	value, exists := s.internal.Get(string(addr))
	if !exists {
		return nil
	}
	return value.([]byte)
}
