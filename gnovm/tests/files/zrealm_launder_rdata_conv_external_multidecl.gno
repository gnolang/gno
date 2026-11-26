// PKGPATH: gno.land/r/launderrh_conv_external_multidecl
package launderrh_conv_external_multidecl

import "gno.land/r/tests/vm/launderrvictim"

// A multi-value declaration converts each result the way a single-value one
// does, so it cannot forge a foreign realm's type any more than `var v T = x`
// can.

func mk() (struct{ Field string }, struct{ Field string }) {
	return struct{ Field string }{"xxx"}, struct{ Field string }{"yyy"}
}

func main(cur realm) {
	var v, w launderrvictim.Immutable = mk()
	println(v.Field, w.Field)
}

// Error:
// illegal conversion to external realm type (call a function of the realm that declares the type to make it)

// TypeCheckError:
// gno.land/r/launderrh_conv_external_multidecl/zrealm_launder_rdata_conv_external_multidecl.gno:15:6: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_external_multidecl/zrealm_launder_rdata_conv_external_multidecl.gno:15:9: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer)
