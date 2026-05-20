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

// Set stores a byte slice in the AVL tree using runtime.Caller().Address()
// as the key. The caller identity is sourced from a VM machine op rather
// than a passed-in `rlm realm` value — no rlm-typed parameter means
// nothing for an attacker to forge (omarsy-free pattern).
//
// Migrated from the v2 `(_ int, rlm realm, value)` shape to the v3a
// plain signature.
func (s *LociStore) Set(value []byte) {
	key := string(runtime.Caller().Address())
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
