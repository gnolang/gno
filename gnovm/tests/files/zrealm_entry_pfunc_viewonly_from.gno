// PKGPATH: gno.land/r/zrealm_entry_pfunc_viewonly_from
package zrealm_entry_pfunc_viewonly_from

// A view-only closure calls the victim's ApplyNC with EvilWrite; the victim's
// code, entered without crossing, runs read-only and writes nothing.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/p/demo/tests/pfuncbind"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	cl := pfuncbind.Bind(pfuncvictim.ApplyNC, launderattack.EvilWrite)
	pfuncvictim.Call(cross(cur), cl)
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): i<VPBlock(1,0)>.Field
