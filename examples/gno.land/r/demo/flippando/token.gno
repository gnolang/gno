package flippando

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc20"
)

var (
	flipToken = grc20.NewAdminToken("Flippando", "FLIP", 6)
	lockedSupply int64
)

