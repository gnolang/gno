// PKGPATH: gno.land/r/laundero
package laundero

// A /p/ closure built at /p/ init acts for no realm: called here it runs
// view-only, so its write to the victim's object is refused. Guard against a
// regression to a nil realm, under which the write would silently land.

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
// cannot modify object owned by another realm from library code (call a function of the owning realm): o<VPBlock(1,0)>.Field
