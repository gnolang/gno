// PKGPATH: gno.land/r/test
package test

import "chain/runtime"

func Register() {
	runtime.AssertOriginCall()
}

// main is the message's script, so the closure is the entry. Asserting inside
// it is an origin call; reaching Register, a named function of this realm,
// from it is not: a script can hand Register back as a value, a closure it
// cannot.
func main() {
	fn := func() {
		runtime.AssertOriginCall()
	}
	fn()
	println("ok")

	defer func() {
		println("named func from closure panicked:", recover())
	}()
	fn2 := func() {
		Register()
	}
	fn2()
}

// Output:
// ok
// named func from closure panicked: invalid non-origin call
