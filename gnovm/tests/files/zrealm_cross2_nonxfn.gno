// PKGPATH: gno.land/r/demo/cross2nonxfn
package cross2nonxfn

// cross2(rlm) at Args[0] of a NON-crossing function is rejected at
// preprocess time. The parent CallExpr's Func must resolve to a
// crossing function (one whose first param is `cur realm`).
//
// Here `helper` uses the `(_ int, rlm realm, ...)` migration
// convention — it's non-crossing-syntactically (first param is int,
// not realm). cross2 at Args[0] doesn't make sense here, and is
// rejected.

func helper(_ int, rlm realm) string {
	return rlm.PkgPath()
}

func main(cur realm) {
	// cross2 at Args[0] of a non-crossing call.
	println(helper(cross2(cur), nil))
}

// Error:
// gno.land/r/demo/cross2nonxfn/zrealm_cross2_nonxfn.gno:19:17-28: cross2(rlm) can only be used as the first argument to a crossing-function call

// TypeCheckError:
// gno.land/r/demo/cross2nonxfn/zrealm_cross2_nonxfn.gno:19:17: cannot use cross2(cur) (value of interface type realm) as int value in argument to helper
