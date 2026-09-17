// PKGPATH: gno.land/r/launderm
package launderm

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Attack M: nil-realm hole via a /p/-STAMPED receiver.
//
// launderpkg.PInitData is allocated during /p/launderpkg init, so its
// StructValue is /p/-stamped. Dispatching a method on it borrows m.Realm
// to /p/launderpkg's realm. UseMutator then dispatches EvilMutator.Run
// (a no-anchor int, which keeps the borrowed realm), and Run writes to
// the victim's /r/-stamped Immutable.
//
// Before Approach C (/p/ packages getting their own realm), the /p/
// borrow set m.Realm = nil, and IsReadonly short-circuited to "nothing
// is readonly" — so the write to victim data succeeded. Now the /p/
// borrow lands on /p/launderpkg's (immutable) realm, the PkgID-mismatch
// readonly gate fires, and the cross-realm write is rejected.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/p/demo/tests/launderpkg/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	victim := laundervictim.GetImm() // *launderpkg.Immutable, laundervictim-stamped
	launderpkg.PInitData.UseMutator(victim, launderattack.EvilMutator(0))
	println("after:", laundervictim.ReadImm())
}

// Error:
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): i<VPBlock(1,1)>.Field
