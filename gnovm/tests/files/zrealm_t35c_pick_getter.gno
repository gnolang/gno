// PKGPATH: gno.land/r/zrealm_t35c_pick_getter
// T35c: as pick_leak, the nested call being the victim's own getter.

package zrealm_t35c_pick_getter

import "gno.land/r/tests/vm/t35c/victim"

type P struct{}

func (p *P) Pick() func() { return victim.Get().Bump }

func main(cur realm) {
	victim.RunPicked(cross(cur), &P{})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/t35c/victim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
