package access

import (
	"std"

	"gno.land/p/gnoswap/consts"
)

// Pre-defined role kinds (follows the each contract's name)
const (
	ROLE_ADMIN          = "admin"
	ROLE_COMMUNITY_POOL = "community_pool"
	ROLE_DEVOPS         = "devops"
	ROLE_GOVERNANCE     = "governance"
	ROLE_GOV_STAKER     = "gov_staker"
	ROLE_ROUTER         = "router"
	ROLE_POOL           = "pool"
	ROLE_POSITION       = "position"
	ROLE_STAKER         = "staker"
	ROLE_LAUNCHPAD      = "launchpad"
	ROLE_EMISSION       = "emission"
	ROLE_PROTOCOL_FEE   = "protocol_fee"
	ROLE_GOV_XGNS       = "xgns"

	PERM_ACCESS = "access" // Permission name for access control
)

// initializeDefaultRoles initializes the default roles with their corresponding addresses from consts
// dependent with the gnoswap contracts.
func initializeDefaultRoles() (*config, error) {
	cfg := newConfig()

	// Set up roles with their corresponding addresses from consts
	//
	// TODO: Make sure to initialize the map with JSON or
	// another appropriate method instead of const initially.
	defaultRoles := map[string]std.Address{
		ROLE_ADMIN:          consts.ADMIN,
		ROLE_COMMUNITY_POOL: consts.COMMUNITY_POOL_ADDR,
		ROLE_DEVOPS:         consts.DEV_OPS,
		ROLE_GOVERNANCE:     consts.GOV_GOVERNANCE_ADDR,
		ROLE_GOV_STAKER:     consts.GOV_STAKER_ADDR,
		ROLE_ROUTER:         consts.ROUTER_ADDR,
		ROLE_POOL:           consts.POOL_ADDR,
		ROLE_POSITION:       consts.POSITION_ADDR,
		ROLE_STAKER:         consts.STAKER_ADDR,
		ROLE_LAUNCHPAD:      consts.LAUNCHPAD_ADDR,
		ROLE_EMISSION:       consts.EMISSION_ADDR,
		ROLE_PROTOCOL_FEE:   consts.PROTOCOL_FEE_ADDR,
		ROLE_GOV_XGNS:       consts.GOV_XGNS_ADDR,
	}

	cfg.roles = defaultRoles
	return cfg, initialize(cfg)
}
