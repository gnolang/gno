// PKGPATH: gno.land/r/launderrh_conv_external_return
package launderrh_conv_external_return

import "gno.land/r/tests/vm/launderrvictim"

// `return f()` converts each result the way `return x` does, so it cannot forge
// a foreign realm's type any more than `return x` can.

func mk() (struct{ Field string }, bool) { return struct{ Field string }{"xxx"}, true }

func ret() (launderrvictim.Immutable, bool) { return mk() }

func main(cur realm) {
	v, ok := ret()
	println(v.Field, ok)
}

// Error:
// cannot allocate gno.land/r/tests/vm/launderrvictim.Immutable in realm gno.land/r/launderrh_conv_external_return

// TypeCheckError:
// gno.land/r/launderrh_conv_external_return/zrealm_launder_rdata_conv_external_return.gno:11:13: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_external_return/zrealm_launder_rdata_conv_external_return.gno:14:2: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_external_return/zrealm_launder_rdata_conv_external_return.gno:14:11: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer)
