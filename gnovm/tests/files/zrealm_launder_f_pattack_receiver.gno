// PKGPATH: gno.land/r/launderf
package launderf

// Explicit-entry rule: a /p/ method called from here has no power over
// another realm's object, so the write below is refused. The analysis below
// was written before the explicit-entry rule.

// Attack F: attacker invokes /p/attack.Tamperer.TamperMethod(p, s).
// The Tamperer{} literal is constructed in attacker's realm (launderf),
// so the receiver is launderf-stamped. PushFrameCall's borrow rule
// finds recvOID.PkgID = launderf == m.Realm.ID = launderf — no borrow.
// m.Realm stays launderf for TamperMethod's body. The write to
// o.Field with base.PkgID=laundervictim → readonly panic.
//
// The receiver-borrow exploit only works when the receiver has the
// victim's stamp, which requires the receiver to be aliased into
// victim's state (not a fresh attacker literal).

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	p := laundervictim.GetPtr()
	launderattack.Tamperer{}.TamperMethod(p, "pwnd")
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): o<VPBlock(1,1)>.Field
