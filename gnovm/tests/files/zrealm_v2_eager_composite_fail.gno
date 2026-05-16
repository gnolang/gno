// PKGPATH: gno.land/r/bar
// MAXALLOC: 100000000
package bar

import "gno.land/r/tests/vm/crossrealm"

// interrealm v2 Phase 2: composite literal of a foreign /r/-declared type
// from outside its declaring realm panics at the eager-constructor
// check (doOpStructLit). Authority cannot be forged.

func main() {
	x := crossrealm.LocalStruct{A: 1}
	println(x.A)
}

// Error:
// cannot allocate gno.land/r/tests/vm/crossrealm.LocalStruct in realm gno.land/r/bar
