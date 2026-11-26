// PKGPATH: gno.land/r/t73/escape_lit_embed_nested
// `realm` embedded through another embedded interface is refused too.
package escape_lit_embed_nested

type R interface{ realm }

type R2 interface{ R }

func main(cur realm) {
	get := func() []R2 { return []R2{cur} }
	println(get()[0].PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit_embed_nested/zrealm_cur_escape_lit_embed_nested.gno:10:9-41: a function literal's result cannot hold a realm value: []gno.land/r/t73/escape_lit_embed_nested.R2
