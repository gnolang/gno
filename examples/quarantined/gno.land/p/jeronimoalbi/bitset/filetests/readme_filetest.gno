package main

import "gno.land/p/jeronimoalbi/bitset"

const (
	PermRead   = 0
	PermWrite  = 1
	PermDelete = 2
	PermAdmin  = 3
)

func main() {
	var perms bitset.BitSet
	perms.Set(PermRead)
	perms.Set(PermWrite)

	println("Can read:", perms.IsSet(PermRead))
	println("Can write:", perms.IsSet(PermWrite))
	println("Can delete:", perms.IsSet(PermDelete))
	println("Is admin:", perms.IsSet(PermAdmin))
	println("Permissions set:", perms.Len())
}

// Output:
// Can read: true
// Can write: true
// Can delete: false
// Is admin: false
// Permissions set: 2
