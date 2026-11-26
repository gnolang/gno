// PKGPATH: gno.land/r/tests/vm/lendcrossing
package lendcrossing

import plib "gno.land/p/demo/tests/lendcalllib"

var count int

func run(cur realm, f func()) { f() }

func bounce(cur realm, f func()) { run(cross(cur), f) }

// A lent literal writes only in the crossing call that lent it: run in a new
// crossing call of the same realm, the lending frame still live, it is
// refused (the panic crosses a realm boundary, so it ends the run).
func main(cur realm) {
	r := try(func() { plib.Call(func() { count++ }) })
	println("x01 control: lent and run in the lending crossing", count, r)
	bounce(cur, func() { count++ })
	println("x02 not reached", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// x01 control: lent and run in the lending crossing 1 undefined

// Error:
// gno.land/r/tests/vm/lendcrossing's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendcrossing) package{})).count
