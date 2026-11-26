// PKGPATH: gno.land/r/demo/entrymethodvalue
package entrymethodvalue

import "gno.land/r/tests/vm/viewdefault"

// The owner's view.Set handed out as a method value carries no authority:
// called by a stranger it runs view-only. The owner's own entry still writes
// the view.

func main(cur realm) {
	println("stranger call:", try(func() { viewdefault.ViewSetter()("k", "pwned") }))
	println("view:", viewdefault.Read("k"))
	viewdefault.SetView(cross(cur), "k", "v1")
	println("owner wrote:", viewdefault.Read("k"))
}

// try runs fn and returns what it panicked with: a refused write in a
// view-only frame is recovered by the caller, not revived.
func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// stranger call: cannot modify object owned by another realm from library code (call a crossing function of the owning realm): tree<VPBlock(1,0)>.node
// view: v0
// owner wrote: v1
