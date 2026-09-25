// PKGPATH: gno.land/r/demo/crossnoarg
package crossnoarg

// cross() with no arguments is rejected at preprocess time. The
// shim signature is `func cross(rlm realm, s ...any) realm`, so the Go
// typechecker catches the missing realm before our preprocessor
// rule fires.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross()))
}

// Error:
// gno.land/r/demo/crossnoarg/zrealm_cross_noarg.gno:14:17-24: cross takes the in-scope realm and optionally the chain.Coins to send

// TypeCheckError:
// gno.land/r/demo/crossnoarg/zrealm_cross_noarg.gno:14:23: not enough arguments in call to cross
// 	have ()
// 	want (realm, ...any)
