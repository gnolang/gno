// PKGPATH: gno.land/r/launderrh_conv_val_pun
package launderrh_conv_val_pun

// CONVERSION VECTOR (struct value, not pointer):
// Convert a foreign /p/-typed VALUE (not pointer) to a parallel
// /p/-type. Even if Case 1 allows the conversion when source IS a
// pure local copy (no taint), this attack relies on aliasing — which
// a value copy breaks anyway. But what if the value retains taint
// (which we saw in conv_val_then_addr)? Then Case 1 fires.
//
// This test exercises the Tamper-conversion on a VALUE retrieved
// from victim's slice (no pointer aliasing). Even if conversion is
// allowed, the local Tamper value's Set would only mutate the local
// copy. We test both outcomes.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	bv := launderrvictim.GetBareSlice()[0] // value copy
	// Convert Bare-value to launderattack.Tamper-value.
	tv := launderattack.Tamper(bv)
	(&tv).Set("local-pwnd")
	println("local tv.Field:", tv.Field)
	println("victim slice[0]:", launderrvictim.ReadBareSlice0())
}

// Error:
// illegal conversion of readonly or externally stored value
