package validators

import (
	"gno.land/p/nt/poa"
	"gno.land/p/sys/validators"
)

func init() {
	// The default valset protocol is PoA
	vp = poa.NewPoA()

	// No changes to apply initially
	changes = make([]validators.Validator, 0)
}
