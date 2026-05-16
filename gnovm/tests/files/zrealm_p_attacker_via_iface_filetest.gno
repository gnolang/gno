// PKGPATH: gno.land/r/caller
// MAXALLOC: 100000000
package caller

import (
	"chain/runtime"

	"gno.land/r/tests/vm/crossrealm"
)

// PLAN3 Phase 3: /p/-method dispatched on an unreal foreign receiver
// borrows m.Realm to the receiver's PkgID (Layer 2 expansion).
// Pre-PLAN3 the borrow only fired for REAL foreign receivers; under
// Phase 2's allocation-time PkgID stamping, unreal receivers also
// carry their authoring realm's PkgID and Layer 2 follows.
//
// crossrealm.Make1() returns a *p_crossrealm.Container constructed
// inside /r/tests/vm/crossrealm (Layer 1 borrow). The Container is
// returned to /r/caller still bearing PkgID=crossrealm. When /r/caller
// invokes the /p/-declared Touch() method, Layer 2 fires:
// runtime.CurrentRealm() inside Touch reports /r/tests/vm/crossrealm,
// not /r/caller.

func main(cur realm) {
	c := crossrealm.Make1() // returns *p_crossrealm.Container, PkgID=crossrealm
	c.Touch()               // /p/-method; Layer 2 borrows m.Realm to crossrealm
	println("after Touch A=", c.A)
	println("caller-side CurrentRealm:", runtime.CurrentRealm().PkgPath())
}

// Output:
// after Touch A= 2
// caller-side CurrentRealm: gno.land/r/caller
