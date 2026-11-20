// PKGPATH: gno.land/r/ctw_method
// The realm's own method on its stored object, called without crossing, is read-only.
package ctw_method

import "gno.land/r/tests/vm/ctw"

func main(cur realm) {
	ctw.Saved.Inc()
}

// Error:
// gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): l<VPBlock(1,0)>.N
