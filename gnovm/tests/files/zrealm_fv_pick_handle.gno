// PKGPATH: gno.land/r/zrealm_fv_pick_handle
// Control: the picker returns a handle of a declared victim writer
// after a nested call (handles are stamped view-only, never recorded).

package zrealm_fv_pick_handle

import "gno.land/r/tests/vm/pickvault/victim"

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
// gno.land/r/tests/vm/pickvault/victim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): vault<~VPBlock(3,1)>.N
