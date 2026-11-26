// PKGPATH: gno.land/r/t73/escape_lit_embed_named
// A declared type whose underlying type is the realm type is refused too.
package escape_lit_embed_named

type D realm

func main(cur realm) {
	get := func() D { return cur }
	println(get().PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit_embed_named/zrealm_cur_escape_lit_embed_named.gno:8:9-32: a function literal's result cannot hold a realm value: gno.land/r/t73/escape_lit_embed_named.D
