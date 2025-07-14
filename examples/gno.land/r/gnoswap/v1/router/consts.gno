package router

import (
	"gno.land/r/gnoswap/v1/access"
)

var (
	poolAddr, _        = access.GetAddress(access.ROLE_POOL)
	routerAddr, _      = access.GetAddress(access.ROLE_ROUTER)
	positionAddr, _    = access.GetAddress(access.ROLE_POSITION)
	protocolFeeAddr, _ = access.GetAddress(access.ROLE_PROTOCOL_FEE)
)
