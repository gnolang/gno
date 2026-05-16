// PKGPATH: gno.land/r/bar
// MAXALLOC: 100000000
package bar

import "gno.land/r/tests/vm/crossrealm"

// PLAN3 Phase 2: new() on a foreign /r/-declared type from outside
// its declaring realm panics at the eager-constructor check
// (uverse new builtin). Authority cannot be forged via *T.

func main() {
	p := new(crossrealm.LocalStruct)
	println(p.A)
}

// Error:
// cannot allocate gno.land/r/tests/vm/crossrealm.LocalStruct in realm gno.land/r/bar
