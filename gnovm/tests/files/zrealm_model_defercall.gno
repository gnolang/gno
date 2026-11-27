// PKGPATH: gno.land/r/tests/vm/modeldefercall
package modeldefercall

var count int
var armed bool

// lookVar holds a declared function: calling it calls a function value.
var lookVar = look

func look(h func()) { h() }

// A deferred literal that calls look keeps f's loan; deferring a call that
// takes a literal capturing f ends it, whether look is named or a value.
func deferLit(f func()) {
	func() { defer func() { look(func() { f() }) }() }()
	armed = true
	f()
	armed = false
}

func deferByName(f func()) {
	func() { defer look(func() { f() }) }()
	armed = true
	f()
	armed = false
}

func deferByValue(f func()) {
	func() { defer lookVar(func() { f() }) }()
	armed = true
	f()
	armed = false
}

func main(cur realm) {
	r := try(func() { deferLit(func() { bump() }) })
	println("d01 deferred literal calls look, then f", count, r)
	r = try(func() { deferByName(func() { bump() }) })
	println("d02 deferred call of look, then f", count, r)
	r = try(func() { deferByValue(func() { bump() }) })
	println("d03 deferred call through a package var, then f", count, r)
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
// d01 deferred literal calls look, then f 1 undefined
// d02 deferred call of look, then f 1 gno.land/r/tests/vm/modeldefercall's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): count<~VPBlock(4,0)>
// d03 deferred call through a package var, then f 1 gno.land/r/tests/vm/modeldefercall's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): count<~VPBlock(4,0)>
