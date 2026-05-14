// PKGPATH: gno.land/r/demo/cross2extra
package cross2extra

// cross2 with more than one argument is rejected at preprocess time.

func callee(cur realm) string {
	return cur.PkgPath()
}

func main(cur realm) {
	println(callee(cross2(cur, cur)))
}

// Error:
// gno.land/r/demo/cross2extra/zrealm_cross2_extra.gno:11:17-33: cross2 takes exactly one argument: the in-scope realm

// TypeCheckError:
// gno.land/r/demo/cross2extra/zrealm_cross2_extra.gno:11:29: too many arguments in call to cross2
// 	have (realm, realm)
// 	want (realm)
