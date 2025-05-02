package users

import (
	"std"

	"gno.land/r/gov/dao"
	susers "gno.land/r/sys/users"
)

var paused = false // XXX: replace with p/moul/authz

//----------------------------------------
// Privileged mutators.

func setPaused(newPausedValue bool) {
	crossing()
	paused = newPausedValue
}

func updateUsername(userData *susers.UserData, newName string) error {
	crossing()
	// UpdateName must be called from this realm.
	return userData.UpdateName(newName)
}

func deleteUserdata(userData *susers.UserData) error {
	crossing()
	// Delete must be called from this realm.
	return userData.Delete()
}

func setRegisterPrice(newPrice int64) {
	crossing()
	registerPrice = newPrice
}

//----------------------------------------
// Public API

// NewSetPausedExecutor allows GovDAO to pause or unpause this realm
func NewSetPausedExecutor(newPausedValue bool) dao.ProposalRequest {
	cb := func() error {
		cross(setPaused)(newPausedValue)
		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest("Pause users/v1 realm", "", e)
}

// ProposeNewName allows GovDAO to propose a new name for an existing user
// The associated address and all previous names of a user that changes a name
// are preserved, and all resolve to the new name.
func ProposeNewName(addr std.Address, newName string) dao.ProposalRequest {
	if matched := reUsername.MatchString(newName); !matched {
		panic(ErrInvalidUsername)
	}

	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func() error {
		err := cross(updateUsername)(userData, newName)
		return err
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest("Propose a new name using users/v1 realm", "", e)
}

// ProposeDeleteUser allows GovDAO to propose deletion of a user
// This will make the associated address and names unresolvable.
// WARN: After deletion, the same address WILL NOT be able to register a new name.
func ProposeDeleteUser(addr std.Address, desc string) dao.ProposalRequest {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func() error {
		return cross(deleteUserdata)(userData)
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest("Propose deleting a name using users/v1 realm", desc, e)
}

// ProposeNewRegisterPrice allows GovDAO to update the price of registration
func ProposeNewRegisterPrice(newPrice int64) dao.ProposalRequest {
	if newPrice < 0 {
		panic("invalid price")
	}

	cb := func() error {
		cross(setRegisterPrice)(newPrice)
		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(
		"Propose change the price for name registration using users/v1 realm",
		"",
		e,
	)
}
