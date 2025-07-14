package protocol_fee

import "gno.land/r/gnoswap/v1/access"

var (
	protocolFeeAddr, _ = access.GetAddress(access.ROLE_PROTOCOL_FEE)
	govStakerAddr, _   = access.GetAddress(access.ROLE_GOV_STAKER)
	devOpsAddr, _      = access.GetAddress(access.ROLE_DEVOPS)
)
