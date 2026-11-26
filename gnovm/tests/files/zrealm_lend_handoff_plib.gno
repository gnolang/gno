// PKGPATH: gno.land/r/tests/vm/lendhandoffplib
package lendhandoffplib

import lib "gno.land/p/demo/tests/chaldlib"

var count int

func nop() {}

// A wrapper capturing the lent f is handed to a by-value call of a /p/
// function whose parameter escapes: f's loan ends, so f is then refused.
func handOff(f func()) {
	wrap(func() { f() })
	f()
}

func wrap(h func()) {
	k := lib.Branch
	k(h, nop, true)
}

func main(cur realm) {
	r := try(func() { lib.Call(func() { count++ }) })
	println("p01 control: lent to a /p/ function", count, r)
	r = try(func() { handOff(func() { count++ }) })
	println("p02 wrapper handed off, then f", count, r)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// p01 control: lent to a /p/ function 1 undefined
// p02 wrapper handed off, then f 1 gno.land/r/tests/vm/lendhandoffplib's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/tests/vm/lendhandoffplib) package{})).count
