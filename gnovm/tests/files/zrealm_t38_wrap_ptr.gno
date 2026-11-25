// PKGPATH: gno.land/r/zrealm_t38_wrap_ptr
// The victim hands a pointer to its own fresh writer to a caller, which returns
// it: the pointer left the victim's realm, so the writer behind it is shared
// and the victim's call through it is refused, as for a by-value return.

package zrealm_t38_wrap_ptr

import "gno.land/r/tests/vm/t35c/victim"

type W struct{}

func (w *W) WrapPtr(p *func()) *func() { return p }

func main(cur realm) {
	println("same:", revive(func() { victim.RunWrappedPtr(cross(cur), &W{}) }))
	println("count:", victim.Count())
}

// Output:
// same: gno.land/r/tests/vm/t35c/victim's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// count: 0
