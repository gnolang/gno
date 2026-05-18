// PKGPATH: gno.land/r/launderd
package launderd

// Attack D: /p/attack.TamperDirect writes through the pointer in its
// OWN body. The /p/ function has no receiver, so PushFrameCall does
// not borrow — m.Realm stays attacker. The write inside TamperDirect
// sees base.PkgID=laundervictim, m.Realm=launderd → readonly panic.
//
// CONCLUSION: /p/attack writing in its own body is closed by the
// readonly check, NOT by CLOSE_LAUNDER. The data's PkgID stamp stays
// laundervictim throughout (CLOSE_LAUNDER ensures no laundering on
// pointer-deref), and the borrow rule doesn't grant /p/attack victim
// authority for a non-method call.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	p := laundervictim.GetPtr()
	launderattack.TamperDirect(p, "pwnd")
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): o<VPBlock(1,0)>.Field
