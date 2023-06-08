package administrable

import (
	"fmt"
	"std"
)

// New returns s Set object initialized with the caller as the unique admin.
// It is recommended to use this function in the `init()` function of the calling realm.
func New() Set {
	addr := std.GetOrigCaller()
	return Set{
		addrs: []std.Address{addr},
	}
}

// NewWithAddress returns s Set object initialized with the provided address as the admin.
func NewWithAddress(addr std.Address) Set {
	s := New()
	s.ReplaceAll(addr)
	return s
}

// Set is an object containing the configuration and allowing to apply filters.
// It's suited to be used as a contract-side global variable or can be embedded in another go object.
type Set struct {
	// XXX: replace with map[std.Address]struct{}?
	addrs []std.Address
}

// List returns a slice containing the addresses of all administrators.
func (s *Set) List() []std.Address {
	return s.addrs
}

// Has checks if an address is in the list of administrators.
func (s *Set) HasAccess(addr std.Address) bool {
	for _, entry := range s.addrs {
		if entry == addr {
			return true
		}
	}
	return false
}

// CurrentHasAccess checks if the caller or prevRealm is an administrator.
func (s *Set) CurrentHasAccess() bool {
	if s.HasAccess(std.GetOrigCaller()) {
		return true
	}

	// XXX: also check for std.PrevRealm, when merged.

	return false
}

// AssertCurrentHasAccess checks whether the std.GetOrigCaller or std.PrevRealm is whitelisted as an admin.
// If not, it panics indicating restricted access.
func (s *Set) AssertCurrentHasAccess() {
	if !s.CurrentHasAccess() {
		panic("restricted ares.")
	}
}

// Add adds an address to the list of administrators.
// It requires the caller or prevRealm to be an admin, otherwise, it panics.
func (s *Set) Add(addr std.Address) {
	if s.HasAccess(addr) {
		// XXX: panic?
		return
	}
	s.addrs = append(s.addrs, addr)
}

// Del removes an address from the list of administrators.
// It requires the caller or prevRealm to be an admin, otherwise, it panics.
func (s *Set) Del(addr std.Address) (success bool) {
	// XXX: should we prevent deleting self?

	// prevent deleting the last admin.
	if len(s.addrs) == 1 {
		panic("cannot have no admin.")
	}

	for i, entry := range s.addrs {
		if entry == addr {
			s.addrs = append(s.addrs[:i], s.addrs[i+1:]...)
			return true
		}
	}
	return false
}

// ReplaceAll removes all existing admins and replaces them with a new one.
// It requires the caller or prevRealm to be an admin, otherwise, it panics.
func (s *Set) ReplaceAll(addr std.Address) {
	s.addrs = []std.Address{addr}
}
