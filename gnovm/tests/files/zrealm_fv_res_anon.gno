// PKGPATH: gno.land/r/fvresanon
package fvresanon

import "gno.land/r/tests/vm/fvx"

func f() (r struct{ X fvx.T }, p *fvx.T) { p = &r.X; return }

func main(cur realm) {
	_, p := f()
	p.N = 5
	println(fvx.Read(cross(cur), p))
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.T in realm gno.land/r/fvresanon

// TypeCheckError:
// gno.land/r/fvresanon/zrealm_fv_res_anon.gno:6:13: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvresanon/zrealm_fv_res_anon.gno:9:10: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
