// PKGPATH: gno.land/r/zrealm_fv_own_any
// Control: a method value the attacker bound in its own crossing call,
// handed to the victim's crossing call (any). Must not write.

package zrealm_fv_own_any

import "gno.land/r/tests/vm/pickvault/victim"

func main(cur realm) {
	f := victim.Get().Bump
	victim.RunAny(cross(cur), f)
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/pickvault/victim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
