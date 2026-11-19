// PKGPATH: gno.land/r/zrealm_home_third
package zrealm_home_third

// A third realm relays cbvictim's value back to cbvictim's gate: its hold
// is not the maker's.

import (
	"gno.land/r/tests/vm/cbother"
	"gno.land/r/tests/vm/cbvictim"
)

func main(cur realm) {
	cbother.Relay(cross(cur))
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): Owner<~VPBlock(3,0)>
