// PKGPATH: gno.land/r/t73/escape_lit_bytype_b
// The rule is by type: a package-level literal that returns its argument.
package escape_lit_bytype_b

var pick = func(_ int, a realm) realm { return a }

func main(cur realm) {
	println(pick(0, cur).PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit_bytype_b/zrealm_cur_escape_lit_bytype_b.gno:5:12-51: a function literal's result cannot hold a realm value: .uverse.realm
