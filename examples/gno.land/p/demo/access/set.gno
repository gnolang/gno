package access

import (
	"fmt"
	"std"
)

// New returns a Set object initialized with the caller as the unique authorized address.
// It is recommended to use this function in the `init()` function of the calling realm.
func New() Set {
	addr := std.GetOrigCaller()
	return Set{
		addrs: []std.Address{addr},
	}
}

// NewWithAddress returns a Set object initialized with the provided address as authorized.
func NewWithAddress(addr std.Address) Set {
	s := New()
	s.ReplaceAll(addr)
	return s
}

// Set is an object containing the configuration and allowing the application of filters.
// It is suited to be used as a contract-side global variable or can be embedded in another Go object.
type Set struct {
	addrs []std.Address
}

// List returns a slice containing all the authorized addresses.
func (s *Set) List() []std.Address {
	return s.addrs
}

// HasAccess checks if the provided address is in the list of authorized ones.
func (s *Set) HasAccess(addr std.Address) bool {
	for _, entry := range s.addrs {
		if entry == addr {
			return true
		}
	}
	return false
}

// CurrentHasAccess checks if the caller or prevRealm is authorized.
func (s *Set) CurrentHasAccess() bool {
	if s.HasAccess(std.GetOrigCaller()) {
		return true
	}

	// XXX: also check for std.PrevRealm, when merged.

	return false
}

// AssertCurrentHasAccess checks whether the std.GetOrigCaller or std.PrevRealm is whitelisted as authorized.
// If not, it panics indicating restricted access.
func (s *Set) AssertCurrentHasAccess() {
	if !s.CurrentHasAccess() {
		panic("restricted access.")
	}
}

// Add adds an address to the list of authorized addresses.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) Add(addr std.Address) {
	if s.addrs == nil {
		s.addrs = make([]std.Address)
	}
	if s.HasAccess(addr) {
		// XXX: Consider panicking or handling duplicate addition differently.
		return
	}
	s.addrs = append(s.addrs, addr)
}

// ForceDel removes an address from the list of authorized addresses.
// It won't panic if there is only one address.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) ForceDel(addr std.Address) (success bool) {
	for i, entry := range s.addrs {
		if entry == addr {
			s.addrs = append(s.addrs[:i], s.addrs[i+1:]...)
			return true
		}
	}
	return false
}

// Del removes an address from the list of authorized addresses.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) Del(addr std.Address) (success bool) {
	// XXX: should we prevent deleting self?

	// Prevent deleting the last authorized address.
	if len(s.addrs) == 1 {
		panic("cannot have no authorized address.")
	}
	return s.ForceDel(addr)
}

// ReplaceAll removes all existing authorized addresses and replaces them with a new one.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) ReplaceAll(addr std.Address) {
	s.addrs = []std.Address{addr}
}
