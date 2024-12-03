package users

import (
	"std"

	"gno.land/p/demo/ownable"
	p "gno.land/p/demo/pausable"
)

const (
	adminAddr              = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul
	WhitelistAddEvent      = "WhitelistAdd"
	WhitelistDeleteEvent   = "WhitelistDelete"
	VerifyFuncUpdatedEvent = "VerifyFuncUpdated"

	gnolandUsers = "gno.land/r/gnoland/users" // preregistered with store write perms
)

var (
	admin    = ownable.NewWithAddress(adminAddr) // Package owner
	pausable = p.NewFromOwnable(admin)
)

func init() {
	// Set initial whitelisted caller
	callerWhitelist.Set(std.DerivePkgAddr(gnolandUsers).String(), struct{}{})
}

// AddToWhitelist adds a new address to the caller whitelist
func AddToWhitelist(addr std.Address) error {
	if err := admin.CallerIsOwner(); err != nil {
		return err
	}

	if !addr.IsValid() {
		return ErrInvalidAddress
	}

	if callerWhitelist.Has(addr.String()) {
		return ErrAlreadyWhitelisted
	}

	callerWhitelist.Set(addr.String(), struct{}{})
	std.Emit(WhitelistAddEvent, "address", addr.String())

	return nil
}

// DelFromWhitelist removes a caller from the whitelist
func DelFromWhitelist(addr std.Address) error {
	if err := admin.CallerIsOwner(); err != nil {
		return err
	}

	if !callerWhitelist.Has(addr.String()) {
		return ErrNotWhitelisted
	}

	_, removed := callerWhitelist.Remove(addr.String())
	if !removed {
		// do what?
	}

	std.Emit(WhitelistDeleteEvent, "address", addr.String())

	return nil
}

// UpdateVerifyCall updates the method that verifies the namespace.
func UpdateVerifyCall(check VerifyNameFunc) error {
	if err := admin.CallerIsOwner(); err != nil {
		return err
	}

	checkFunc = check
	std.Emit(VerifyFuncUpdatedEvent)

	return nil
}

// Exposing Ownable & Pausable APIs

func Pause() error {
	return pausable.Pause()
}

func Unpause() error {
	return pausable.Unpause()
}

func GetOwner() std.Address {
	return admin.Owner()
}

func TransferOwnership(newOwner std.Address) error {
	if err := admin.TransferOwnership(newOwner); err != nil {
		return err
	}

	return nil
}
