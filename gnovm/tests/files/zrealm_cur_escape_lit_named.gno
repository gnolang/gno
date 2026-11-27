// PKGPATH: gno.land/r/t73/escape_lit_named
// The rule looks through a declared type to what it is made of.
package escape_lit_named

type holder struct {
	n int
	r map[string]realm
}

func main(cur realm) {
	get := func() *holder { return &holder{r: map[string]realm{"c": cur}} }
	println(get().r["c"].PkgPath())
}

// Error:
// gno.land/r/t73/escape_lit_named/zrealm_cur_escape_lit_named.gno:11:9-73: a function literal's result cannot hold a realm value: *gno.land/r/t73/escape_lit_named.holder
