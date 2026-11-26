// PKGPATH: gno.land/r/tests/vm/lendcapture
package lendcapture

import plib "gno.land/p/demo/tests/lendcalllib"

var count int

// keep hands a crossing wrapper of its lent f to be kept, then runs the kept
// wrapper in this same crossing (a crossing value called with cur).
func keep(cur realm, f func()) {
	var box func(realm)
	plib.KeepX(func(cur realm) { f() }, &box)
	box(cur)
}

// A lent literal loses its loan when a value capturing it is kept, even while
// the call it was lent to is still running.
func main(cur realm) {
	r := try(func() { keep(cur, func() { count++ }) })
	println("k01 captured by a kept wrapper, run during the loan", count, r)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// k01 captured by a kept wrapper, run during the loan 0 gno.land/r/tests/vm/lendcapture's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendcapture) package{})).count
