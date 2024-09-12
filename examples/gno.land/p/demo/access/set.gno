package access

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

// Set is an object containing the configuration and allowing the application of filters
// It is suited to be used as a contract-side global variable or can be embedded in another Go object
type Set struct {
	*ownable.Ownable           // owner in ownable is superuser
	authorized       *avl.Tree // std.Addr > struct{}{}
}

// New returns a Set object initialized with the caller as the unique authorized address
// It is recommended to use this function in the `init()` function of the calling realm
func New() *Set {
	a := &Set{
		ownable.New(),
		avl.NewTree(),
	}

	// Add owner to auth list
	a.authorized.Set(a.Owner().String(), struct{}{})
	return a
}

// NewWithAddress returns a Set object initialized with the provided address as authorized
func NewWithAddress(addr std.Address) *Set {
	a := &Set{
		ownable.NewWithAddress(addr),
		avl.NewTree(),
	}

	// Add owner to the set
	a.authorized.Set(a.Owner().String(), struct{}{})
	return a
}

//// List returns the AVL tree containing all the authorized addresses
//func (s Set) List() avl.Tree {
//	return *s.authorized
//}

// HasAccess checks if the provided address is in the list of authorized ones.
func (s Set) HasAccess(addr std.Address) bool {
	return s.authorized.Has(addr.String())
}

// CallerHasAccess checks if the caller or prevRealm is authorized.
func (s Set) CallerHasAccess() bool {
	return s.HasAccess(std.PrevRealm().Addr())
}

// AssertCallerHasAccess checks whether the std.GetOrigCaller or std.PrevRealm is whitelisted as authorized.
// If not, it panics indicating restricted access.
func (s *Set) AssertCallerHasAccess() {
	if !s.CallerHasAccess() {
		panic("access: caller not in set")
	}
}

// Add adds an address to the list of authorized addresses.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) Add(addr std.Address) {
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
	if len(s.addrs) == 0 {
		panic("should not happen")
	}

	// XXX: should we prevent deleting self?

	// Prevent deleting the last authorized address.
	if len(s.addrs) == 1 {
		panic("cannot have no authorized address.")
	}
	return s.ForceDel(addr)
}

// ReplaceAll removes all existing authorized addresses and replaces them with a new one.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) Replace(addrs []std.Address) {
	s.addrs = addrs
}
