package validators

import (
	"std"

	"gno.land/p/nt/poa"
)

func init() {
	// The default valset protocol is PoA
	vp = poa.NewPoA()

	// Explicitly set the initial set
	std.SetParamStrings(valsetPrevKey, []string{})
}
