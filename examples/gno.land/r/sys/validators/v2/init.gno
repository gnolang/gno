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

	// No realms are allowed to modify the validator set initially
	allowedRealms = avl.NewTree()
}
