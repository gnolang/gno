// PKGPATH: gno.land/r/tests/vm/t55lendframe
package t55lendframe

import plib "gno.land/p/demo/tests/t55plib"

var count int

// relay lends a crossing wrapper of its lent f to plib.Relay, which keeps it:
// the wrapper loses its own loan there, but f's loan is only ended by the
// call it was lent to returning.
func relay(cur realm, f func(), box *func(realm)) {
	plib.Relay(func(cur realm) { f() }, box)
	(*box)(cur)
}

func main(cur realm) {
	var box func(realm)
	r := try(func() { relay(cur, func() { count++ }, &box) })
	println("f01 control: kept wrapper run while the loan holds", count, r)
	r = try(func() { box(cur) })
	println("f02 the same wrapper run after the call returned", count, r)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// f01 control: kept wrapper run while the loan holds 1 undefined
// f02 the same wrapper run after the call returned 1 gno.land/r/tests/vm/t55lendframe's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/t55lendframe) package{})).count
