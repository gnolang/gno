package config

import (
	"std"

	"gno.land/p/demo/ownable"
)

// Owner is an instance of the safe-object ownable.Owner
var Owner = ownable.NewWithAddress(std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"))
