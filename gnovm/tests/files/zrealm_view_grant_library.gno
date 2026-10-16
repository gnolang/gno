// PKGPATH: gno.land/r/demo/grantlibrary
package grantlibrary

import "gno.land/r/tests/vm/viewdefault"

// /p/ code cannot grant a realm's stored object on its behalf, but a /p/
// constructor can grant the object it hands out, and that handle writes
// the owner's stored state under the owner's authority (the teller shape).

func main(cur realm) {
	println("library grant on stored view:", revive(func() { viewdefault.PublishView() }))
	func() {
		defer func() { println("view still refuses:", recover()) }()
		viewdefault.GetView().Set("k", "pwned") // same-realm panic: recover, not revive
	}()
	viewdefault.NewViewWriter().Set("k", "via-writer")
	println("owner sees:", viewdefault.Read("k"))
}

// Output:
// library grant on stored view: mutable: library code cannot grant a stored object; the owning realm must call mutable(x) itself
// view still refuses: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
// owner sees: via-writer
