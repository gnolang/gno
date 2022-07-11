package acl

import "std"

func HasPerm(addr std.Address, verb string, resource string) bool {
	return false
}

func HasRole(addr std.Address, role string) bool {
	return HasPerm(addr, "role", role)
}
