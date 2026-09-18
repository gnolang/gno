// PKGPATH: gno.land/r/demo/curdeclvar
package curdeclvar

// A package-level realm-typed `cur` used to be legal, and the nil placeholder
// in test packages relied on it. It is refused now: the name is reserved for
// a crossing function's first parameter, so that a realm-typed `cur` anywhere
// in the tree can be read as "this frame's identity" with no scope
// resolution. A realm-typed package-level value under any other name is still
// legal (see zrealm_cur_other_legal.gno).

var cur realm

func main() {
	println("unused")
}

// Error:
// gno.land/r/demo/curdeclvar/zrealm_cur_decl_var.gno:11:5-14: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
