// PKGPATH: gno.land/r/demo/cross2stray
package cross2stray

// cross2(rlm) is only valid as the first argument of a crossing call.
// Standalone usage (assignment, ExprStmt, etc.) is rejected at
// preprocess time. This pins the constraint so future code can't
// accidentally use cross2 as a runtime IsCurrent-strict assertion
// outside its intended position.

func main(cur realm) {
	// Statement-level standalone — rejected.
	r := cross2(cur)
	_ = r
}

// Error:
// gno.land/r/demo/cross2stray/zrealm_cross2_stray.gno:12:7-18: cross2(rlm) can only be used as the first argument to a crossing-function call
