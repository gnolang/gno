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
// illegal conversion to external realm type
