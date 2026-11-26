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

// TypeCheckError:
// gno.land/r/launderrh_conv_recover_pkg/zrealm_launder_rdata_conv_recover_pkg.gno:11:5: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_recover_pkg/zrealm_launder_rdata_conv_recover_pkg.gno:11:9: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer)
