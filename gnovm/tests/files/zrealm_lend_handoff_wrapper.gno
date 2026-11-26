// PKGPATH: gno.land/r/tests/vm/lendhandoffwrapper
package lendhandoffwrapper

var count int

func esc(x func()) { _ = x }

func lent(f func()) { f() }

// A wrapper capturing the lent f is handed to a by-value call of a function
// that lets it escape: f's loan ends, so calling f afterwards is refused.
func handOff(f func()) {
	wrap(func() { f() })
	f()
}

func wrap(h func()) {
	k := esc
	k(h)
}

func main(cur realm) {
	r := try(func() { lent(func() { count++ }) })
	println("w01 control: lent and called", count, r)
	r = try(func() { handOff(func() { count++ }) })
	println("w02 wrapper handed off, then f", count, r)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// w01 control: lent and called 1 undefined
// w02 wrapper handed off, then f 1 gno.land/r/tests/vm/lendhandoffwrapper's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhandoffwrapper) package{})).count
