package common

import (
	"std"

	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/access"
)

const (
	ErrNoPermission = "caller(%s) has no permission"
)

// AssertCaller checks if the caller is the given address.
func AssertCaller(caller, addr std.Address) error {
	if caller != addr {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

func SatisfyCond(cond bool) error {
	if !cond {
		return ufmt.Errorf("given condition is not satisfied the permission check")
	}
	return nil
}

// AdminOnly checks if the caller is the admin.
func AdminOnly(caller std.Address) error {
	if err := access.AdminOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// GovernanceOnly checks if the caller is the gov governance contract.
func GovernanceOnly(caller std.Address) error {
	if err := access.GovernanceOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// GovStakerOnly checks if the caller is the gov staker contract.
func GovStakerOnly(caller std.Address) error {
	if err := access.GovStakerOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// RouterOnly checks if the caller is the router contract.
func RouterOnly(caller std.Address) error {
	if err := access.RouterOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// PoolOnly checks if the caller is the pool contract.
func PoolOnly(caller std.Address) error {
	if err := access.PoolOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// PositionOnly checks if the caller is the position contract.
func PositionOnly(caller std.Address) error {
	if err := access.PositionOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// StakerOnly checks if the caller is the staker contract.
func StakerOnly(caller std.Address) error {
	if err := access.StakerOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// LaunchpadOnly checks if the caller is the launchpad contract.
func LaunchpadOnly(caller std.Address) error {
	if err := access.LaunchpadOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// EmissionOnly checks if the caller is the emission contract.
func EmissionOnly(caller std.Address) error {
	if err := access.EmissionOnly(caller); err != nil {
		return ufmt.Errorf(ErrNoPermission, caller.String())
	}
	return nil
}

// UserOnly checks if the caller is a user.
func UserOnly(prev std.Realm) error {
	if !prev.IsUser() {
		return ufmt.Errorf("caller(%s) is not a user", prev.PkgPath())
	}
	return nil
}
