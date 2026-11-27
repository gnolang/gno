// PKGPATH: gno.land/r/t73/escape_lit_embed
// An interface that embeds `realm` is the realm type under another name.
package escape_lit_embed

type R interface{ realm }

func main(cur realm) {
	get := func() R { return cur }
	println(get().PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit_embed/zrealm_cur_escape_lit_embed.gno:8:9-32: a function literal's result cannot hold a realm value: gno.land/r/t73/escape_lit_embed.R
