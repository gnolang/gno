// PKGPATH: gno.land/r/zrealm_t35c_pick_held
// T35c control: the picker binds the method with no nested call first.

package zrealm_t35c_pick_held

import "gno.land/r/tests/vm/t35c/victim"

type P struct{ v *victim.Vault }

func (p *P) Pick() func() { return p.v.Bump }

func main(cur realm) {
	victim.RunPicked(cross(cur), &P{v: victim.Get()})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/t35c/victim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
