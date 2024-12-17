package users

import (
	"std"

	"gno.land/p/demo/ownable"
	p "gno.land/p/demo/pausable"
)

const (
	adminAddr            = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul
	WhitelistAddEvent    = "WhitelistAdd"
	WhitelistDeleteEvent = "WhitelistDelete"

	gnolandUsers = "gno.land/r/gnoland/users" // preregistered with store write perms
)

var (
	AdminStore = ownable.NewWithAddress(adminAddr) // Package owner
	PauseStore = p.NewFromOwnable(AdminStore)
)

func init() {
	// Set initial whitelisted caller
	callerWhitelist.Set(std.DerivePkgAddr(gnolandUsers).String(), struct{}{})
}

// AddToWhitelist adds a new address to the caller whitelist
func AddToWhitelist(addr std.Address) error {
	if err := AdminStore.CallerIsOwner(); err != nil {
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
	if err := AdminStore.CallerIsOwner(); err != nil {
		return err
	}

	if !callerWhitelist.Has(addr.String()) {
		return ErrNotWhitelisted
	}

	_, removed := callerWhitelist.Remove(addr.String())
	if !removed {
		panic("failed to remove from whitelist") // should never happen
	}

	std.Emit(WhitelistDeleteEvent, "address", addr.String())

	return nil
}

// Exposing Ownable & Pausable APIs

func IsPaused() bool {
	return PauseStore.IsPaused()
}

func Pause() error {
	return PauseStore.Pause()
}

func Unpause() error {
	return PauseStore.Unpause()
}

func GetOwner() std.Address {
	return AdminStore.Owner()
}

func TransferOwnership(newOwner std.Address) error {
	if err := AdminStore.TransferOwnership(newOwner); err != nil {
		return err
	}

	return nil
}
