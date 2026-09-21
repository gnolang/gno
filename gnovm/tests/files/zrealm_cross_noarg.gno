// PKGPATH: gno.land/r/demo/crossnoarg
package crossnoarg

// cross() with no arguments is rejected at preprocess time. The
// shim signature is `func cross(rlm realm) realm`, so the Go
// typechecker catches arity mismatches before our preprocessor
// rule fires.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross()))
}

// Error:
// gno.land/r/demo/crossnoarg/zrealm_cross_noarg.gno:14:17-24: cross takes exactly one argument: the in-scope realm

// TypeCheckError:
// gno.land/r/demo/crossnoarg/zrealm_cross_noarg.gno:14:23: not enough arguments in call to cross
// 	have ()
// 	want (realm)
