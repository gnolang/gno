// PKGPATH: gno.land/r/demo/ingressuser
package ingressuser

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/r/tests/vm/viewdefault"
)

// Same rule in the argument direction: a tree this realm hands to another
// realm is a view there, whether it was stored earlier (t) or allocated in
// this transaction (fresh). Storing it hands no write power.

var t = avl.NewTree()

func main(cur realm) {
	viewdefault.Store(cross(cur), t)
	println("ungranted write by holder:", revive(func() { viewdefault.WriteInbox(cross(cur)) }))
	println("nothing landed:", t.Get("from") == nil)

	fresh := avl.NewTree()
	println("ungranted fresh tree:", revive(func() { viewdefault.StoreAndWrite(cross(cur), fresh) }))
	println("fresh empty:", fresh.Get("from") == nil)
}

// Output:
// ungranted write by holder: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// nothing landed: true
// ungranted fresh tree: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// fresh empty: true
