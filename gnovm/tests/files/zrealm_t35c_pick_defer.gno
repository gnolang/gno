// PKGPATH: gno.land/r/zrealm_t35c_pick_defer
// T35c: as pick_getter, the victim defers the picked callback.

package zrealm_t35c_pick_defer

import "gno.land/r/tests/vm/t35c/victim"

type P struct{}

func (p *P) Pick() func() { return victim.Get().Bump }

func main(cur realm) {
	victim.DeferPicked(cross(cur), &P{})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/t35c/victim's code is read-only here: a function value writes only in the crossing call that made it: v<VPBlock(1,0)>.N
