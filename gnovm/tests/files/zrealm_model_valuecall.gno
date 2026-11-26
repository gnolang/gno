// PKGPATH: gno.land/r/tests/vm/modelvaluecall
package modelvaluecall

var count int
var armed bool

// lookVar holds a declared function: calling it calls a function value.
var lookVar = look

func look(h func()) { h() }

// A literal capturing the lent f, written in place for a call of this realm:
// f keeps its loan when look is called by name, and loses it when look is
// reached as a value, which runs read-only in this same realm.
func byName(f func()) {
	look(func() { f() })
	armed = true
	f()
	armed = false
}

func byPkgVar(f func()) {
	lookVar(func() { f() })
	armed = true
	f()
	armed = false
}

func byLocal(f func()) {
	g := lookVar
	g = look
	g(func() { f() })
	armed = true
	f()
	armed = false
}

func main(cur realm) {
	r := try(func() { byName(func() { bump() }) })
	println("v01 call by name, then f", count, r)
	r = try(func() { byPkgVar(func() { bump() }) })
	println("v02 call through a package var, then f", count, r)
	r = try(func() { byLocal(func() { bump() }) })
	println("v03 call through a reassigned local, then f", count, r)
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
// v01 call by name, then f 1 undefined
// v02 call through a package var, then f 1 gno.land/r/tests/vm/modelvaluecall's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): count<~VPBlock(4,0)>
// v03 call through a reassigned local, then f 1 gno.land/r/tests/vm/modelvaluecall's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): count<~VPBlock(4,0)>
