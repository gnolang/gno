package users

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	p "github.com/gnolang/gno/examples/gno.land/p/demo/pausable"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

const (
	adminAddr              = "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq" // @moul
	WhitelistAddEvent      = "WhitelistAdd"
	WhitelistDeleteEvent   = "WhitelistDelete"
	VerifyFuncUpdatedEvent = "VerifyFuncUpdated"
)

var (
	admin    = ownable.NewWithAddress(adminAddr) // Package owner
	pausable = p.NewFromOwnable(admin)
)

func init() {
	callerWhitelist.Set(std.DerivePkgAddr("gno.land/r/gnoland/users"), struct{}{})
}

func AddWhitelistedCaller(addr std.Address) error {
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

func DelWhitelistedCaller(addr std.Address) error {
	if err := admin.CallerIsOwner(); err != nil {
		return err
	}

	if !callerWhitelist.Has(addr.String()) {
		return ErrDoesNotExist
	}

	callerWhitelist.Remove(addr.String())
	std.Emit(WhitelistDeleteEvent, "address", addr.String())

	return nil
}

// AdminUpdateVerifyCall updates the method that verifies the namespace.
func AdminUpdateVerifyCall(check VerifyNameFunc) error {
	if err := admin.CallerIsOwner(); err != nil {
		return err
	}

	checkFunc = check
	std.Emit(VerifyFuncUpdatedEvent)

	return nil
}

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
