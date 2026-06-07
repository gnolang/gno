// PKGPATH: gno.land/r/launderrh_conv_ptr_chain
package launderrh_conv_ptr_chain

// CONVERSION VECTOR (composite + pun chain):
// Reach a *launderpkg.Bare inside a /r/-victim ptr-slice element,
// then convert it to *launderattack.Tamper.
//
// Question: does Case 1 fire when the source pointer was obtained via
// `slice[i]` from a foreign slice (rather than directly from a
// getter return)?
//
// Hypothesis: Case 1 fires — readonly taint propagates through
// container access.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	ps := launderrvictim.GetBarePtrSlice()
	innerPtr := ps[0] // *launderpkg.Bare from foreign slice
	t := (*launderattack.Tamper)(innerPtr)
	t.Set("pwnd-via-ptr-chain")
	println("after:", launderrvictim.ReadBarePtrSlice0())
}

// Error:
// illegal conversion of readonly or externally stored value
