package authz

import "std"

// XXX: consider
// - TimebasedMembership
// - CoinbasedMembership

// Membership is the interface to check if an address is a member of a set.
type Membership interface {
	Has(addr std.Address) bool
}

// Nil membership always returns false.
func (m *Membership) Has(addr std.Address) bool { return false }

// CustomMembership returns a membership that matches any address that
// satisfies the given function.
func CustomMembership(f func(addr std.Address) bool) Membership {
	return &customMembership{f: f}
}

type customMembership struct {
	f func(addr std.Address) bool
}

func (m *customMembership) Has(addr std.Address) bool {
	return m.f(addr)
}

// Special Memberships

// Anyone returns a membership that matches any address.
func Anyone() Membership { return &anyoneMembership{} }

type anyoneMembership struct{}

func (m *anyoneMembership) Has(addr std.Address) bool {
	return true
}q

// Boolean operations
//

// And returns a new membership that is the intersection of the given
// memberships.
func And(subs ...Membership) Membership {
	return &andMembership{subs: subs}
}

// Or returns a new membership that is the union of the given memberships.
func Or(subs ...Membership) Membership {
	return &orMembership{subs: subs}
}

// Not returns a new membership that is the complement of the given membership.
func Not(sub Membership) Membership {
	return &notMembership{orig: sub}
}

// Xor returns a new membership that is the symmetric difference of the given
// memberships.
func Xor(a, b Membership) Membership {
	return &xorMembership{a: a, b: b}
}

type (
	andMembership struct{ subs []Membership }
	orMembership  struct{ subs []Membership }
	xorMembership struct{ a, b Membership }
	notMembership struct{ orig Membership }
)

func (and andMembership) Has(addr std.Address) bool {
	for _, sub := range and.subs {
		if !sub.Has(addr) {
			return false
		}
	}
	return true
}

func (or orMembership) Has(addr std.Address) bool {
	for _, sub := range or.subs {
		if sub.Has(addr) {
			return true
		}
	}
	return false
}

func (not notMembership) Has(addr std.Address) bool {
	return !not.orig.Has(addr)
}

func (m xorMembership) Has(addr std.Address) bool {
	a := m.a.Has(addr)
	b := m.b.Has(addr)
	return a != b
}
