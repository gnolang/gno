// PKGPATH: gno.land/r/zrealm_fv_pick_defer
// As pick_getter, the victim defers the picked callback.

package zrealm_fv_pick_defer

import "gno.land/r/tests/vm/pickvault/victim"

type P struct{}

func (p *P) Pick() func() { return victim.Get().Bump }

func main(cur realm) {
	victim.DeferPicked(cross(cur), &P{})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/pickvault/victim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
