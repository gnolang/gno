// PKGPATH: gno.land/r/zrealm_entry_pfunc_samepkg
package zrealm_entry_pfunc_samepkg

// The library's own exported writer, passed in by the attacker and called by
// the library through a value, is still a value: it runs view-only.

import (
	"gno.land/p/demo/tests/pfuncbind"
	"gno.land/r/tests/vm/pfuncvictim"
)

func main(cur realm) {
	pfuncvictim.Hof(cross(cur), pfuncbind.Clobber)
	println("victim:", pfuncvictim.Read())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): x<VPBlock(1,0)>.Field
