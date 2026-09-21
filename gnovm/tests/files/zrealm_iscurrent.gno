// PKGPATH: gno.land/r/test/iscurrent
package iscurrent

// Regression test for the `realm.IsCurrent()` uverse method.
//
// IsCurrent returns true iff the receiver is the captured cur of the
// topmost crossing frame on the live call stack. The comparison is
// pointer-identity on the underlying *HeapItemValue minted by
// installCrossingCur — addr/pkgPath equality is NOT sufficient.

func main(cur realm) {
	// Receiver is the current crossing frame's Cur — IsCurrent is true.
	println("cur.IsCurrent():", cur.IsCurrent())

	// cur.Previous() returns the prev field — a different HIV pointer.
	// IsCurrent must be false even though it represents a real realm.
	println("cur.Previous().IsCurrent():", cur.Previous().IsCurrent())

	// Pass cur through a non-crossing helper — pointer identity is
	// preserved across the parameter pass; IsCurrent is still true.
	println("helper(cur):", checkCurrent(0, cur))

	// Pass cur.Previous() through the same helper — false.
	println("helper(prev):", checkCurrent(0, cur.Previous()))

	// Capture cur in a local variable, pass through helper. Same
	// identity, true.
	captured := cur
	println("helper(captured):", checkCurrent(0, captured))

	// Take a derived value, hold it across calls — still pointer-
	// distinct from the topmost crossing frame's Cur.
	prev := cur.Previous()
	println("helper(prev held):", checkCurrent(0, prev))
}

// (_ int, rlm realm) is the non-crossing-helper convention: a function
// whose first realm-typed parameter is *not* named `cur` would otherwise
// be interpreted as a crossing function. The leading int sentinel keeps
// this a plain helper.
func checkCurrent(_ int, rlm realm) bool {
	return rlm.IsCurrent()
}

// Output:
// cur.IsCurrent(): true
// cur.Previous().IsCurrent(): false
// helper(cur): true
// helper(prev): false
// helper(captured): true
// helper(prev held): false
