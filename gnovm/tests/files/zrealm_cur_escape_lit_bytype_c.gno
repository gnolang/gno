// PKGPATH: gno.land/r/t73/escape_lit_bytype_c
// The rule is by type: an inner literal that returns nil.
package escape_lit_bytype_c

func main(cur realm) {
	none := func() realm { return nil }
	println(none() == nil)
}

// Error:
// gno.land/r/t73/escape_lit_bytype_c/zrealm_cur_escape_lit_bytype_c.gno:6:10-37: a function literal's result cannot hold a realm value: .uverse.realm
