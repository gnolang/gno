package users

import (
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/r/demo/users"
)

const admin = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul

type VerifyNameFunc func(enabled bool, address std.Address, name string) bool

var (
	owner     = ownable.NewWithAddress(admin) // Package owner
	checkFunc = VerifyNameByUser              // Checking namespace callback
	enabled   = false                         // For now this package is disabled by default
)

func IsEnabled() bool { return enabled }

// This method ensures that the given address has ownership of the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return checkFunc(enabled, address, name)
}

// VerifyNameByUser checks from the `users` package that the user has correctly
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

	if user := users.GetUserByName(name); user != nil {
		return user.Address == address
	}

	return false
}

// Admin calls

// Enable this package.
func AdminEnable() {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	enabled = true
}

// Disable this package.
func AdminDisable() {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	enabled = false
}

// AdminUpdateVerifyCall updates the method that verifies the namespace.
func AdminUpdateVerifyCall(check VerifyNameFunc) {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	checkFunc = check
}

// AdminTransferOwnership transfers the ownership to a new owner.
func AdminTransferOwnership(newOwner std.Address) error {
	if err := owner.CallerIsOwner(); err != nil {
		panic(err)
	}

	return owner.TransferOwnership(newOwner)
}
