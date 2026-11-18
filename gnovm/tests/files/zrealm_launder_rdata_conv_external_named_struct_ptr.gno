// PKGPATH: gno.land/r/launderrh_conv_external_named_struct_ptr
package launderrh_conv_external_named_struct_ptr

import "gno.land/r/tests/vm/launderrvictim"

type Local launderrvictim.Immutable

func main(cur realm) {
	local := &Local{Field: "xxx"}
	_ = (*launderrvictim.Immutable)(local)
}

// Error:
// cannot allocate struct{Field string} (declared in gno.land/r/tests/vm/launderrvictim) in realm gno.land/r/launderrh_conv_external_named_struct_ptr

// TypeCheckError:
// gno.land/r/launderrh_conv_external_named_struct_ptr/zrealm_launder_rdata_conv_external_named_struct_ptr.gno:6:6: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer); gno.land/r/launderrh_conv_external_named_struct_ptr/zrealm_launder_rdata_conv_external_named_struct_ptr.gno:10:6: cannot hold a value of gno.land/r/tests/vm/launderrvictim.Immutable outside its realm (use a pointer)
