// PKGPATH: gno.land/r/demo/curdecllocal
package curdecllocal

// The old shadow shape, refused at the declaration: `cur` is reserved by
// name, not by type, so a `cur` local in an inner block is refused whatever
// its type (this one is realm-typed; `cur := 41` fails the same way).

func main(cur realm) {
	{
		cur := cur
		_ = cur
	}
	println("ok")
}

// Error:
// gno.land/r/demo/curdecllocal/zrealm_cur_decl_local.gno:10:3-13: `cur` is a declaration-reserved identifier: it may only be declared as the first parameter of a crossing function
