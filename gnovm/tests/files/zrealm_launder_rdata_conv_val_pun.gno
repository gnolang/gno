// PKGPATH: gno.land/r/launderrh_conv_val_pun
package launderrh_conv_val_pun

// CONVERSION VECTOR (struct value, not pointer):
// Convert a foreign /p/-typed VALUE (not pointer) to a parallel
// /p/-type.
//
// Post-#5747: the value-copy out of victim's slice produces a fresh
// realm-local struct. The conversion to Tamper succeeds (no aliasing
// to victim), and the local Set mutates only the local copy —
// victim's slice is unchanged. The defense against actual hijacks
// (aliased pointer conversion) lives in the doOpConvert path and is
// exercised by conv_val_then_addr.

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

// Output:
// local tv.Field: local-pwnd
// victim slice[0]: slice0
