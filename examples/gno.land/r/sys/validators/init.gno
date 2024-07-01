package validators

import (
	"gno.land/p/nt/poa"
	"gno.land/p/sys/validators"
)

func init() {
	v = &valset{
		p:       poa.NewPoA(),
		changes: make([]validators.Validator, 0),
	}
}
