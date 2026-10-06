// PKGPATH: gno.land/r/demo/crossstray
package crossstray

// cross(rlm) is only valid as the first argument of a crossing call.
// Standalone usage (assignment, ExprStmt, etc.) is rejected at
// preprocess time. This pins the constraint so future code can't
// accidentally use cross as a runtime IsCurrent-strict assertion
// outside its intended position.

func main(cur realm) {
	// Statement-level standalone — rejected.
	r := cross(cur)
	_ = r
}

// Error:
// gno.land/r/demo/crossstray/zrealm_cross_stray.gno:12:7-17: cross(rlm) can only be used as the first argument to a crossing-function call
