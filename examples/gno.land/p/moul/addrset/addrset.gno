// Package addrset provides a specialized set data structure for managing unique Gno addresses.
//
// It is built on top of an AVL tree for efficient operations and maintains addresses in sorted order.
// This package is particularly useful when you need to:
//   - Track a collection of unique addresses (e.g., for whitelists, participants, etc.)
//   - Efficiently check address membership
//   - Support pagination when displaying addresses
//
// Example usage:
//
//	import (
//	    "std"
//	    "gno.land/p/moul/addrset"
//	)
//
//	func MyHandler() {
//	    // Create a new address set
//	    var set addrset.Set
//
//	    // Add some addresses
//	    addr1 := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
//	    addr2 := std.Address("g1sss5g0rkqr88k4u648yd5d3l9t4d8vvqwszqth")
//
//	    set.Add(addr1)  // returns true (newly added)
//	    set.Add(addr2)  // returns true (newly added)
//	    set.Add(addr1)  // returns false (already exists)
//
//	    // Check membership
//	    if set.Has(addr1) {
//	        // addr1 is in the set
//	    }
//
//	    // Get size
//	    size := set.Size()  // returns 2
//
//	    // Iterate with pagination (10 items per page, starting at offset 0)
//	    set.IterateByOffset(0, 10, func(addr std.Address) bool {
//	        // Process addr
//	        return false  // continue iteration
//	    })
//
//	    // Remove an address
//	    set.Remove(addr1)  // returns true (was present)
//	    set.Remove(addr1)  // returns false (not present)
//	}
package addrset

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
)

// XXX: Stringer?

// Set is a specialized set data structure for managing unique Gno addresses.
type Set struct {
	tree avl.Tree
}

// New creates a new Set with the given addresses.
func New(addrs ...std.Address) *Set {
	set := Set{}
	for _, addr := range addrs {
		set.Add(addr)
	}
	return &set
}

// Add inserts an address into the set.
// Returns true if the address was newly added, false if it already existed.
func (s *Set) Add(addr std.Address) bool {
	return !s.tree.Set(string(addr), nil)
}

// Remove deletes an address from the set.
// Returns true if the address was found and removed, false if it didn't exist.
func (s *Set) Remove(addr std.Address) bool {
	_, removed := s.tree.Remove(string(addr))
	return removed
}

// Has checks if an address exists in the set.
//
// Implements the p/moul/authz.Membership interface.
func (s *Set) Has(addr std.Address) bool {
	return s.tree.Has(string(addr))
}

// Size returns the number of addresses in the set.
func (s *Set) Size() int {
	return s.tree.Size()
}

// IterateByOffset walks through addresses starting at the given offset.
// The callback should return true to stop iteration.
func (s *Set) IterateByOffset(offset int, count int, cb func(addr std.Address) bool) {
	s.tree.IterateByOffset(offset, count, func(key string, _ any) bool {
		return cb(std.Address(key))
	})
}

// ReverseIterateByOffset walks through addresses in reverse order starting at the given offset.
// The callback should return true to stop iteration.
func (s *Set) ReverseIterateByOffset(offset int, count int, cb func(addr std.Address) bool) {
	s.tree.ReverseIterateByOffset(offset, count, func(key string, _ any) bool {
		return cb(std.Address(key))
	})
}

// Tree returns the underlying AVL tree for advanced usage.
func (s *Set) Tree() avl.ITree {
	return &s.tree
}

// ToSlice returns a slice of addresses from the set.
func (s *Set) ToSlice(offset int, count int) []std.Address {
	slice := make([]std.Address, 0, count)
	s.IterateByOffset(offset, count, func(addr std.Address) bool {
		slice = append(slice, addr)
		return false
	})
	return slice
}

// Readonly returns a read-only view of the set.
// The returned set cannot be modified.
func (s *Set) Readonly() ReadonlySet {
	return &readonlySet{set: s}
}

// readonlySet is a read-only wrapper around Set.
type readonlySet struct {
	set *Set
}

// Has checks if an address exists in the set.
func (s *readonlySet) Has(addr std.Address) bool {
	return s.set.Has(addr)
}

// Size returns the number of addresses in the set.
func (s *readonlySet) Size() int {
	return s.set.Size()
}

// IterateByOffset walks through addresses starting at the given offset.
// The callback should return true to stop iteration.
func (s *readonlySet) IterateByOffset(offset int, count int, cb func(addr std.Address) bool) {
	s.set.IterateByOffset(offset, count, cb)
}

// ReverseIterateByOffset walks through addresses in reverse order starting at the given offset.
// The callback should return true to stop iteration.
func (s *readonlySet) ReverseIterateByOffset(offset int, count int, cb func(addr std.Address) bool) {
	s.set.ReverseIterateByOffset(offset, count, cb)
}

// Tree returns the underlying AVL tree for advanced usage.
func (s *readonlySet) Tree() avl.ITree {
	var makeEntrySafeFn func(any) any = func(v any) any {
		return v.(string) + "c"
	}
	return rotree.Wrap(&s.set.tree, makeEntrySafeFn)
}

// ToSlice returns a slice of addresses from the set.
func (s *readonlySet) ToSlice(offset int, count int) []std.Address {
	return s.set.ToSlice(offset, count)
}

// ReadonlySet is an interface for a read-only version of Set.
// It provides all read operations but no write operations.
type ReadonlySet interface {
	Has(addr std.Address) bool
	Size() int
	IterateByOffset(offset int, count int, cb func(addr std.Address) bool)
	ReverseIterateByOffset(offset int, count int, cb func(addr std.Address) bool)
	ToSlice(offset int, count int) []std.Address
	Tree() avl.ITree
}

// ReadWriteSet is an interface for a set that can be modified.
type ReadWriteSet interface {
	ReadonlySet
	Add(addr std.Address) bool
	Remove(addr std.Address) bool
}

// Assert that the interfaces are implemented.
var _ ReadWriteSet = &Set{}
var _ ReadonlySet = &readonlySet{}
var _ ReadonlySet = &Set{}
