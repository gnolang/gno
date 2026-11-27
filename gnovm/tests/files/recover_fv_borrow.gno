// PKGPATH: gno.land/r/tests/vm/recoverfvborrow
package recoverfvborrow

import "gno.land/r/tests/vm/recoverfvowner"

// A method on another realm's object, called through a function value, has
// no write power over it, so its panic is recovered here (master aborts).

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

func main(cur realm) {
	o := recoverfvowner.Get()
	ops := []func(){func() { o.Boom() }, func() { o.Inc() }}
	println("rec:", try(ops[0]))
	println("write:", try(ops[1]))
	println("owner state:", o.N())
	println("OUTCOME caught")
}

// Output:
// rec: owner boom
// write: gno.land/r/tests/vm/recoverfvowner's code was entered without crossing and is read-only (call its crossing function): o<VPBlock(1,0)>.n
// owner state: 0
// OUTCOME caught
