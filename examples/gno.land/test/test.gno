package test

import "std"

func ValidAddress(addr std.Address) string {
	if addr.Valid() {
		return "address is valid"
	} else {
		return "address is not valid"
	}
}
