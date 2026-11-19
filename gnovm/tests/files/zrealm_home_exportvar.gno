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
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
