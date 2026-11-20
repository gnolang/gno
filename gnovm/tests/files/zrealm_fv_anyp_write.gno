// PKGPATH: gno.land/r/fvanypwrite
package fvanypwrite

import "gno.land/r/tests/vm/fvx"

func main(cur realm) {
	x := any(fvx.Ptr())
	x.(*fvx.T).N = 9
}

// Error:
// cannot directly modify readonly tainted object (call a crossing function of the owning realm): x<VPBlock(1,1)>.(*(typeval{gno.land/r/tests/vm/fvx.T})).N
