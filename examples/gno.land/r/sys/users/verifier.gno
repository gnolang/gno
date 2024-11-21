package users

import (
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

type VerifyNameFunc func(enabled bool, address std.Address, name string) bool

var (
	checkFunc = VerifyNameByUser // Checking namespace callback
)

// IsAuthorizedAddressForName ensures that the given address has ownership of the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return checkFunc(pausable.IsPaused(), address, name)
}

// VerifyNameByUser checks from the `r/gnoland/users` package that the user has correctly
// registered the given name.
// This function considers as valid an `address` that matches the `name`.
func VerifyNameByUser(enable bool, address std.Address, name string) bool {
	if !enable {
		return true
	}

	// Allow user with their own address as name
	if address.String() == name {
		return true
	}

	if user := ResolveUsername(name); user != nil {
		return user.Address == address
	}

	return false
}
