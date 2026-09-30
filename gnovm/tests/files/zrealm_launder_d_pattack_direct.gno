// PKGPATH: gno.land/r/launderd
package launderd

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

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
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): o<VPBlock(1,0)>.Field
