package users

import (
	"std"

	"gno.land/r/demo/users"
)

// For now this package is disable by default
var enable = false

var admin = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"

// Admin call to Activate or Desactivate this package.
func AdminEnable() {
	assertIsAdmin()

	enable = true
}

func AdminDisable() {
	assertIsAdmin()

	enable = false
}

func IsEnable() bool { return enable }

// This method ensure that the given address have ownership on the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
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

func assertIsAdmin() {
	caller := std.GetOrigCaller()
	if admin != caller.String() {
		panic("access restricted")
	}
}
