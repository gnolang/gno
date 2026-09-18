// PKGPATH: gno.land/r/demo/curdecltypelocal
package curdecltypelocal

// A function-local type declaration is a binding too, predefined through the
// same path as a package-level one.

func main() {
	type cur int
	var c cur = 1
	println(c)
}

// Error:
// gno.land/r/demo/curdecltypelocal/zrealm_cur_decl_type_local.gno:8:7-14: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
