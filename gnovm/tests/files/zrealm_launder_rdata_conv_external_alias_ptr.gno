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
