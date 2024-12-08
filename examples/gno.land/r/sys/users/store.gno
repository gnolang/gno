package users

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	p "gno.land/p/demo/pausable"
	susers "gno.land/p/sys/users"
)

var (
	uStore             = avl.NewTree() // username > *susers.User
	aStore             = avl.NewTree() // address > *susers.User
	callerWhitelist    = avl.NewTree() // addr > struct{}{}
	preRestrictedNames = []string{
		"bitcoin", "cosmos", "newtendermint", "ethereum",
	}
)

// WriteUser writes to the sys/users store
// Only whitelisted callers can call this function
// Writing to the store can be paused by the admin
func WriteUser(user susers.User) error {
	if pausable.IsPaused() {
		return p.ErrPaused
	}

	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrUnauthorizedWrite
	}

	if err := verifyUser(user); err != nil {
		return err
	}

	if uStore.Has(user.Name) {
		return ErrAlreadyExists
	}

	uStore.Set(user.Name, &user)
	aStore.Set(user.Address.String(), &user)

	return nil
}

// XXX: should we have a DeleteUser? Maybe with a separate whitelist?

// IsOnWhitelist checks if the given address has
// permission to write to the user store
func IsOnWhitelist(addr std.Address) bool {
	return callerWhitelist.Has(addr.String())
}

// UserCount returns the number of registered users
func UserCount() int {
	return aStore.Size()
}

// XXX: Should we have this? it's not modifiable directly.
func GetUserStore() *avl.Tree {
	return uStore
}

func verifyUser(user susers.User) error {
	if strings.TrimSpace(user.Name) == "" {
		return ErrEmptyUsername
	}

	if !user.Address.IsValid() {
		return ErrInvalidAddress
	}

	return nil
}
