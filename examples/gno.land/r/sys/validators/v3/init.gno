package validators

import (
	"chain/params"

	"gno.land/p/nt/poa/v0"
)

func init() {
	// The default valset protocol is PoA.
	vp = poa.NewPoA()

	// Explicitly initialize the previous valset as empty.
	params.SetStrings(valsetPrevKey, []string{})
}
