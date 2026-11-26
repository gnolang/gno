// PKGPATH: gno.land/r/zrealm_fv_return_own
// The victim hands its writer to a caller's wrapper, which returns it
// unchanged (R2b: a result another realm hands over has no maker). Must not write.

package zrealm_fv_return_own

import "gno.land/r/tests/vm/pickvault/victim"

type W struct{}

func (w *W) Wrap(bump func()) func() { return bump }

func main(cur realm) {
	victim.RunWrapped(cross(cur), &W{})
	println("count:", victim.Count())
}

// Error:
// gno.land/r/tests/vm/pickvault/victim's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): v<VPBlock(1,0)>.N
