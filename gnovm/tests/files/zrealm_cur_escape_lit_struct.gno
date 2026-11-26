// PKGPATH: gno.land/r/t73/escape_lit_struct
// A realm inside a struct result of a literal is refused too.
package escape_lit_struct

func main(cur realm) {
	get := func() struct{ r realm } { return struct{ r realm }{cur} }
	println(get().r.PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit_struct/zrealm_cur_escape_lit_struct.gno:6:9-67: a function literal's result cannot hold a realm value: struct{r .uverse.realm}
