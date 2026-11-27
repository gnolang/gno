// PKGPATH: gno.land/r/tests/vm/lendhandoffescape
package lendhandoffescape

var count int

func esc(x func()) { _ = x }

func lent(f func()) { f() }

// A lent value handed by a by-value call to a parameter that escapes loses
// its own loan: calling it afterwards is refused.
func handOff(h func()) {
	k := esc
	k(h)
	h()
}

func main(cur realm) {
	r := try(func() { lent(func() { count++ }) })
	println("e01 control: lent and called", count, r)
	r = try(func() { handOff(func() { count++ }) })
	println("e02 handed to an escaping parameter, then called", count, r)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// e01 control: lent and called 1 undefined
// e02 handed to an escaping parameter, then called 1 gno.land/r/tests/vm/lendhandoffescape's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhandoffescape) package{})).count
