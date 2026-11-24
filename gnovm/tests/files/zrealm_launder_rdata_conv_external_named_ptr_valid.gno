// PKGPATH: gno.land/r/launderrh_conv_external_named_ptr_valid
package launderrh_conv_external_named_ptr_valid

import "gno.land/r/tests/vm/launderrvictim"

type LocalPtr *launderrvictim.Immutable

func isNil(p *launderrvictim.Immutable) bool { return p == nil }

// A nil pointer points at nothing, so converting it through a local
// named pointer type and back forges nothing and must be allowed.
func main(cur realm) {
	var ptr *launderrvictim.Immutable
	local := LocalPtr(ptr)
	println(local == nil)

	back := (*launderrvictim.Immutable)(local)
	println(back == nil)

	var implicit *launderrvictim.Immutable = local
	println(implicit == nil)

	println(isNil(local))
}

// Output:
// true
// true
// true
// true
