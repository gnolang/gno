package access

import (
	"std"

	"gno.land/r/gnoswap/v1/rbac"
)

// checkRole checks if the caller has the given role and optionally updates the role address
func checkRole(roleName string, caller std.Address) error {
	return rbac.CheckPermission(roleName, PERM_ACCESS, caller)
}

// AdminOnly checks if the caller is the admin and optionally updates the admin address
func AdminOnly(caller std.Address) error {
	return checkRole(ROLE_ADMIN, caller)
}

// GovernanceOnly checks if the caller is the governance and optionally updates the governance address
func GovernanceOnly(caller std.Address) error {
	return checkRole(ROLE_GOVERNANCE, caller)
}

// GovStakerOnly checks if the caller is the gov staker and optionally updates the gov staker address
func GovStakerOnly(caller std.Address) error {
	return checkRole(ROLE_GOV_STAKER, caller)
}

// RouterOnly checks if the caller is the router and optionally updates the router address
func RouterOnly(caller std.Address) error {
	return checkRole(ROLE_ROUTER, caller)
}

// PoolOnly checks if the caller is the pool and optionally updates the pool address
func PoolOnly(caller std.Address) error {
	return checkRole(ROLE_POOL, caller)
}

// PositionOnly checks if the caller is the position and optionally updates the position address
func PositionOnly(caller std.Address) error {
	return checkRole(ROLE_POSITION, caller)
}

// StakerOnly checks if the caller is the staker and optionally updates the staker address
func StakerOnly(caller std.Address) error {
	return checkRole(ROLE_STAKER, caller)
}

// LaunchpadOnly checks if the caller is the launchpad and optionally updates the launchpad address
func LaunchpadOnly(caller std.Address) error {
	return checkRole(ROLE_LAUNCHPAD, caller)
}

// EmissionOnly checks if the caller is the emission and optionally updates the emission address
func EmissionOnly(caller std.Address) error {
	return checkRole(ROLE_EMISSION, caller)
}

// ProtocolFeeOnly checks if the caller is the protocol fee and optionally updates the protocol fee address
func ProtocolFeeOnly(caller std.Address) error {
	return checkRole(ROLE_PROTOCOL_FEE, caller)
}

// GovXGNSOnly checks if the caller is the gov xgns and optionally updates the gov xgns address
func GovXGNSOnly(caller std.Address) error {
	return checkRole(ROLE_GOV_XGNS, caller)
}
