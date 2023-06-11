// This package is used to manage the validator set.
package validators

import (
	"std"

	"gno.land/p/demo/avl"
)

var (
	validators = avl.NewTree() // addr -> *Validator
)

var (
	maxValidators = 20
	nativeDenom   = "ugnot"
	admin         = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
)

func validatorKey(addr std.Address) string {
	return addr.String()
}
