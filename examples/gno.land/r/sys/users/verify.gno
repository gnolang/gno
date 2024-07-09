package users

import (
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/r/demo/users"
)

const admin = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @jae

type VerifyNameFunc func(enabled bool, address std.Address, name string) bool

var (
	owner     = ownable.NewWithAddress(admin) // Pkg owner
	checkFunc = DefaultCheckName              // Checking namesapce callback
	enabled   = false                         // For now this package is disable by default
)

func IsEnable() bool { return enabled }

// This method ensure that the given address have ownership on the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return checkFunc(enabled, address, name)
}

func DefaultCheckName(enable bool, address std.Address, name string) bool {
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
