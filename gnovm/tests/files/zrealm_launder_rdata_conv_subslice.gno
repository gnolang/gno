// PKGPATH: gno.land/r/launderrh_conv_subslice
package launderrh_conv_subslice

// CONVERSION VECTOR (compound subslice + pun):
// Take a subslice of victim's []Bare, then convert each element's
// address to *Tamper. Confirms Case 1 fires after subslicing.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	s := launderrvictim.GetBareSlice()
	s2 := s[0:1]
	innerPtr := &s2[0]
	t := (*launderattack.Tamper)(innerPtr)
	t.Set("pwnd-via-subslice-pun")
	println("after:", launderrvictim.ReadBareSlice0())
}

// Error:
// illegal conversion of readonly or externally stored value
