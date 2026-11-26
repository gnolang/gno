// PKGPATH: gno.land/r/fvforgelocal
package fvforgelocal

import "gno.land/r/tests/vm/fvx"

type Local fvx.T

// As forge_ro, through a local named type derived from fvx.T.
func main(cur realm) {
	var p *fvx.T
	fvx.Do(func() { p = (*fvx.T)(&Local{N: 6}) })
	println(fvx.Read(cross(cur), p))
}

// Error:
// illegal conversion to external realm type

// TypeCheckError:
// gno.land/r/fvforgelocal/zrealm_fv_forge_local.gno:6:6: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer); gno.land/r/fvforgelocal/zrealm_fv_forge_local.gno:11:22: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
