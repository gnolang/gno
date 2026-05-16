// PKGPATH: gno.land/r/demo/cross2closuresib
package cross2closuresib

// Runtime IsCurrent-strict catches a closure-captured rlm invoked from
// a SIBLING crossing frame (different from the one that captured rlm).
// The closure captures rlm from main's frame, then is passed via a
// cross-call into runClosure. When runClosure invokes the closure,
// runClosure's frame is the topmost crossing frame, not main's —
// the captured rlm's HIV no longer matches the topmost frame's Cur.
// cross2(rlm) panics.

func callee(cur realm) string {
	return cur.PkgPath()
}

// runClosure is a crossing function that invokes the passed closure
// from inside its own crossing frame. Calling fn() here creates a
// non-crossing call frame for fn ABOVE runClosure's crossing frame —
// fn's captured rlm (main's cur) is stale relative to runClosure's
// topmost crossing frame.
func runClosure(cur realm, fn func() string) {
	println(fn())
}

func main(cur realm) {
	rlm := cur
	fn := func() string {
		return callee(cross2(rlm))
	}
	runClosure(cross2(cur), fn)
}

// Error:
// cross2: rlm is not the current cur (stale capture or sibling frame)
