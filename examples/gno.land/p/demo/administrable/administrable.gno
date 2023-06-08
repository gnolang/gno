package administrable

import (
	"fmt"
	"std"
)

// New returns an Administrable object initialized with the caller as the unique admin.
// It is recommended to use this function in the `init()` function of the calling realm.
func New() *Administrable {
	return NewWithAddress(std.GetOrigCaller())
}

// NewWithAddress returns an Administrable object initialized with the provided address as the admin.
func NewWithAddress(addr std.Address) *Administrable {
	a := Administrable{
		addrs: make(map[std.Address]struct{}),
	}
	a.add(addr)
	return &a
}

// Administrable is an object containing the configuration and allowing to apply filters.
// It's suited to be used as a contract-side global variable or can be embedded in another go object.
type Administrable struct {
	addrs map[std.Address]struct{}
}

// List returns a slice containing the addresses of all administrators.
func (a *Administrable) List() []std.Address {
	list := make([]std.Address, len(a.addrs))
	idx := 0
	for addr := range a.addrs {
		list[idx] = addr
		idx++
	}
	return list
}

// Has checks if an address is in the list of administrators.
func (a *Administrable) Has(addr std.Address) bool {
	_, found := a.addrs[addr]
	return found
}

// IsAdmin checks if the caller or prevRealm is an administrator.
func (a *Administrable) IsAdmin() bool {
	if a.Has(std.GetOrigCaller()) {
		return true
	}

	// XXX: also check for std.PrevRealm, when merged.

	return false
}

// AssertIsAdmin checks whether the std.GetOrigCaller or std.PrevRealm is whitelisted as an admin.
// If not, it panics indicating restricted access.
func (a *Administrable) AssertIsAdmin() {
	if !a.IsAdmin() {
		panic("restricted area.")
	}
}

// Add adds an address to the list of administrators.
// It requires the caller or prevRealm to be an admin, otherwise, it panics.
func (a *Administrable) Add(addr std.Address) {
	a.AssertIsAdmin()
	a.add(addr)
}

// Del removes an address from the list of administrators.
// It requires the caller or prevRealm to be an admin, otherwise, it panics.
func (a *Administrable) Del(addr std.Address) {
	a.AssertIsAdmin()
	a.del(addr)
}

func (a *Administrable) add(addr std.Address) {
	// XXX: should we panic if addr was already present?
	a.addrs[addr] = struct{}{}
}

func (a *Administrable) del(addr std.Address) {
	// XXX: should we prevent deleting self?
	if len(a.addrs) == 0 {
		panic("cannot have no admin.")
	}
	delete(a.addrs, addr)
}
