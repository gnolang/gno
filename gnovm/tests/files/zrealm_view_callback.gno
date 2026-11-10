// PKGPATH: gno.land/r/demo/viewcallback
package viewcallback

import "gno.land/r/tests/vm/viewdefault"

// A method value acts for the realm that bound it. v.Set bound here on a view
// stays refused when a granted handle's Iterate calls it back under the
// owner's storage; h.Set bound here on the handle still commits; a method
// value the owner bound and handed out writes as the owner.

func main(cur realm) {
	h := viewdefault.GetHandle()
	v := viewdefault.GetView()
	h.Set("a", "from-caller")
	// The refusal unwinds through the owner's borrowed frames: revive, not recover.
	println("view callback:", revive(func() { h.Iterate("", "", v.Set) }))
	println("view:", v.Get("k"), v.Get("a"), v.Size())
	v.Iterate("", "", h.Set)
	println("handle:", h.Get("k"), h.Size())
	// Bound by the owner and handed out: writes the view as the owner.
	viewdefault.ViewSetter()("o", "owner-bound")
	println("owner-bound:", v.Get("o"))
}

// Output:
// view callback: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
// view: v0 undefined 1
// handle: v0 2
// owner-bound: owner-bound
