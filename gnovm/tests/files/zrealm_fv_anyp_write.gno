// PKGPATH: gno.land/r/fvanypwrite
package fvanypwrite

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	x := any(fvx.Ptr())
	x.(*fvx.T).N = 9
}

// Error:
// cannot directly modify readonly tainted object (write it in a named function of the realm that owns it, during a crossing call into that realm; a /p/ or standard-library package's object cannot be written after its init): x<VPBlock(1,1)>.(*(typeval{gno.land/r/tests/vm/fvx.T})).N
