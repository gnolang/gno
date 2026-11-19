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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
