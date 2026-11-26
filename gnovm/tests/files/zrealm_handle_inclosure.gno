// PKGPATH: gno.land/r/zrealm_handle_inclosure
package zrealm_handle_inclosure

// The victim's own declared function, taken as a value inside a closure in
// read-only code, has no maker: the victim's gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	println("closure:", revive(func() { cbvictim.RunS(cross(cur), cbvictim.Get2()) }))
	println("closure var:", revive(func() { cbvictim.RunS(cross(cur), cbvictim.Get3()) }))
	println("owner:", cbvictim.Owner)
}

// Output:
// closure: gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
// closure var: gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
// owner: victim
