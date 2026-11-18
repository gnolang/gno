// PKGPATH: gno.land/r/fvresh
package fvresh

import "gno.land/r/tests/vm/fvx"

func f() (h fvx.H, p *fvx.H) { p = &h; return }

func main(cur realm) {
	_, p := f()
	p[0] = 7
	println(p[0])
}

// Error:
// cannot allocate gno.land/r/tests/vm/fvx.H in realm gno.land/r/fvresh

// TypeCheckError:
// gno.land/r/fvresh/zrealm_fv_res_h.gno:6:13: cannot hold a value of gno.land/r/tests/vm/fvx.H outside its realm (use a pointer); gno.land/r/fvresh/zrealm_fv_res_h.gno:9:10: cannot hold a value of gno.land/r/tests/vm/fvx.H outside its realm (use a pointer)
