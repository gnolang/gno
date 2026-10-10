// PKGPATH: gno.land/r/demo/curdecllocal
package curdecllocal

// The old shadow shape, refused now at the declaration: a realm-typed local
// named `cur` is not any frame's identity, and keeping "a realm-typed `cur`
// is always the crossing parameter" true is what lets the write rules key on
// name plus type alone. An inner block does not help — the name is reserved
// wherever it is declared.

func main(cur realm) {
	{
		cur := cur
		_ = cur
	}
	println("ok")
}

// Error:
// gno.land/r/demo/curdecllocal/zrealm_cur_decl_local.gno:12:3-13: a realm-typed `cur` may only be declared as the first parameter of a crossing function
