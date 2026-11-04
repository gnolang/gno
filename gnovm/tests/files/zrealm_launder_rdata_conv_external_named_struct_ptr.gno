// PKGPATH: gno.land/r/launderrh_conv_external_named_struct_ptr
package launderrh_conv_external_named_struct_ptr

import "gno.land/r/tests/vm/launderrvictim"

type Local launderrvictim.Immutable

func main(cur realm) {
	local := &Local{Field: "xxx"}
	_ = (*launderrvictim.Immutable)(local)
}

// Error:
// illegal conversion to external realm type
