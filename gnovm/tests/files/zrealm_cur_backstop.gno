// PKGPATH: gno.land/r/demo/curbackstop
package curbackstop

// THIS TEST EXISTS TO MAKE THE RUNTIME CHECK DELETABLE-DETECTABLE.
//
// doOpCall's check — that a crossing function entered without cross() takes
// its CALLER's own identity and not some other realm value — is the one layer
// that does not depend on having enumerated every syntactic route into the
// `cur` slot. That enumeration has been incomplete twice (`&cur`, then
// `for _, cur = range`). Without this test the whole suite passes with the
// check deleted, so a dead-code pass removes it and CI says nothing.
//
// THE SHAPE IT PINS IS AN ORDINARY MISTAKE. A closure captures `cur` in one
// crossing frame and is called from another; inside the closure `cur` still
// names the first frame's realm, while the frame actually running is the
// second's. That is a stale capture, and the error says so. The cross()
// spelling of the same mistake is already refused — see
// zrealm_cross_closuresib.gno — so this makes the two paths agree.
//
// An earlier version of this test reached the check through the type-alias gap
// in the `cur` naming rules. That gap is closed now and this shape replaced it.
// If this shape is ever made legal, find another route — do not delete the
// coverage.

func target(cur realm) {
	println("target:", cur.PkgPath())
}

func exec(cur realm, cb func()) {
	cb()
}

func main(cur realm) {
	cb := func() { target(cur) } // captures main's cur
	exec(cross(cur), cb)         // ...but runs inside exec's frame
}

// Error:
// crossing function gno.land/r/demo/curbackstop.target was entered without cross(), so it takes its caller's identity — but the value passed is not the caller's own cur (a stale capture from another frame, a sibling frame, or a rebound cur). Pass the caller's own `cur` unchanged, or use cross(cur) to enter as this realm
