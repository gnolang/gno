// PKGPATH: gno.land/r/demo/curdeclresult
package curdeclresult

// A named realm-typed result is a declaration too, and it is not the identity
// of any frame, so it is refused like the rest.

func f(x int) (cur realm) {
	return
}

func main() {
	_ = f(0)
	println("ok")
}

// Error:
// gno.land/r/demo/curdeclresult/zrealm_cur_decl_result.gno:7:1-26: a realm-typed `cur` may only be declared as the first parameter of a crossing function
