// PKGPATH: gno.land/r/ctw_iter
// A realm's stateful iterator handed to a caller cannot be advanced without
// crossing: Next is the realm's code, entered read-only, and the iterator was
// made in an earlier read-only stretch.
package ctw_iter

import "gno.land/r/tests/vm/ctw"

func main(cur realm) {
	it := ctw.NewIter(3)
	it.Next()
}

// Error:
// gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): it<VPBlock(1,0)>.i
