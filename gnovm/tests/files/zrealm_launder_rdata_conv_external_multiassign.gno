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
// illegal conversion to external realm type
