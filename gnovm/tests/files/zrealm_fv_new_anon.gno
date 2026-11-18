// PKGPATH: gno.land/r/fvnewanon
package fvnewanon

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := new(struct{ X fvx.T })
	println(p.X.N)
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvnewanon

// TypeCheckError:
// gno.land/r/fvnewanon/zrealm_fv_new_anon.gno:7:19: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
