// PKGPATH: gno.land/r/zrealm_entry_pfunc_var
package zrealm_entry_pfunc_var

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.ApplyVar(cross(cur), launderattack.EvilWrite)
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): i<VPBlock(1,0)>.Field
