// PKGPATH: gno.land/r/zrealm_t38_wrap_ptr_call
// The victim's fresh writer, handed out behind a pointer, is called by the
// caller's code inside the victim's call: that code is not in the victim's
// import tree, so it runs read-only (R2c) and the write is refused.
package zrealm_t38_wrap_ptr_call

import "gno.land/r/tests/vm/t35c/victim"

type W struct{}

func (w *W) WrapPtr(p *func()) *func() { (*p)(); return p }

func main(cur realm) {
	println("same:", revive(func() { victim.RunWrappedPtr(cross(cur), &W{}) }))
	println("count:", victim.Count())
}

// Output:
// same: gno.land/r/tests/vm/t35c/victim's code was entered without crossing and is read-only (call its crossing function): v<VPBlock(1,0)>.N
// count: 0
