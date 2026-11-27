// PKGPATH: gno.land/r/zrealm_fv_pick_keep
// The victim stores the picked (leaked-id) method value, then a second
// crossing call in the same tx runs it. Must not write.

package zrealm_fv_pick_keep

import "gno.land/r/tests/vm/pickvault/victim"

type P struct{}

func (p *P) Pick() func() { return victim.Get().Bump }

func main(cur realm) {
	victim.Keep(cross(cur), &P{})
	victim.RunKept(cross(cur))
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/pickvault/victim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
