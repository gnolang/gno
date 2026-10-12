// PKGPATH: gno.land/r/demo/crossrlm
package crossrlm

// cross(rlm) works with the `(_ int, rlm realm, ...)` non-crossing
// helper convention used throughout the migration. The helper isn't
// itself a crossing function — its first parameter is `int`, not
// `realm` — but `rlm` is in scope and threadable. cross(rlm)
// validates IsCurrent at runtime, so the rlm doesn't have to come
// from a containing crossing function.

func callee(cur realm) string {
	return cur.PkgPath()
}

// helper is a non-crossing function that takes rlm via the (_ int,
// rlm realm) convention. It can use cross(rlm) directly.
func helper(_ int, rlm realm) string {
	return callee(cross(rlm))
}

func main(cur realm) {
	// Call the helper, threading cur as rlm.
	println("via helper:   ", helper(0, cur))
	// And bare cross from main for comparison.
	println("bare cross:   ", callee(cross(cur)))
}

// Output:
// via helper:    gno.land/r/demo/crossrlm
// bare cross:    gno.land/r/demo/crossrlm
