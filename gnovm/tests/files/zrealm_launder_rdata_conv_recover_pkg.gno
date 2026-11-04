// PKGPATH: gno.land/r/launderrh_conv_recover_pkg
package launderrh_conv_recover_pkg

import "gno.land/r/tests/vm/launderrvictim"

// A refused conversion in a package-level var keeps its message: no call is on
// the stack, so nothing can recover it.

var s = struct{ Field string }{"x"}

var v = launderrvictim.Immutable(s)

func main(cur realm) {
	println(v.Field)
}

// Error:
// illegal conversion to external realm type
