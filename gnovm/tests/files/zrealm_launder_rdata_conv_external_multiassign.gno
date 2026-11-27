// PKGPATH: gno.land/r/launderrh_conv_external_multiassign
package launderrh_conv_external_multiassign

import "gno.land/r/tests/vm/launderrvictim"

// A multi-value assignment converts each result the way a single-value one
// does, so it cannot forge a foreign realm's type any more than `v = x` can.

func mk() (struct{ Field string }, bool) { return struct{ Field string }{"xxx"}, true }

func main(cur realm) {
	var v launderrvictim.Immutable
	var ok bool
	v, ok = mk()
	println(v.Field, ok)
}

// Error:
// cannot allocate gno.land/r/tests/vm/launderrvictim.Immutable in realm gno.land/r/launderrh_conv_external_multiassign

// TypeCheckError:
// gno.land/r/launderrh_conv_external_multiassign/zrealm_launder_rdata_conv_external_multiassign.gno:12:6: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_external_multiassign/zrealm_launder_rdata_conv_external_multiassign.gno:14:2: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer)
