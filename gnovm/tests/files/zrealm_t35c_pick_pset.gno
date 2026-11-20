// PKGPATH: gno.land/r/zrealm_t35c_pick_pset
// T35c: as pick_getter, the picked value being the /p/ avl Set method bound
// on the victim's stored tree. Must not write.

package zrealm_t35c_pick_pset

import "gno.land/r/tests/vm/t35c/victim"

type P struct{}

func (p *P) PickSet() func(string, any) bool { return victim.Tree().Set }

func main(cur realm) {
	victim.RunSet(cross(cur), &P{})
	println("size:", victim.Size())
}

// Error:
// cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
