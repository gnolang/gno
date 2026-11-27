// PKGPATH: gno.land/r/launderrh_conv_external_multidecl_pkg
package launderrh_conv_external_multidecl_pkg

import "gno.land/r/tests/vm/launderrvictim"

// The package-level form of zrealm_launder_rdata_conv_external_multidecl.gno:
// the declaration runs at package init and is refused there.

func mk() (struct{ Field string }, struct{ Field string }) {
	return struct{ Field string }{"xxx"}, struct{ Field string }{"yyy"}
}

var v, w launderrvictim.Immutable = mk()

func main(cur realm) {
	println(v.Field, w.Field)
}

// Error:
// illegal conversion to external realm type

// TypeCheckError:
// gno.land/r/launderrh_conv_external_multidecl_pkg/zrealm_launder_rdata_conv_external_multidecl_pkg.gno:13:5: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_external_multidecl_pkg/zrealm_launder_rdata_conv_external_multidecl_pkg.gno:13:8: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer)
