// PKGPATH: gno.land/r/demo/curdecltype
package curdecltype

// A package-level type name is a binding (predefineRecursively2).

type cur int

func main() {
	var c cur = 1
	println(c)
}

// Output:
// (1 gno.land/r/demo/curdecltype.cur)

// TypeCheckError:
// gno.land/r/demo/curdecltype/zrealm_cur_decl_type.gno:6:6: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
