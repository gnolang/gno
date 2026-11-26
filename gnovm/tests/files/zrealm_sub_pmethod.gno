// PKGPATH: gno.land/r/t75a
package t75a

import (
	lib "gno.land/p/demo/tests/t75sub"
	"gno.land/r/tests/vm/t75y"
)

var own = &lib.Obj{}

// rlm.Sub() in a /p/ method stays allowed, also on another realm's object.
func main(cur realm) {
	println("own recv:", own.Mint("a", cur))
	println("top-level:", lib.Mint("b", cur))
	println("foreign recv:", t75y.Get().Mint("c", cur))
}

// Output:
// own recv: gno.land/r/t75a#a
// top-level: gno.land/r/t75a#b
// foreign recv: gno.land/r/t75a#c
