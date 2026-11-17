// PKGPATH: gno.land/r/ctw_ctor
// A constructor called without crossing builds the realm's object; the
// caller gets a read-only view of it.
package ctw_ctor

import "gno.land/r/tests/vm/ctw"

func main(cur realm) {
	l := ctw.NewLocal(5)
	println(l.N)
	l.Inc()
}

// Output:
// 5

// Error:
// gno.land/r/tests/vm/ctw's code was entered without crossing and is read-only (call its crossing function): l<VPBlock(1,0)>.N
