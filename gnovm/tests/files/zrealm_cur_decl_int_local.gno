// PKGPATH: gno.land/r/demo/curdeclint
package curdeclint

// `cur` is reserved by name, not by type: an int local named `cur` in a
// crossing function reads as the identity and is refused like the builtins.

func main(cur realm) {
	{
		cur := 41
		_ = cur
	}
	println("ok")
}

// Error:
// gno.land/r/demo/curdeclint/zrealm_cur_decl_int_local.gno:9:3-12: `cur` is reserved: it may only be declared as the first parameter of a crossing function
