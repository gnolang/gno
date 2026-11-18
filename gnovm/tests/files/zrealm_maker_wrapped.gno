// PKGPATH: gno.land/r/zrealm_maker_wrapped
package zrealm_maker_wrapped

// The caller wraps the victim's writer in its own closure: the closure is the
// caller's, and the victim's writer it calls without crossing runs read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), func(s string) { cbvictim.SetOwner(s) })
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
