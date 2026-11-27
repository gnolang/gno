// PKGPATH: gno.land/r/zrealm_fv_pick_getter
// As pick_leak, the nested call being the victim's own getter.

package zrealm_fv_pick_getter

import "gno.land/r/tests/vm/pickvault/victim"

type P struct{}

func (p *P) Pick() func() { return victim.Get().Bump }

func main(cur realm) {
	victim.RunPicked(cross(cur), &P{})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/pickvault/victim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
