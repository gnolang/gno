// This package is used to manage the validator set.
package validators

import (
	"std"

	"gno.land/p/demo/avl"
)

var (
	validators = avl.NewTree() // addr -> *Validator
)

func validatorKey(addr std.Address) string {
	return addr.String()
}
