package users

import (
	"std"

	"gno.land/r/demo/users"
)

// For now this package is disable by default
var enable = false

var admin = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul

// Admin call to Activate or Desactivate this package.
func AdminActivate(t bool) {
	assertAdmin()

	enable = t
}

// This method ensure that the given address have ownership on the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	if !enable {
		return true
	}

	if user := users.GetUserByName(name); user != nil {
		return user.Address == address
	}

	return false
}

func assertIsAdmin() {
	caller := std.GetOrigCaller()
	if admin != caller {
		panic("access restricted.")
	}
}
