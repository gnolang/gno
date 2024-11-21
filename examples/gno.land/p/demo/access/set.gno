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
	s := &Set{
		ownable.New(),
		avl.NewTree(),
	}

	// Add owner to auth list
	s.authorized.Set(a.Owner().String(), struct{}{})
	return s
}

// NewWithAddress returns a Set object initialized with the provided address as authorized
func NewWithAddress(addr std.Address) *Set {
	s := &Set{
		ownable.NewWithAddress(addr),
		avl.NewTree(),
	}

	// Add owner to the set
	s.authorized.Set(a.Owner().String(), struct{}{})
	return s
}

// List returns a list containing all the authorized addresses
func (s Set) List() []std.Address {
	var addrs []std.Address
	if s.authorized.Size() == 0 {
		return addrs
	}
	// todo test
	s.authorized.Iterate("", "", func(key string, value interface{}) bool {
		if value == struct{}{} {
			addrs = append(addrs, std.Address(key))
		}
		return false
	})

	return addrs
}

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
		panic("access: caller is not authorized")
	}
}

// Add adds an address to the list of authorized addresses.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) Add(addr std.Address) {
	if s.HasAccess(addr) {
		panic("access: addr already has access")
	}

	s.authorized.Set(addr.String(), struct{}{})
}

// Del removes an address from the list of authorized addresses.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) Del(addr std.Address) error {
	if s.authorized.Size() == 0 {
		return ErrNoMembersInSet
	}

	if s.authorized.Size() == 1 {
		return ErrLastMember
	}

	return s.ForceDel(addr)
}

// ForceDel removes an address from the list of authorized addresses.
// It will not panic if the last member of the set is removed
// WARN: In this case, the ownership of the set is fully removed.
func (s *Set) ForceDel(addr std.Address) error {
	if _, removed := s.authorized.Remove(addr.String()); !removed {
		return ErrNotInSet
	}

	return nil
}

// ReplaceAll removes all existing authorized addresses and replaces them with a new one.
// It requires the caller or prevRealm to be authorized, otherwise, it panics.
func (s *Set) Replace(addrs []std.Address) {
	s.addrs = addrs
}
