package access

import (
	"std"

	"gno.land/r/gnoswap/v1/rbac"
)

// checkAndUpdateRole checks if the caller has the given role and optionally updates the role address
func checkAndUpdateRole(roleName string, caller std.Address, newAddress ...std.Address) error {
	if err := rbac.CheckPermission(roleName, PERM_ACCESS, caller); err != nil {
		return err
	}

	// the number of new addresses that can be registered is limited to one
	if len(newAddress) == 1 {
		return UpdateRoleAddress(roleName, newAddress[0])
	}
	return nil
}

// AdminOnly checks if the caller is the admin and optionally updates the admin address
func AdminOnly(caller std.Address, newAdmin ...std.Address) error {
	return checkAndUpdateRole(ROLE_ADMIN, caller, newAdmin...)
}

// GovernanceOnly checks if the caller is the governance and optionally updates the governance address
func GovernanceOnly(caller std.Address, newGovernance ...std.Address) error {
	return checkAndUpdateRole(ROLE_GOVERNANCE, caller, newGovernance...)
}

// GovStakerOnly checks if the caller is the gov staker and optionally updates the gov staker address
func GovStakerOnly(caller std.Address, newGovStaker ...std.Address) error {
	return checkAndUpdateRole(ROLE_GOV_STAKER, caller, newGovStaker...)
}

// RouterOnly checks if the caller is the router and optionally updates the router address
func RouterOnly(caller std.Address, newRouter ...std.Address) error {
	return checkAndUpdateRole(ROLE_ROUTER, caller, newRouter...)
}

// PoolOnly checks if the caller is the pool and optionally updates the pool address
func PoolOnly(caller std.Address, newPool ...std.Address) error {
	return checkAndUpdateRole(ROLE_POOL, caller, newPool...)
}

// PositionOnly checks if the caller is the position and optionally updates the position address
func PositionOnly(caller std.Address, newPosition ...std.Address) error {
	return checkAndUpdateRole(ROLE_POSITION, caller, newPosition...)
}

// StakerOnly checks if the caller is the staker and optionally updates the staker address
func StakerOnly(caller std.Address, newStaker ...std.Address) error {
	return checkAndUpdateRole(ROLE_STAKER, caller, newStaker...)
}

// LaunchpadOnly checks if the caller is the launchpad and optionally updates the launchpad address
func LaunchpadOnly(caller std.Address, newLaunchpad ...std.Address) error {
	return checkAndUpdateRole(ROLE_LAUNCHPAD, caller, newLaunchpad...)
}

// EmissionOnly checks if the caller is the emission and optionally updates the emission address
func EmissionOnly(caller std.Address, newEmission ...std.Address) error {
	return checkAndUpdateRole(ROLE_EMISSION, caller, newEmission...)
}
