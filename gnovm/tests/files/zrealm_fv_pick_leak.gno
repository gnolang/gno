// PKGPATH: gno.land/r/zrealm_fv_pick_leak
// A picker the victim's crossing call runs read-only returns a method
// value of the victim's vault, bound after a nested call returned (the return
// restores the crossing id into the read-only stretch). Must not write.

package zrealm_fv_pick_leak

import "gno.land/r/tests/vm/pickvault/victim"

type P struct{}

func helper() int { return 1 }

func (p *P) Pick() func() {
	helper()
	return victim.Get().Bump
}

func main(cur realm) {
	victim.RunPicked(cross(cur), &P{})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/pickvault/victim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
