// PKGPATH: gno.land/r/zrealm_t35c_own_slice
// T35c control: a method value the attacker bound in its own crossing call,
// handed to the victim's crossing call (slice). Must not write.

package zrealm_t35c_own_slice

import "gno.land/r/tests/vm/t35c/victim"

func main(cur realm) {
	f := victim.Get().Bump
	victim.RunAll(cross(cur), []func(){f})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/t35c/victim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
