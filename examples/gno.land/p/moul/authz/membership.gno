package authz

import "std"

// XXX: consider
// - TimebasedMembership
// - CoinbasedMembership

// Membership is the interface for a set of addresses (members).
type Membership interface {
	Has(addr std.Address) bool
}

// Anyone returns a membership that matches any address.
func Anyone() Membership { return &AnyoneMembership{} }

type AnyoneMembership struct{}

func (m *AnyoneMembership) Has(addr std.Address) bool {
	return true
}

// BoolOps
//

type (
	AndMembership struct{ subs []Membership }
	OrMembership  struct{ subs []Membership }
	XorMembership struct{ a, b Membership }
	NotMembership struct{ orig Membership }
)

// And returns a new membership that is the intersection of the given
// memberships.
func And(subs ...Membership) Membership {
	return &AndMembership{subs: subs}
}

func (and AndMembership) Has(addr std.Address) bool {
	for _, sub := range and.subs {
		if !sub.Has(addr) {
			return false
		}
	}
	return true
}

// Or returns a new membership that is the union of the given memberships.
func Or(subs ...Membership) Membership {
	return &OrMembership{subs: subs}
}

func (or OrMembership) Has(addr std.Address) bool {
	for _, sub := range or.subs {
		if sub.Has(addr) {
			return true
		}
	}
	return false
}

// Not returns a new membership that is the complement of the given membership.
func Not(sub Membership) Membership {
	return &NotMembership{orig: sub}
}

func (not NotMembership) Has(addr std.Address) bool {
	return !not.Has(addr)
}

// Xor returns a new membership that is the symmetric difference of the given
// memberships.
func Xor(a, b Membership) Membership {
	return &XorMembership{a: a, b: b}
}

func (m XorMembership) Has(addr std.Address) bool {
	a := m.a.Has(addr)
	b := m.b.Has(addr)
	return a != b
}
