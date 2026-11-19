// PKGPATH: gno.land/r/zrealm_home_recv
package zrealm_home_recv

// A function-typed receiver is handed over like an argument: cbvictim.H
// (maker cbvictim) runs read-only in its own crossing method.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.H.Run(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
