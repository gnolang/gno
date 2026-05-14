// PKGPATH: gno.land/r/demo/cross2previous
package cross2previous

// cross2 requires a bare identifier. A method-call expression like
// cur.Previous() is rejected at preprocess time. Even if it were
// accepted, cur.Previous() returns a different realm whose HIV is
// not the topmost frame's Cur — the runtime IsCurrent-strict check
// would reject it. The preprocess-time check catches it earlier.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross2(cur.Previous())))
}

// Error:
// gno.land/r/demo/cross2previous/zrealm_cross2_previous.gno:15:17-39: cross2 argument must be a bare realm-typed identifier (a name, not an expression)
