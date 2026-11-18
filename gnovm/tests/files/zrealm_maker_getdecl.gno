// PKGPATH: gno.land/r/zrealm_maker_getdecl
package zrealm_maker_getdecl

// KNOWN: a realm that returns or exports its own declared function hands out a key to its callback-running gates.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	println("owner:", cbvictim.Owner)
	cbvictim.RunS(cross(cur), cbvictim.GetDecl())
	println("owner:", cbvictim.Owner)
	cbvictim.RunS(cross(cur), cbvictim.FV)
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: victim
// owner: picked
// owner: picked
