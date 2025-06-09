package validators

import (
	"gno.land/p/nt/avl"
	"gno.land/p/nt/poa"
)

func init() {
	// The default valset protocol is PoA
	vp = poa.NewPoA()

	// No changes to apply initially
	changes = avl.NewTree()
}
