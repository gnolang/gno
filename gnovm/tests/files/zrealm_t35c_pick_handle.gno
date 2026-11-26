// PKGPATH: gno.land/r/zrealm_t35c_pick_handle
// T35c control: the picker returns a handle of a declared victim writer
// after a nested call (handles are stamped view-only, never recorded).

package zrealm_t35c_pick_handle

import "gno.land/r/tests/vm/t35c/victim"

type P struct{}

func (p *P) Pick() func() {
	victim.Get()
	return victim.BumpFn
}

func main(cur realm) {
	victim.RunPicked(cross(cur), &P{})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/t35c/victim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): vault<~VPBlock(3,1)>.N
