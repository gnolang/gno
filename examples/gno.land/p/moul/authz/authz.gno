// Package authz provides simple and flexible authorization control for
// privileged actions.
package authz

import (
	"std"
)

/*
func Example() {
	authority, priv := New()
	ok := Can(authority, caller, "propose")
	AssertCan(authority, caller, "propose")
	priv.Add("vote", "g1user1", "g1user2")
	priv.Add(authz.Admin, "g1admin")
	priv.Add("propose", authz.Anyone)
	priv.Add("propose")
	// combine
	// tokenMembers
	// coinMembers
	// andXXX
	// timeBased
	// and(thisCallbackReturnTrue)
	// anyCallbackReturnsTrue
}
*/

// Authority is the core interface for authorization checks.
// It provides a single method to retrieve the membership of an action.
//
// Example: if auth.MembersFor("propose").Has(caller)
type Authority interface {
	MembersFor(action string) Membership
}

// Can checks if the given address is authorized to perform the action.
func Can(auth Authority, addr std.Address, action string) bool {
	return auth.MembersFor(action).Has(addr)
}

// CanByCurrent checks if the current realm is authorized to perform an action.
//
// Example: if authz.CanByCurrent(auth, "propose") {}
func CanByCurrent(auth Authority, action string) bool {
	addr := std.CurrentRealm().Address()
	return Can(auth, addr, action)
}

// CanByPrevious checks if the previous realm is authorized to perform an
// action.
//
// Example: if authz.CanByPrevious(auth, "propose") {}
func CanByPrevious(auth Authority, action string) bool {
	addr := std.PreviousRealm().Address()
	return Can(auth, addr, action)
}

// AssertCan panics if the caller is not authorized to perform an action.
//
// Example: authz.AssertCan(auth, caller, "vote")
func AssertCan(auth Authority, addr std.Address, action string) {
	if !auth.MembersFor(action).Has(addr) {
		panic("unauthorized: " + addr.String() + " cannot " + action)
	}
}

// AssertCanByCurrent panics if the current realm is not authorized to perform
// an action.
//
// Example: authz.AssertCanByCurrent(auth, "propose")
func AssertCanByCurrent(auth Authority, action string) {
	addr := std.CurrentRealm().Address()
	AssertCan(auth, addr, action)
}

// AssertCanByPrevious panics if the previous realm is not authorized to perform
// an action.
//
// Example: authz.AssertCanByPrevious(auth, "propose")
func AssertCanByPrevious(auth Authority, action string) {
	addr := std.PreviousRealm().Address()
	AssertCan(auth, addr, action)
}
