// PKGPATH: gno.land/r/demo/curdecltype
package curdecltype

// A package-level type name is a binding (predefineRecursively2).

type cur int

func main() {
	var c cur = 1
	println(c)
}

// Error:
// gno.land/r/demo/curdecltype/zrealm_cur_decl_type.gno:6:6-13: `cur` is reserved: it may only be declared as the first parameter of a crossing function
