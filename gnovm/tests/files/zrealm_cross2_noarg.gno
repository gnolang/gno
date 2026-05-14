// PKGPATH: gno.land/r/demo/cross2noarg
package cross2noarg

// cross2() with no arguments is rejected at preprocess time. The
// shim signature is `func cross2(rlm realm) realm`, so the Go
// typechecker catches arity mismatches before our preprocessor
// rule fires.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross2()))
}

// Error:
// gno.land/r/demo/cross2noarg/zrealm_cross2_noarg.gno:14:17-25: cross2 takes exactly one argument: the in-scope realm

// TypeCheckError:
// gno.land/r/demo/cross2noarg/zrealm_cross2_noarg.gno:14:24: not enough arguments in call to cross2
// 	have ()
// 	want (realm)
