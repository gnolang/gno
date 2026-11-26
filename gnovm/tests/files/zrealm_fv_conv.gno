// PKGPATH: gno.land/r/fvconv
package fvconv

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	t := fvx.T(struct{ N int }{N: 5})
	println(t.N)
}

// Error:
// illegal conversion to external realm type (call a function of the realm that declares the type to make it)

// TypeCheckError:
// gno.land/r/fvconv/zrealm_fv_conv.gno:7:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvconv/zrealm_fv_conv.gno:7:7: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
