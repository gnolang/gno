package emission

import (
	"gno.land/r/gnoswap/v1/access"
)

var (
	adminAddr, _         = access.GetAddress(access.ROLE_ADMIN)
	stakerAddr, _        = access.GetAddress(access.ROLE_STAKER)
	govStakerAddr, _     = access.GetAddress(access.ROLE_GOV_STAKER)
	communityPoolAddr, _ = access.GetAddress(access.ROLE_COMMUNITY_POOL)
	emissionAddr, _      = access.GetAddress(access.ROLE_EMISSION)
	devOpsAddr, _        = access.GetAddress(access.ROLE_DEVOPS)
)
