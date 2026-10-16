// PKGPATH: gno.land/r/demo/curdecltypeswitch
package curdecltypeswitch

// A type switch's `cur := x.(type)` declares a fresh binding per clause, typed
// by the clause: in `case realm:` it is realm-typed. That is a declaration
// like the others and is refused for the same reason; the clause variable is
// defined in the SwitchClauseStmt TRANS_BLOCK arm, not through defineOrDecl.

func main(cur realm) {
	var x any = cur
	switch cur := x.(type) {
	case realm:
		println("declared realm cur:", cur.PkgPath())
	}
	println("ok")
}

// Error:
// gno.land/r/demo/curdecltypeswitch/zrealm_cur_decl_typeswitch.gno:12:2-13:48: a realm-typed `cur` may only be declared as the first parameter of a crossing function
