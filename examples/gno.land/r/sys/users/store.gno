package users

import (
	"std"

	"gno.land/p/demo/avl"
	p "gno.land/p/demo/pausable"
	pusers "gno.land/p/demo/users"
)

var (
	uStore             = avl.NewTree() // username > *pusers.User
	aStore             = avl.NewTree() // address > *pusers.User
	callerWhitelist    = avl.NewTree() // addr > struct{}{}
	preRestrictedNames = []string{
		"bitcoin", "cosmos", "newtendermint", "ethereum",
	}
)

// WriteUser writes to the sys/users stores
// Only whitelisted callers, such as r/gnoland/users, can call this function
// Writing to the store can be paused by the admin
func WriteUser(user *pusers.User) error {
	if pausable.IsPaused() {
		return p.ErrPaused
	}

	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrUnauthorizedWrite
	}

	username := user.Name
	addr := user.Address

	if uStore.Has(username) {
		return ErrAlreadyExists
	}

	uStore.Set(username, user)
	aStore.Set(addr.String(), user)

	return nil
}

// IsOnWhitelist checks if the given address has
// permission to write to the user store
func IsOnWhitelist(addr std.Address) bool {
	return callerWhitelist.Has(addr.String())
}

// UserCount returns the number of registered users
func UserCount() int {
	return aStore.Size()
}
