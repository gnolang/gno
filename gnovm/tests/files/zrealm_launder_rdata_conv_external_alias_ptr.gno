// PKGPATH: gno.land/r/launderrh_conv_external_alias_ptr
package launderrh_conv_external_alias_ptr

import "gno.land/r/tests/vm/launderrvictim"

type Local struct {
	Field string
}

type P = *launderrvictim.Immutable

func main(cur realm) {
	local := &Local{Field: "xxx"}
	_ = P(local)
}

// Error:
// illegal conversion to external realm type

// TypeCheckError:
// gno.land/r/launderrh_conv_external_alias_ptr/zrealm_launder_rdata_conv_external_alias_ptr.gno:14:6: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer)
