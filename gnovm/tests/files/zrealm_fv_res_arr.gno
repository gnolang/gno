// PKGPATH: gno.land/r/fvresarr
package fvresarr

import "gno.land/r/tests/vm/fvx"

func f() (r [2]fvx.T, p *fvx.T) { p = &r[1]; return }

func main(cur realm) {
	_, p := f()
	p.N = 5
	println(fvx.Read(cross(cur), p))
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvresarr

// TypeCheckError:
// gno.land/r/fvresarr/zrealm_fv_res_arr.gno:6:13: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvresarr/zrealm_fv_res_arr.gno:9:10: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
