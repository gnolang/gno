// PKGPATH: gno.land/r/fvarr
package fvarr

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := new(struct {
		X fvx.H
		Y fvx.A
	})
	println(len(p.X), p.Y[1].N)
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.H in realm gno.land/r/fvarr

// TypeCheckError:
// gno.land/r/fvarr/zrealm_fv_arr.gno:8:3: cannot hold a value of gno.land/r/tests/vm/fvx.H outside its realm (use a pointer); gno.land/r/fvarr/zrealm_fv_arr.gno:9:3: cannot hold a value of gno.land/r/tests/vm/fvx.A outside its realm (use a pointer)
