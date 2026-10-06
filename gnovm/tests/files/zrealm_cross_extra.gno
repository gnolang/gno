// PKGPATH: gno.land/r/demo/crossextra
package crossextra

// cross with more than one argument is rejected at preprocess time.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross(cur, cur)))
}

// Error:
// gno.land/r/demo/crossextra/zrealm_cross_extra.gno:11:17-32: cross takes exactly one argument: the in-scope realm

// TypeCheckError:
// gno.land/r/demo/crossextra/zrealm_cross_extra.gno:11:28: too many arguments in call to cross
// 	have (realm, realm)
// 	want (realm)
