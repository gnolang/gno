// PKGPATH: gno.land/r/demo/viewcallback
package viewcallback

import (
	"gno.land/p/demo/tests/viewgrant/v0"
	"gno.land/r/tests/vm/viewdefault"
)

// A method value of a /p/ method writes a stored object only when the realm
// that bound it calls it from its own code; a foreign caller gets no write.

func main(cur realm) {
	h := viewdefault.GetHandle()
	v := viewdefault.GetView()
	viewdefault.HandleSet(cross(cur), "k", "v0")
	viewdefault.HandleSet(cross(cur), "a", "from-caller")
	viewdefault.HandleSet(cross(cur), "g", "granted")
	// The owner's library code calls the caller's method value: no write.
	println("view callback:", revive(func() { viewdefault.HandleIterate(cross(cur), v.Set) }))
	// Same through a defer statement inside the owner's code.
	println("deferred callback:", revive(func() { viewdefault.CallDeferred(cross(cur), v.Set) }))
	// A stranger's own code runs the owner's method value through h, which
	// only reads h: no write, and the view is unchanged below.
	println("stranger callback:", try(func() { h.Iterate("", "", viewdefault.ViewSetter()) }))
	println("view:", v.Get("k"), v.Get("a"), v.Get("d"), v.Size())
	// Bound by the owner on its view and called by a stranger: no write.
	println("owner-bound:", try(func() { viewdefault.ViewSetter()("o", "x") }))
	println("owner-bound hook:", revive(func() { viewdefault.RunHook(cross(cur), viewgrant.Wrap(viewdefault.ViewSetter())) }))
	// A method value the owner bound outside a gate has no binder: its gate writes nothing.
	println("owner-bound callback:", revive(func() { viewdefault.HandleIterate(cross(cur), viewdefault.ViewSetter()) }))
	println("view:", v.Get("k"), v.Get("a"), v.Get("o"), v.Size())
	viewdefault.CopyToView(cross(cur))
	println("owner's own:", v.Get("own"))
	println("handle:", h.Get("k"), h.Size())
}

// try runs fn and returns what it panicked with: a refused write in a
// view-only frame is recovered by the caller, not revived.
func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// view callback: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// deferred callback: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// stranger callback: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// view: v0 undefined undefined 1
// owner-bound: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// owner-bound hook: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// owner-bound callback: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// view: v0 undefined undefined 1
// owner's own: copy
// handle: v0 3
