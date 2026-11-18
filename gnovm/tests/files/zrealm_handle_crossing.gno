// PKGPATH: gno.land/r/zrealm_handle_crossing
package zrealm_handle_crossing

// A crossing function taken as a value still crosses when called with cur.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	run := cbvictim.RunS
	run(cross(cur), cbvictim.GateDecl(cross(cur)))
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: picked
