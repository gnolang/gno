package staker

import (
	"gno.land/r/gnoswap/v1/access"
)

var (
	communityPoolAddr, _ = access.GetAddress(access.ROLE_COMMUNITY_POOL)
	stakerAddr, _        = access.GetAddress(access.ROLE_STAKER)
	positionAddr, _      = access.GetAddress(access.ROLE_POSITION)
	protocolFeeAddr, _   = access.GetAddress(access.ROLE_PROTOCOL_FEE)
	devOpsAddr, _        = access.GetAddress(access.ROLE_DEVOPS)
)
