// PKGPATH: gno.land/r/tests/vm/lendplibset
package lendplibset

import lib "gno.land/p/demo/tests/lendanylib"

var own = &lib.Item{N: 1}

// A library closure, called by the library's own receiver method, writes
// through a pointer to the realm's stored object: refused.
func main(cur realm) {
	lib.Lib.Set(own, 9)
	println(own.N)
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): it<VPBlock(1,0)>.N
