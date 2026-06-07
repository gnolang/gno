// PKGPATH: gno.land/r/demo/crossprevious
package crossprevious

// cross requires a bare identifier. A method-call expression like
// cur.Previous() is rejected at preprocess time. Even if it were
// accepted, cur.Previous() returns a different realm whose HIV is
// not the topmost frame's Cur — the runtime IsCurrent-strict check
// would reject it. The preprocess-time check catches it earlier.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross(cur.Previous())))
}

// Error:
// gno.land/r/demo/crossprevious/zrealm_cross_previous.gno:15:17-38: cross argument must be a bare realm-typed identifier (a name, not an expression)
