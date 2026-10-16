// PKGPATH: gno.land/r/demo/curdeclfunctype2
package curdeclfunctype2

// A function type's later parameter named `cur` is refused by name, whatever
// its type; the realm-typed form is pinned by zrealm_cur_alias_second.gno.

type F func(x int, cur int)

func main() {
	var f F
	_ = f
}

// Error:
// gno.land/r/demo/curdeclfunctype2/zrealm_cur_decl_functype_second.gno:7:8-28: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
