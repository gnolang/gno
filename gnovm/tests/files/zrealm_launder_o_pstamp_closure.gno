// PKGPATH: gno.land/r/laundero
package laundero

// UPDATED (view-by-default): a /p/-method write on an object another realm
// owns now borrows only if the owner granted mutable(x). None of these
// fixtures grant, so the write described below is refused; the expected
// output pins the refusal. The original analysis is kept for context.

// Attack O: borrow rule #3 (closure) laundering canary.
//
// launderattack.StoredHook is a /p/-package-level closure (a FuncLit
// evaluated at /p/launderattack init, so it's stamped /p/launderattack).
// Invoking it triggers borrow rule #3, which borrows m.Realm to the
// closure's construction realm (/p/launderattack) — NOT the caller, and NOT
// nil. The hook's write to the victim's /r/laundervictim-stamped Object is
// then rejected by the readonly gate (/r/laundervictim != /p/launderattack).
//
// Regression guard: if rule #3 ever borrowed to a nil realm (e.g. the /p/
// frozen realm regressed), m.Realm would be nil, IsReadonly would
// short-circuit to "allow", and this write would silently succeed.

import (
	"gno.land/p/demo/tests/launderattack/v0"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	p := laundervictim.GetPtr() // *launderpkg.Object, /r/laundervictim-stamped
	launderattack.StoredHook(p)
	println("after:", laundervictim.ReadPtr())
}

// Error:
// cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): o<VPBlock(1,0)>.Field
