// PKGPATH: gno.land/r/zrealm_entry_pfunc_echo
package zrealm_entry_pfunc_echo

// A /p/ function value the victim hands back is still a value: ApplyOne
// calls it through a variable, so it runs view-only.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	f := pfuncvictim.EchoSlice([]pfuncvictim.Fn{launderattack.EvilWrite})
	pfuncvictim.ApplyOne(cross(cur), f)
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): i<VPBlock(1,0)>.Field
