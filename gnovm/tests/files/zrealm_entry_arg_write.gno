// PKGPATH: gno.land/r/demo/entryargwrite
package entryargwrite

import (
	"gno.land/p/demo/tests/viewgrant/v0"
	"gno.land/r/tests/vm/entries"
)

// A /p/ method that writes its argument writes only for the realm that owns
// the argument: CopyTo called by a stranger on its own Pair or on the
// victim's own Pair, with the victim's B, is refused and B stays unchanged.

func main(cur realm) {
	a := viewgrant.NewPair()
	func() {
		defer func() { println("copy to victim:", recover()) }()
		a.CopyTo(entries.GetB())
	}()
	println("B unchanged:", entries.ReadBFrom() == "")
	func() {
		defer func() { println("victim A:", recover()) }()
		entries.GetPair().CopyTo(entries.GetB())
	}()
	println("victim A: B unchanged:", entries.ReadBFrom() == "")
	entries.OwnerSetB(cross(cur), "v1")
	println("owner wrote B:", entries.ReadB())
}

// Output:
// copy to victim: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// B unchanged: true
// victim A: cannot modify object owned by another realm from library code (call a function of the owning realm): tree<VPBlock(1,0)>.node
// victim A: B unchanged: true
// owner wrote B: v1
