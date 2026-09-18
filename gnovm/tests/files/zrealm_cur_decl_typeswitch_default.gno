// PKGPATH: gno.land/r/demo/curdecltypeswitchdef
package curdecltypeswitchdef

// The other type-switch shape: when the switched value is itself realm-typed,
// the `default` clause (and any multi-type clause) binds `cur` with the tag
// type, so it is realm-typed there too and is refused for the same reason.

func main(cur realm) {
	var x realm = cur
	switch cur := x.(type) {
	default:
		println("declared realm cur:", cur.PkgPath())
	}
	println("ok")
}

// Error:
// gno.land/r/demo/curdecltypeswitchdef/zrealm_cur_decl_typeswitch_default.gno:11:2-12:48: `cur` is reserved: it may only be declared as the first parameter of a crossing function
