// PKGPATH: gno.land/r/zrealm_t38_wrap_ptr
// The victim hands a pointer to its own fresh writer to a caller, which returns
// it: R2b strips only values, so the victim's call through it writes (count 1).

package zrealm_t38_wrap_ptr

import "gno.land/r/tests/vm/t35c/victim"

type W struct{}

func (w *W) WrapPtr(p *func()) *func() { return p }

func main(cur realm) {
	println("same:", revive(func() { victim.RunWrappedPtr(cross(cur), &W{}) }))
	println("count:", victim.Count())
}

// Output:
// same: undefined
// count: 1
