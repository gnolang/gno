// PKGPATH: gno.land/r/laundero
package laundero

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
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/laundervictim"
)

func main(cur realm) {
	p := laundervictim.GetPtr() // *launderpkg.Object, /r/laundervictim-stamped
	launderattack.StoredHook(p)
	println("after:", laundervictim.ReadPtr())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): o<VPBlock(1,0)>.Field
