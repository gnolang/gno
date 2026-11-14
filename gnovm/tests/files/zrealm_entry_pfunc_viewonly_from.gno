// PKGPATH: gno.land/r/zrealm_entry_pfunc_viewonly_from
package zrealm_entry_pfunc_viewonly_from

// A view-only closure calls the victim's ApplyOne with EvilWrite; the victim
// then calls that value through a variable, so it runs view-only.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/p/demo/tests/pfuncbind"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	cl := pfuncbind.Bind(pfuncvictim.ApplyOne, launderattack.EvilWrite)
	pfuncvictim.Call(cl)
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
