// PKGPATH: gno.land/r/zrealm_home_provider
package zrealm_home_provider

// The gate calls this realm's provider, which returns cbvictim's value: the
// returning code's hold (this realm) strips it.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunGet(cross(cur), func() func(string) { return cbvictim.FV })
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): Owner<~VPBlock(3,0)>
