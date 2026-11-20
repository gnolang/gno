// PKGPATH: gno.land/r/caller
// MAXALLOC: 100000000
package caller

import (
	"chain/runtime/unsafe"

	"gno.land/r/tests/vm/crossrealm"
)

// crossrealm.Make1() returns a *p_crossrealm.Container built inside
// /r/tests/vm/crossrealm and carrying PkgID=crossrealm. A foreign caller's
// /p/-declared Touch() is refused and the owner's state is unchanged.

func main(cur realm) {
	c := crossrealm.Make1() // returns *p_crossrealm.Container, PkgID=crossrealm
	func() {
		defer func() { println("caller Touch:", recover()) }()
		c.Touch()
	}()
	println("after Touch A=", c.A)
	println("caller-side CurrentRealm:", unsafe.CurrentRealm().PkgPath())
}

// Output:
// caller Touch: cannot modify object owned by another realm from library code (call a function of the owning realm): c<VPBlock(1,0)>.A
// after Touch A= 1
// caller-side CurrentRealm: gno.land/r/caller
