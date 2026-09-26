// PKGPATH: gno.land/r/caller
// MAXALLOC: 100000000
package caller

import (
	"chain/runtime/unsafe"

	"gno.land/r/tests/vm/crossrealm"
)

// /p/-method dispatched on a foreign receiver the owner granted with
// mutable(x) borrows m.Realm to the receiver's PkgID (borrow rule #2),
// whether the receiver is real or, as here, allocated in this transaction.
//
// crossrealm.Make1() returns a *p_crossrealm.Container constructed and
// granted inside /r/tests/vm/crossrealm (borrow rule #1). The Container is
// returned to /r/caller still bearing PkgID=crossrealm. When /r/caller
// invokes the /p/-declared Touch() method, borrow rule #2 fires:
// unsafe.CurrentRealm() inside Touch reports /r/tests/vm/crossrealm,
// not /r/caller.

func main(cur realm) {
	c := crossrealm.Make1() // returns *p_crossrealm.Container, PkgID=crossrealm
	c.Touch()               // /p/-method; borrow rule #2 borrows m.Realm to crossrealm
	println("after Touch A=", c.A)
	println("caller-side CurrentRealm:", unsafe.CurrentRealm().PkgPath())
}

// Output:
// after Touch A= 2
// caller-side CurrentRealm: gno.land/r/caller
