// PKGPATH: gno.land/r/zrealm_maker_ownclosure
package zrealm_maker_ownclosure

// A closure a realm made in a gate is a key to its own callback-running
// gates: whoever reads it may hand it back, and it writes for its maker.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.Init(cross(cur))
	cbvictim.RunS(cross(cur), cbvictim.Keep)
	println("owner:", cbvictim.Owner)
}

// Output:
// owner: picked
