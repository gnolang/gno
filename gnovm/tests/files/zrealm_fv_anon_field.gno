// PKGPATH: gno.land/r/fvanonfield
package fvanonfield

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	p := fvx.HolderField()
	h := p.H
	println(h.T.N)
}

// Error:
// cannot copy gno.land/r/tests/vm/fvx.T outside realm gno.land/r/tests/vm/fvx (use a pointer)

// TypeCheckError:
// gno.land/r/fvanonfield/zrealm_fv_anon_field.gno:8:2: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvanonfield/zrealm_fv_anon_field.gno:8:7: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
