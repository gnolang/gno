// PKGPATH: gno.land/r/bar
// MAXALLOC: 100000000
package bar

import "gno.land/r/tests/vm/crossrealm"

// Composite literal of a foreign /r/-declared type from outside its
// declaring realm panics at the construction-time check
// (doOpStructLit). Authority cannot be forged.

func main() {
	x := crossrealm.LocalStruct{A: 1}
	println(x.A)
}

// Error:
// cannot allocate gno.land/r/tests/vm/crossrealm.LocalStruct in realm gno.land/r/bar

// TypeCheckError:
// gno.land/r/bar/zrealm_v2_construction_composite_fail.gno:12:2: cannot hold a value of gno.land/r/tests/vm/crossrealm.LocalStruct outside its realm (use a pointer); gno.land/r/bar/zrealm_v2_construction_composite_fail.gno:12:7: cannot hold a value of gno.land/r/tests/vm/crossrealm.LocalStruct outside its realm (use a pointer)
