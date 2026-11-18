// PKGPATH: gno.land/r/launderf
package launderf

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

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
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): o<VPBlock(1,1)>.Field
