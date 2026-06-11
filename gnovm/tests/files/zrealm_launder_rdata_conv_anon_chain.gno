// PKGPATH: gno.land/r/launderrh_conv_anon_chain
package launderrh_conv_anon_chain

// CONVERSION VECTOR (anon -> named -> attack):
// Try chaining conversions through an anonymous-struct intermediary:
//   *launderpkg.Bare  ->  *struct{Field string}  ->  *launderattack.Tamper
// If Case 1 fires only on the FIRST conversion, this 2-step chain
// would still be blocked at step 1. If Case 1 somehow lets the
// anonymous-target step through (it doesn't, per conv_anon_struct.gno),
// the second step might succeed since the intermediate value is no
// longer "foreign-readonly" once it crosses into an anonymous type.
//
// Hypothesis: Case 1 fires at step 1.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	w := launderrvictim.GetWithBarePtr()
	anon := (*struct{ Field string })(w.Inner)
	t := (*launderattack.Tamper)(anon)
	t.Set("pwnd-via-anon-chain")
	println("after:", launderrvictim.ReadBarePtr())
}

// Error:
// illegal conversion of readonly or externally stored value
