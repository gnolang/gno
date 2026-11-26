// PKGPATH: gno.land/r/tests/vm/t57bump
package t57bump

import lib "gno.land/p/demo/tests/t57lib"

// A library closure, called by value from the library's own receiver method,
// writes the library's Box: it runs read-only, so the write is refused before
// the immutable-package check (A97).
func main(cur realm) {
	println(lib.Lib.Bump())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): b<VPBlock(1,0)>.n
