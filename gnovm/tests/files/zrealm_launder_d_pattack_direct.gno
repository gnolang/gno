// PKGPATH: gno.land/r/launderd
package launderd

// Explicit-entry rule: a /p/ function with no receiver has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// Attack D: /p/attack.TamperDirect writes through the pointer in its
// OWN body. The /p/ function has no receiver, so PushFrameCall does
// not borrow — m.Realm stays attacker. The write inside TamperDirect
// sees base.PkgID=laundervictim, m.Realm=launderd → readonly panic.
//
// CONCLUSION: /p/attack writing in its own body is closed by the
// readonly check. The data's PkgID stamp stays laundervictim
// throughout, and the borrow rule doesn't grant /p/attack victim
// authority for a non-method call.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	p := laundervictim.GetPtr()
	launderattack.TamperDirect(p, "pwnd")
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): o<VPBlock(1,0)>.Field
