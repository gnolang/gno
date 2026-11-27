// PKGPATH: gno.land/r/t73/escape_lit_bytype_a
// The rule is by type: a literal called where written is refused too.
package escape_lit_bytype_a

func main(cur realm) {
	println(func() realm { return cur }().PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit_bytype_a/zrealm_cur_escape_lit_bytype_a.gno:6:10-37: a function literal's result cannot hold a realm value: .uverse.realm
