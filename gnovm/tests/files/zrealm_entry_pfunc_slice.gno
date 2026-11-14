// PKGPATH: gno.land/r/zrealm_entry_pfunc_slice
package zrealm_entry_pfunc_slice

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.ApplySlice([]pfuncvictim.Fn{launderattack.EvilWrite})
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
