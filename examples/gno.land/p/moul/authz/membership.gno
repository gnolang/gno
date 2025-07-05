package authz

import "std"

// XXX: consider
// - TimebasedMembership
// - CoinbasedMembership

// NewTokenHolderMembership returns a membership that matches addresses that
// hold at least the required amount of tokens.
func NewTokenHolderMembership(requiredCoins std.Coins) Membership {
	return &tokenHolderMembership{required: requiredCoins}
}

type tokenHolderMembership struct {
	required std.Coins
}

func (m *tokenHolderMembership) Has(addr std.Address) bool {
	// TODO: This is a placeholder implementation.
	// In a real scenario, this would check the token balance of the address
	// against the required amount. For now, we'll return false.
	return false
}

// NewCustomMembership is an alias for NewFuncMembership for backward compatibility.
func NewCustomMembership(f func(addr std.Address) bool) Membership {
	return NewFuncMembership(f)
}

// Membership is the interface to check if an address is a member of a set.
type Membership interface {
	Has(addr std.Address) bool
}

// nilMembership is a zero/noop membership implementation that is always empty.
type nilMembership struct{}

func (m *nilMembership) Has(addr std.Address) bool { return false }

// NewFuncMembership returns a membership that matches any address that
// satisfies the given function.
func NewFuncMembership(f func(addr std.Address) bool) Membership {
	return &funcMembership{f: f}
}

type funcMembership struct {
	f func(addr std.Address) bool
}

func (m *funcMembership) Has(addr std.Address) bool {
	return m.f(addr)
}

// Special Memberships

// Anyone returns a membership that matches any address.
func Anyone() Membership { return &anyoneMembership{} }

type anyoneMembership struct{}

func (m *anyoneMembership) Has(addr std.Address) bool {
	return true
}

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

func (and *andMembership) Has(addr std.Address) bool {
	for _, sub := range and.subs {
		if !sub.Has(addr) {
			return false
		}
	}
	return true
}

func (or *orMembership) Has(addr std.Address) bool {
	for _, sub := range or.subs {
		if sub.Has(addr) {
			return true
		}
	}
	return false
}

func (not *notMembership) Has(addr std.Address) bool {
	return !not.orig.Has(addr)
}

func (m *xorMembership) Has(addr std.Address) bool {
	a := m.a.Has(addr)
	b := m.b.Has(addr)
	return a != b
}

/* XXX: consider PBAC.MembershipFor()
// MembershipFromPBAC returns a membership that is the same as the membership
// for the given action in the given PBAC.
func MembershipFromPBAC(pbac *PBAC, action string) Membership {
	return &readonlyMembershipFromPBAC{pbac: pbac, action: action}
}

type readonlyMembershipFromPBAC struct {
	pbac   *PBAC
	action string
}

func (romw *readonlyMembershipFromPBAC) Has(addr std.Address) bool {
	return romw.pbac.Can(addr, romw.action)
}
*/
