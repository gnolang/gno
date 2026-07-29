// PKGPATH: gno.land/r/launderrh_conv_bare_pun
package launderrh_conv_bare_pun

// CONVERSION VECTOR 1: foreign /p/-typed pointer (no methods) converted
// to a parallel /p/-type (same struct layout) that HAS a mutator method.
//
// Original Attack G shape, but with /p/-Bare embedded/fielded in /r/-data.
// The Bare type has no Set method; the attacker would gain one via
// (*launderattack.Tamper)(barePtr) and invoke Set under victim authority.
//
// Hypothesis: doOpConvert Case 1 (foreign-readonly source) fires at the
// conversion site and blocks the attack before any method dispatch.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	w := launderrvictim.GetWithBarePtr()
	// w.Inner is *launderpkg.Bare (foreign-readonly).
	// launderattack.Tamper has the SAME shape {Field string} plus a Set method.
	tamper := (*launderattack.Tamper)(w.Inner)
	tamper.Set("pwnd-via-bare-pun")
	println("after:", launderrvictim.ReadBarePtr())
}

// Error:
// illegal conversion of readonly or externally stored value
