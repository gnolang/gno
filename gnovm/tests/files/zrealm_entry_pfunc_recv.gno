// PKGPATH: gno.land/r/zrealm_entry_pfunc_recv
// A function-typed receiver is a value, not a named function: calling it from
// the method runs view-only.
package zrealm_entry_pfunc_recv

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.RunHookOn(launderattack.EvilWrite)
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
