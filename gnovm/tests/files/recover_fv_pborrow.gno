// PKGPATH: gno.land/r/tests/vm/recoverfvpborrow
package recoverfvpborrow

import "gno.land/p/demo/tests/recoverfv"

// A library method on an object the library stores, through a function value,
// has no write power over the library after init: its panic is recovered here
// (master aborts); recover_fv_pborrow_write shows the write is refused.
func main(cur realm) {
	c := recoverfv.Shared()
	ops := []func(){func() { c.Set("x") }}
	println("rec:", try(ops[0]))
	println("lib state:", c.N())
	println("OUTCOME caught")
}

func try(fn func()) (rec any) {
	defer func() { rec = recover() }()
	fn()
	return
}

// Output:
// rec: int is not of type string
// lib state: 0
// OUTCOME caught
