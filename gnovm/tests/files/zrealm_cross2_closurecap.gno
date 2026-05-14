// PKGPATH: gno.land/r/demo/cross2closurecap
package cross2closurecap

// cross2(rlm) works correctly when rlm was captured by a closure.
// The previous design (stashing the rlm path on the outer CallExpr
// and resolving it via Block.GetPointerTo at install time) crashed
// here with a Go nil-deref because closure-captured names live via
// heap-item indirection, not the parent-block chain. The redesign
// routes rlm through the normal eval machinery — cross2 is a real
// function call whose argument is evaluated normally, picking up
// the captured value, then validated via IsCurrent-strict in cross2's
// own body before being returned.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	// Capture a realm-typed local var in a closure invoked from
	// main's own frame. The captured value IS still the topmost
	// crossing frame's Cur (HIV preserved via shallow copy of
	// PointerValue.Base), so IsCurrent-strict passes. (If the
	// closure were invoked from a sibling crossing frame, IsCurrent
	// would fail — see stalerlm.)
	rlm := cur
	fn := func() string {
		return callee(cross2(rlm))
	}
	println(fn())
}

// Output:
// gno.land/r/demo/cross2closurecap
