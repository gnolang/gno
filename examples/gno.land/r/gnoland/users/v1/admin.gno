package users

import (
	"std"

	"gno.land/p/demo/dao"

	"gno.land/r/gov/dao/bridge"
	susers "gno.land/r/sys/users"
)

var paused = false // XXX: replace with p/moul/authz

// NewSetPausedExecutor allows GovDAO to pause or unpause this realm
func NewSetPausedExecutor(newPausedValue bool) dao.Executor {
	cb := func() error {
		paused = newPausedValue
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}

// ProposeNewName allows GovDAO to propose a new name for an existing user
// The associated address and all previous names of a user that changes a name
// are preserved, and all resolve to the new name.
func ProposeNewName(addr std.Address, newName string) dao.Executor {
	if matched := reUsername.MatchString(newName); !matched {
		panic(ErrInvalidUsername)
	}

	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func() error {
		if err := userData.UpdateName(newName); err != nil {
			return err
		}
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}

// ProposeDeleteUser allows GovDAO to propose deletion of a user
// This will make the associated address and names unresolvable.
// WARN: After deletion, the same address WILL NOT be able to register a new name.
func ProposeDeleteUser(addr std.Address) dao.Executor {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func() error {
		if err := userData.Delete(); err != nil {
			return err
		}
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}

// ProposeNewRegisterPrice allows GovDAO to update the price of registration
func ProposeNewRegisterPrice(newPrice int64) dao.Executor {
	if newPrice < 0 {
		panic("invalid price")
	}

	cb := func() error {
		registerPrice = newPrice
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}
