// PKGPATH: gno.land/r/demo/curmethbs
package curmethbs

// The same check must run for a crossing METHOD. It did not: doOpCall asked
// ft.IsCrossing(), and ft there is the UNBOUND type, whose Params[0] is the
// receiver rather than `cur` — so the answer was false for every method and
// the whole block was skipped. No identity was installed and no check ran.
// It asks fv.IsCrossing() now, which preprocess sets from the declared type.

type S struct{}

func (s *S) target(cur realm) {
	println("target:", cur.PkgPath())
}

func (s *S) exec(cur realm, cb func()) {
	cb()
}

func main(cur realm) {
	s := &S{}
	cb := func() { s.target(cur) }
	s.exec(cross(cur), cb)
}

// Error:
// crossing function gno.land/r/demo/curmethbs.target was entered without cross(), so it takes its caller's identity — but the value passed is not the caller's own cur (a stale capture from another frame, a sibling frame, or a rebound cur). Pass the caller's own `cur` unchanged, or use cross(cur) to enter as this realm
