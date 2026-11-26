// PKGPATH: gno.land/r/zrealm_t35c_return_own
// T35c: the victim hands its writer to a caller's wrapper, which returns it
// unchanged (R2b: a result another realm hands over has no maker). Must not write.

package zrealm_t35c_return_own

import "gno.land/r/tests/vm/t35c/victim"

type W struct{}

func (w *W) Wrap(bump func()) func() { return bump }

func main(cur realm) {
	victim.RunWrapped(cross(cur), &W{})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/t35c/victim's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
