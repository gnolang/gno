// PKGPATH: gno.land/r/demo/curdeclfunctype
package curdeclfunctype

// A function type with a first parameter named `cur` must be crossing;
// `func(cur int)` is refused like the FuncDecl form.

type F func(cur int)

func main() {
	var f F
	_ = f
}

// Error:
// gno.land/r/demo/curdeclfunctype/zrealm_cur_decl_functype.gno:7:8-21: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
