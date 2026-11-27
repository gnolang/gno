// PKGPATH: gno.land/r/fvvar
package fvvar

import "gno.land/r/tests/vm/fvx"

var v fvx.T

func main(cur realm) {
	println(v.N)
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvvar

// TypeCheckError:
// gno.land/r/fvvar/zrealm_fv_var.gno:6:5: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
