// Package authz provides simple and flexible authorization control for
// privileged actions.
package authz

import (
	"std"
)

// Authority is the core interface for authorization checks.
// It provides a single method to retrieve the membership of an action.
//
// Example: if auth.MembersFor("propose").Has(caller)
type Authority struct {
	memberships map[string]Membership
}

// Can checks if the given address is authorized to perform the action.
func (a *Authority) Can(addr std.Address, action string) bool {
	return a.memberships[action].Has(addr)
}

// CanByCurrent checks if the current realm is authorized to perform an action.
//
// Example: if authz.CanByCurrent(auth, "propose") {}
func (a *Authority) CanByCurrent(action string) bool {
	addr := std.CurrentRealm().Address()
	return a.Can(addr, action)
}

// CanByPrevious checks if the previous realm is authorized to perform an
// action.
//
// Example: if authz.CanByPrevious(auth, "propose") {}
func (a *Authority) CanByPrevious(action string) bool {
	addr := std.PreviousRealm().Address()
	return a.Can(addr, action)
}

// AssertCan panics if the caller is not authorized to perform an action.
//
// Example: authz.AssertCan(auth, caller, "vote")
func (a *Authority) AssertCan(addr std.Address, action string) {
	if !a.memberships[action].Has(addr) {
		panic("unauthorized: " + addr.String() + " cannot " + action)
	}
}

// AssertCanByCurrent panics if the current realm is not authorized to perform
// an action.
//
// Example: authz.AssertCanByCurrent(auth, "propose")
func (a *Authority) AssertCanByCurrent(action string) {
	addr := std.CurrentRealm().Address()
	a.AssertCan(addr, action)
}

// AssertCanByPrevious panics if the previous realm is not authorized to perform
// an action.
//
// Example: authz.AssertCanByPrevious(auth, "propose")
func (a *Authority) AssertCanByPrevious(action string) {
	addr := std.PreviousRealm().Address()
	a.AssertCan(addr, action)
}
