// PKGPATH: gno.land/r/zrealm_fv_wrap_ptr
// The victim hands a pointer to its own fresh writer to a caller, which returns
// it: the pointer left the victim's realm, so the writer behind it is shared
// and the victim's call through it is refused, as for a by-value return.

package zrealm_fv_wrap_ptr

import "gno.land/r/tests/vm/pickvault/victim"

type W struct{}

func (w *W) WrapPtr(p *func()) *func() { return p }

func main(cur realm) {
	println("same:", revive(func() { victim.RunWrappedPtr(cross(cur), &W{}) }))
	println("count:", victim.Count())
}

// Output:
// same: gno.land/r/tests/vm/pickvault/victim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
// count: 0
