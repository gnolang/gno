// PKGPATH: gno.land/r/tests/vm/lendrelaykeep
package lendrelaykeep

import plib "gno.land/p/demo/tests/lendcalllib"

var count int

// relay lends a crossing wrapper of its lent f to plib.Relay, which keeps it:
// the wrapper and f both lose their loans there, so the kept wrapper runs
// read-only even while relay is still running.
func relay(cur realm, f func(), box *func(realm)) {
	plib.Relay(func(cur realm) { f() }, box)
	(*box)(cur)
}

func main(cur realm) {
	var box func(realm)
	r := try(func() { relay(cur, func() { count++ }, &box) })
	println("f01 kept wrapper run while relay is still running", count, r)
	r = try(func() { box(cur) })
	println("f02 the same wrapper run after the call returned", count, r)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// f01 kept wrapper run while relay is still running 0 gno.land/r/tests/vm/lendrelaykeep's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendrelaykeep) package{})).count
// f02 the same wrapper run after the call returned 0 gno.land/r/tests/vm/lendrelaykeep's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendrelaykeep) package{})).count
