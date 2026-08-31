// Package addrset provides a set of blockchain addresses backed by a
// B+ tree, with a read-only view type for safe cross-realm exposure.
//
// It mirrors the gno.land/p/moul/addrset API on a
// gno.land/p/nt/bptree/v0 backing: a B+ tree packs many entries per
// persisted node, so a stored address costs ~0.9 KB vs the
// one-node-per-entry AVL backing's ~2.0 KB (2.2x asymptotically, 1.6x
// at 10 entries; insert gas ~2.1x less). Prefer this package when sets
// are part of persisted realm state; the omitted Tree() escape hatch is
// deliberate, so the backing store never leaks.
//
// Two behavioral differences from the AVL-backed moul package, both
// consequences of the in-place-mutating backing:
//
//   - the set must NOT be mutated (Add/Remove) from inside an iteration
//     callback — the AVL backing's copy-on-write tolerated it, this one
//     does not;
//   - do not copy a non-zero Set by value — the copies share live tree
//     nodes while their roots and sizes diverge (the AVL backing's
//     copies were independent snapshots).
//
// Example:
//
//	var set addrset.Set // the zero value is an empty, usable set
//
//	set.Add(addr)   // true (newly added)
//	set.Has(addr)   // true
//	set.Remove(addr) // true (was present)
package addrset

import "gno.land/p/nt/bptree/v0"

// Set stores a set of addresses in sorted order. The zero value is an
// empty, usable set.
type Set struct {
	tree bptree.BPTree
}

// Add inserts an address into the set.
// Returns true if the address was newly added, false if it already existed.
func (s *Set) Add(addr address) bool {
	return !s.tree.Set(string(addr), nil)
}

// Remove deletes an address from the set.
// Returns true if the address was found and removed, false if it didn't exist.
func (s *Set) Remove(addr address) bool {
	_, removed := s.tree.Remove(string(addr))
	return removed
}

// Has checks if an address exists in the set.
func (s *Set) Has(addr address) bool {
	return s.tree.Has(string(addr))
}

// Size returns the number of addresses in the set.
func (s *Set) Size() int {
	return s.tree.Size()
}

// IterateByOffset walks through addresses in sorted order, starting at
// the given offset and visiting up to count addresses. The callback
// returns true to stop iteration. The set must not be modified during
// iteration (no Add or Remove from the callback).
func (s *Set) IterateByOffset(offset int, count int, cb func(addr address) bool) {
	s.tree.IterateByOffset(offset, count, func(key string, _ any) bool {
		return cb(address(key))
	})
}

// ReverseIterateByOffset walks through addresses in reverse (descending)
// order, starting at the given offset (counted from the end) and
// visiting up to count addresses. The callback returns true to stop
// iteration. The set must not be modified during iteration (no Add or
// Remove from the callback).
func (s *Set) ReverseIterateByOffset(offset int, count int, cb func(addr address) bool) {
	s.tree.ReverseIterateByOffset(offset, count, func(key string, _ any) bool {
		return cb(address(key))
	})
}
