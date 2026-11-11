// PKGPATH: gno.land/r/demo/viewcallback
package viewcallback

import "gno.land/r/tests/vm/viewdefault"

// A method value of a /p/ method writes an ungranted receiver only when the
// realm that bound it and the realm whose code calls it both own the receiver.

// caught recovers a refusal raised with no borrowed frame in between.
func caught(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

func main(cur realm) {
	h := viewdefault.GetHandle()
	v := viewdefault.GetView()
	h.Set("a", "from-caller")
	// The refusal unwinds through the owner's borrowed frames: revive, not recover.
	println("view callback:", revive(func() { h.Iterate("", "", v.Set) }))
	// Same through a defer statement inside the owner's granted handle.
	println("deferred callback:", revive(func() { viewdefault.NewViewWriter().Deferred(v.Set) }))
	println("view:", v.Get("k"), v.Get("a"), v.Get("d"), v.Size())
	v.Iterate("", "", h.Set)
	println("handle:", h.Get("k"), h.Size())
	// Bound by the owner on its ungranted view and handed out: still a view,
	// called directly or from inside the owner's granted handle.
	println("owner-bound:", caught(func() { viewdefault.ViewSetter()("o", "x") }))
	println("owner-bound callback:", revive(func() { h.Iterate("", "", viewdefault.ViewSetter()) }))
	println("view:", v.Get("o"), v.Get("a"), v.Size())
	// On a granted object it writes; the owner's own code runs its own.
	viewdefault.HandleSetter()("g", "granted")
	println("granted:", h.Get("g"))
	viewdefault.CopyToView(cross(cur))
	println("owner's own:", v.Get("a"), v.Get("g"), v.Size())
}

// Output:
// view callback: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
// deferred callback: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
// view: v0 undefined undefined 1
// handle: v0 2
// owner-bound: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
// owner-bound callback: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
// view: undefined undefined 1
// granted: granted
// owner's own: from-caller granted 3
