// PKGPATH: gno.land/r/demo/crossnonxfn
package crossnonxfn

// cross(rlm) at Args[0] of a NON-crossing function is rejected at
// preprocess time. The parent CallExpr's Func must resolve to a
// crossing function (one whose first param is `cur realm`).
//
// Here `helper` uses the `(_ int, rlm realm, ...)` migration
// convention — it's non-crossing-syntactically (first param is int,
// not realm). cross at Args[0] doesn't make sense here, and is
// rejected.

func helper(_ int, rlm realm) string {
	return rlm.PkgPath()
}

func main(cur realm) {
	// cross at Args[0] of a non-crossing call.
	println(helper(cross(cur), nil))
}

// Error:
// gno.land/r/demo/crossnonxfn/zrealm_cross_nonxfn.gno:19:17-27: cross(rlm) can only be used as the first argument to a crossing-function call

// TypeCheckError:
// gno.land/r/demo/crossnonxfn/zrealm_cross_nonxfn.gno:19:17: cannot use cross(cur) (value of interface type realm) as int value in argument to helper
