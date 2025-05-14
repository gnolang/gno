// Package authz provides simple and flexible authorization control for
// privileged actions.
package authz

import (
	"std"
)

// Authority is the core interface for authorization checks.
// It provides a single method to check if an address is authorized for an action.
type Authority interface {
	// Can checks if the given address is authorized to perform the action.
	Can(addr std.Address, action string) bool
}

// CanByCurrent checks if the current realm is authorized to perform an action.
func CanByCurrent(auth Authority, action string) bool {
	return auth.Can(std.CurrentRealm().Address(), action)
}

// CanByPrevious checks if the previous realm is authorized to perform an action.
func CanByPrevious(auth Authority, action string) bool {
	return auth.Can(std.PreviousRealm().Address(), action)
}

// AssertCan panics if the caller is not authorized to perform an action.
func AssertCan(auth Authority, addr std.Address, action string) bool {
	if !auth.Can(addr, action) {
		panic("unauthorized: " + addr.String() + " cannot " + action)
	}
	return true
}

// AssertCanByCurrent panics if the current realm is not authorized to perform an action.
func AssertCanByCurrent(auth Authority, action string) bool {
	return AssertCan(auth, std.CurrentRealm().Address(), action)
}

// AssertCanByPrevious panics if the previous realm is not authorized to perform an action.
func AssertCanByPrevious(auth Authority, action string) bool {
	return AssertCan(auth, std.PreviousRealm().Address(), action)
}

// Library-level helper functions that work with any Authority implementation

// CanByCurrent checks if the current realm is authorized to perform an action.
func CanByCurrent(auth Authority, action string) bool {
	return auth.Can(std.CurrentRealm().Address(), action)
}

// CanByPrevious checks if the previous realm is authorized to perform an action.
func CanByPrevious(auth Authority, action string) bool {
	return auth.Can(std.PreviousRealm().Address(), action)
}

// AssertCan panics if the caller is not authorized to perform an action.
func AssertCan(auth Authority, addr std.Address, action string) bool {
	if !auth.Can(addr, action) {
		panic("unauthorized")
	}
	return true
}

// AssertCanByCurrent panics if the current realm is not authorized to perform an action.
func AssertCanByCurrent(auth Authority, action string) bool {
	return AssertCan(auth, std.CurrentRealm().Address(), action)
}

// AssertCanByPrevious panics if the previous realm is not authorized to perform an action.
func AssertCanByPrevious(auth Authority, action string) bool {
	return AssertCan(auth, std.PreviousRealm().Address(), action)
}
