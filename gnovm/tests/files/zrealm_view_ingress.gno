// PKGPATH: gno.land/r/demo/ingressuser
package ingressuser

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/r/tests/vm/viewdefault"
)

// Same rule in the argument direction: a tree this realm hands to another
// realm is a view there unless this realm granted it.

var t = avl.NewTree()

func main(cur realm) {
	viewdefault.Store(cross(cur), t)
	println("ungranted write by holder:", revive(func() { viewdefault.WriteInbox(cross(cur)) }))
	viewdefault.Store(cross(cur), mutable(t))
	viewdefault.WriteInbox(cross(cur))
	println("granted write landed here:", t.Get("from"))
}

// Output:
// ungranted write by holder: cannot modify object owned by another realm from library code (the owner must hand it out with mutable(x), or expose a crossing function): tree<VPBlock(1,0)>.node
// granted write landed here: viewdefault
