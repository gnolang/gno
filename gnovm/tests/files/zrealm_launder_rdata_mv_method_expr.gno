// PKGPATH: gno.land/r/launderrh_mv_method_expr
package launderrh_mv_method_expr

// Method-EXPRESSION laundering probe (unbound).
//
// `me := (*launderpkg.Immutable).Apply` is the unbound method value
// — recv is supplied at call time as the first arg. The body of the
// method is /p/launderpkg-declared (no borrow rule #1 from /r/). When
// invoked as me(victimImmPtr, EvilWrite), PushFrameCall's receiver-
// borrow rule fires on the FIRST arg (the recv), which is /r/-stamped
// → m.Realm = /r/launderrvictim → EvilWrite (top-level /p/) inherits
// → SUCCEEDS.
//
// Empirically observed: the unbound method expression DOES NOT
// trigger receiver-borrow at call time the same way the bound
// method value does. The write fires the readonly check and panics.
// This is a divergence vs. the bound-method-value variant
// (zrealm_launder_rdata_mv_method_value.gno), which succeeds.

import (
	"gno.land/p/demo/tests/launderattack"
	"gno.land/p/demo/tests/launderpkg"
	"gno.land/r/tests/vm/launderrvictim"
)

func main(cur realm) {
	println("before:", launderrvictim.ReadEmbed())

	we := launderrvictim.GetWithEmbed()
	// Unbound method expression on /p/-pointer-receiver Apply.
	me := (*launderpkg.Immutable).Apply
	// Explicit recv (&we.Immutable) supplied as first argument.
	me(&we.Immutable, launderattack.EvilWrite)

	println("after: ", launderrvictim.ReadEmbed())
}

// Error:
// cannot directly modify readonly tainted object (use a method or crossing function): i<VPBlock(1,0)>.Field
