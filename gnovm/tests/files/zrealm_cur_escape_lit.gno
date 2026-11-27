// PKGPATH: gno.land/r/t73/escape_lit
// A function literal may not return `cur`: it would outlive the call.
package escape_lit

func main(cur realm) {
	get := func() realm { return cur }
	println(get().PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit/zrealm_cur_escape_lit.gno:6:9-36: a function literal's result cannot hold a realm value: .uverse.realm
