// PKGPATH: gno.land/r/tests/vm/modelhandoff
package modelhandoff

import "gno.land/r/tests/vm/lendcaller"

var count int
var armed bool

func look(h func()) { h() }

// A literal capturing the lent f, written in place for a call: f keeps its
// loan through this realm's own call, and loses it when the call runs
// read-only (another realm's plain function).
func keep(f func()) {
	look(func() { f() })
	armed = true
	f()
	armed = false
}

func readOnly(f func()) {
	lendcaller.Call(func() { f() })
	armed = true
	f()
	armed = false
}

func main(cur realm) {
	r := try(func() { keep(func() { bump() }) })
	println("h01 own call, then f", count, r)
	r = try(func() { readOnly(func() { bump() }) })
	println("h02 read-only call, then f", count, r)
}

func bump() {
	if armed {
		count++
	}
}

func try(fn func()) (r any) {
	defer func() { armed = false; r = recover() }()
	fn()
	return nil
}

// Output:
// h01 own call, then f 1 undefined
// h02 read-only call, then f 1 gno.land/r/tests/vm/modelhandoff's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): count<~VPBlock(4,0)>
