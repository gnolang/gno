// PKGPATH: gno.land/r/fvforgero
package fvforgero

import "gno.land/r/tests/vm/fvx"

// The closure runs inside fvx's read-only code: the conversion is judged by
// the realm that entered it, and refused.
func main(cur realm) {
	var p *fvx.T
	fvx.Do(func() { p = (*fvx.T)(&struct{ N int }{5}) })
	println(fvx.Read(cross(cur), p))
}

// Error:
// illegal conversion to external realm type

// TypeCheckError:
// gno.land/r/fvforgero/zrealm_fv_forge_ro.gno:10:22: cannot hold a value of gno.land/r/tests/vm/fvx.T outside its realm (use a pointer)
