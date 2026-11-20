// PKGPATH: gno.land/r/zrealm_home_exportvar
package zrealm_home_exportvar

// cbvictim.FV is cbvictim's value; handed back to its gate by another realm
// it has no maker, so the gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.FV)
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only in the crossing call that made it: Owner<~VPBlock(3,0)>
