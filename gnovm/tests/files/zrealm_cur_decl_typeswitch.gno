// PKGPATH: gno.land/r/demo/curdecltypeswitch
package curdecltypeswitch

// A type switch's `cur := x.(type)` is a binding per clause, refused at the
// deploy gate.

func main(cur realm) {
	var x any = cur
	switch cur := x.(type) {
	case realm:
		println("declared realm cur:", cur.PkgPath())
	}
	println("ok")
}

// Output:
// declared realm cur: gno.land/r/demo/curdecltypeswitch
// ok

// TypeCheckError:
// gno.land/r/demo/curdecltypeswitch/zrealm_cur_decl_typeswitch.gno:9:9: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
