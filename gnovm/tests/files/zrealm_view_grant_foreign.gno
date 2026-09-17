// PKGPATH: gno.land/r/demo/grantforeign
package grantforeign

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/r/tests/vm/viewdefault"
)

// Only the owner may grant: this realm cannot mutable() viewdefault's tree
// (ordinary panic, recover), and viewdefault cannot mutable() a tree this
// realm allocated (abort across the crossing boundary, revive).

var mine = avl.NewTree()

func main(cur realm) {
	func() {
		defer func() { println("by caller:", recover()) }()
		mutable(viewdefault.GetView())
	}()
	println("by callee:", revive(func() { viewdefault.GrantForeign(cross(cur), mine) }))
}

// Output:
// by caller: mutable: only the owning realm may grant write access to its object
// by callee: mutable: only the owning realm may grant write access to its object
