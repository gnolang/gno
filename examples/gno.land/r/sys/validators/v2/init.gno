package validators

import (
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/poa/v0"
)

func init() {
	// The default valset protocol is PoA
	vp = poa.NewPoA()

	// No changes to apply initially
	changes = bptree.NewBPTree32()
}
