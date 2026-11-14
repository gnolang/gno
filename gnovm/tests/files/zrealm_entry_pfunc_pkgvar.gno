// PKGPATH: gno.land/r/zrealm_entry_pfunc_pkgvar
package zrealm_entry_pfunc_pkgvar

// A package var holding a /p/ function is a value, not a name in source.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.SetHook(launderattack.EvilWrite)
	pfuncvictim.RunHook()
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): i<VPBlock(1,0)>.Field
