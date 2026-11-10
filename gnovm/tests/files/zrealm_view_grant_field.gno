// PKGPATH: gno.land/r/demo/viewfield
package viewfield

import "gno.land/r/tests/vm/viewdefault"

// A grant through a field address marks the object holding the field, the
// same object borrow rule #2 inspects for a method on the pointer: before the
// grant both setters are refused, after it both commit.

func attempt(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			println(name + ": refused")
		}
	}()
	fn()
	println(name + ": ok")
}

func main(cur realm) {
	p := viewdefault.GetPair()
	attempt("SetA before grant", func() { p.SetA("a", "x") })
	viewdefault.GrantPairA(cross(cur))
	attempt("SetA after grant", func() { p.SetA("a", "x") })
	attempt("SetB after grant", func() { p.SetB("b", "y") })
	println("owner sees:", p.Get("a"), p.Get("b"))
}

// Output:
// SetA before grant: refused
// SetA after grant: ok
// SetB after grant: ok
// owner sees: x y
