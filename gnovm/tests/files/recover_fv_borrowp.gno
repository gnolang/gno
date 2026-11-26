// PKGPATH: gno.land/r/tests/vm/recoverfvborrowp
package recoverfvborrowp

import "gno.land/r/tests/vm/recoverfvowner"

// A library method on an object another realm stores, called through a
// function value, is recovered here (master aborts); it cannot write.

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	c := recoverfvowner.GetC()
	ops := []func(){func() { c.Set("x") }, func() { c.Inc() }}
	println("rec:", try(ops[0]))
	println("write:", try(ops[1]))
	println("owner state:", c.N())
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// write: cannot modify object owned by another realm from library code (call a crossing function of the owning realm): c<VPBlock(1,0)>.n
// owner state: 0
// OUTCOME caught
