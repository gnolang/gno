// PKGPATH: gno.land/r/tests/vm/t57set
package t57set

import lib "gno.land/p/demo/tests/t57lib"

var own = &lib.Item{N: 1}

// A library closure, called by the library's own receiver method, writes
// through a pointer to the realm's stored object: refused (A97).
func main(cur realm) {
	lib.Lib.Set(own, 9)
	println(own.N)
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): it<VPBlock(1,0)>.N
