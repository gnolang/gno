// PKGPATH: gno.land/r/zrealm_maker_ownclosure
package zrealm_maker_ownclosure

// A closure a realm made in a gate is no key to its callback-running gates:
// handed back by another realm it has no maker and runs read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.Init(cross(cur))
	cbvictim.RunS(cross(cur), cbvictim.Keep)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/cbvictim) package{})).Owner
