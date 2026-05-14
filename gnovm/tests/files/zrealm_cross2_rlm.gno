// PKGPATH: gno.land/r/demo/cross2rlm
package cross2rlm

// cross2(rlm) works with the `(_ int, rlm realm, ...)` non-crossing
// helper convention used throughout the migration. The helper isn't
// itself a crossing function — its first parameter is `int`, not
// `realm` — but `rlm` is in scope and threadable. cross2(rlm)
// validates IsCurrent at runtime, so the rlm doesn't have to come
// from a containing crossing function.

func callee(cur realm) string {
	return cur.PkgPath()
}

// helper is a non-crossing function that takes rlm via the (_ int,
// rlm realm) convention. It can use cross2(rlm) directly.
func helper(_ int, rlm realm) string {
	return callee(cross2(rlm))
}

func main(cur realm) {
	// Call the helper, threading cur as rlm.
	println("via helper:   ", helper(0, cur))
	// And bare cross from main for comparison.
	println("bare cross:   ", callee(cross))
}

// Output:
// via helper:    gno.land/r/demo/cross2rlm
// bare cross:    gno.land/r/demo/cross2rlm
