// PKGPATH: gno.land/r/zrealm_t35c_pick_keep
// T35c: the victim stores the picked (leaked-id) method value, then a second
// crossing call in the same tx runs it. Must not write.

package zrealm_t35c_pick_keep

import "gno.land/r/tests/vm/t35c/victim"

type P struct{}

func (p *P) Pick() func() { return victim.Get().Bump }

func main(cur realm) {
	victim.Keep(cross(cur), &P{})
	victim.RunKept(cross(cur))
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/t35c/victim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
