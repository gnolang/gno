// loci is a single purpose datastore keyed by the caller's address. It has two
// functions: Set and Get. loci is plural for locus, which is a central or core
// place where something is found or from which it originates. In this case,
// it's a simple key-value store where an address (the key) can store exactly
// one value (in the form of a byte slice). Only the caller can set the value
// for their address, but anyone can retrieve the value for any address.
package loci

import (
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

// Set stores a byte slice in the AVL tree using `rlm.Previous().Address()`
// as the key. rlm must be the caller's own captured cur (asserted via
// rlm.IsCurrent()); the previous walk used to be runtime.PreviousRealm()
// which is forgeable from non-crossing wrappers — an attacker realm
// could overwrite another realm's slot. With IsCurrent + rlm.Previous(),
// the slot key is unforgeable.
func (s *LociStore) Set(_ int, rlm realm, value []byte) {
	if !rlm.IsCurrent() {
		panic("loci: rlm does not match the current crossing frame")
	}
	key := string(rlm.Previous().Address())
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
