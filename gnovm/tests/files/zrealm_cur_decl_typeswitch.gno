// PKGPATH: gno.land/r/demo/curdecltypeswitch
package curdecltypeswitch

// A type switch's `cur := x.(type)` is a binding, refused at the switch
// statement whatever the clauses.

func main(cur realm) {
	var x any = cur
	switch cur := x.(type) {
	case realm:
		println("declared realm cur:", cur.PkgPath())
	}
	println("ok")
}

// Error:
// gno.land/r/demo/curdecltypeswitch/zrealm_cur_decl_typeswitch.gno:9:2-12:3: `cur` is a declaration-reserved identifier: it may only be declared as the first parameter of a crossing function
