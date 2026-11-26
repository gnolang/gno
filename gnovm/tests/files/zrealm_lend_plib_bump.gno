// PKGPATH: gno.land/r/tests/vm/lendplibbump
package lendplibbump

import lib "gno.land/p/demo/tests/lendanylib"

// A library closure, called by value from the library's own receiver method,
// writes the library's Box: it runs read-only, so the write is refused before
// the immutable-package check.
func main(cur realm) {
	println(lib.Lib.Bump())
}

// Error:
// cannot modify object owned by another realm from library code (call a crossing function of the owning realm): b<VPBlock(1,0)>.n
