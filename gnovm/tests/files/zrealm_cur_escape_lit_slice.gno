// PKGPATH: gno.land/r/t73/escape_lit_slice
// A realm inside a slice result of a literal is refused too.
package escape_lit_slice

func main(cur realm) {
	get := func() (int, []realm) { return 1, []realm{cur} }
	_, rs := get()
	println(rs[0].PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit_slice/zrealm_cur_escape_lit_slice.gno:6:9-57: a function literal's result cannot hold a realm value: [].uverse.realm
