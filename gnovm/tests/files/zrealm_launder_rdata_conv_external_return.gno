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
// illegal conversion to external realm type
