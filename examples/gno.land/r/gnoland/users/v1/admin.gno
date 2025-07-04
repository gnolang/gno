package users

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
	susers "gno.land/r/sys/users"
)

var paused = false // XXX: replace with p/moul/authz

//----------------------------------------
// Privileged mutators.

func setPaused(cur realm, newPausedValue bool) {
	paused = newPausedValue
}

func updateUsername(cur realm, userData *susers.UserData, newName string) error {
	// UpdateName must be called from this realm.
	return userData.UpdateName(newName)
}

func deleteUserdata(cur realm, userData *susers.UserData) error {
	// Delete must be called from this realm.
	return userData.Delete()
}

func setRegisterPrice(cur realm, newPrice int64) {
	registerPrice = newPrice
}

//----------------------------------------
// Public API

// ProposeArbitraryName allows GovDAO to propose a fully arbitrary name for an address
func ProposeArbitraryName(name, reason string, addr std.Address) dao.ProposalRequest {
	if name == "" {
		panic("name cannot be empty")
	}

	if addr.String() == "" {
		panic("address cannot be empty")
	}

	userData, _ := susers.ResolveAny(name)
	if userData != nil {
		panic("Name already taken")
	}

	userData = susers.ResolveAddress(addr)

	var cb func() error
	// If address has a name already, update the name
	if userData != nil {
		cb = func() error {
			return updateUsername(cross, userData, name)
		}
	} else {
		cb = func() error {
			return susers.RegisterUser(cross, name, addr)
		}
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(ufmt.Sprintf("User Registry V1: Give `%s` to `%s`", name, addr.String()), reason, e)
}

// ProposeNewPausedValue allows GovDAO to pause or unpause this realm
func ProposeNewPausedValue(newPausedValue bool, reason string) dao.ProposalRequest {
	cb := func() error {
		setPaused(cross, newPausedValue)
		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	if newPausedValue {
		return dao.NewProposalRequest("User Registry V1: Pause", reason, e)
	}

	return dao.NewProposalRequest("User Registry V1: Unpause", reason, e)
}

// ProposeNewName allows GovDAO to propose a new name for an existing user
// The associated address and all previous names of a user that changes a name
// are preserved, and all resolve to the new name.
// The new name can be completely arbitrary.
func ProposeNewName(addr std.Address, newName, reason string) dao.ProposalRequest {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func() error {
		err := updateUsername(cross, userData, newName)
		return err
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(ufmt.Sprintf("User Registry V1: Rename user `%s` to `%s`", userData.Name(), newName), reason, e)
}

// ProposeDeleteUser allows GovDAO to propose deletion of a user
// This will make the associated address and names unresolvable.
// WARN: After deletion, the same address WILL NOT be able to register a new name.
func ProposeDeleteUser(addr std.Address, reason string) dao.ProposalRequest {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func() error {
		return deleteUserdata(cross, userData)
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(ufmt.Sprintf("User Registry V1: Delete user `%s`", userData.Name()), reason, e)
}

// ProposeNewRegisterPrice allows GovDAO to update the price of registration
func ProposeNewRegisterPrice(newPrice int64, reason string) dao.ProposalRequest {
	if newPrice < 0 {
		panic("invalid price")
	}

	cb := func() error {
		setRegisterPrice(cross, newPrice)
		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(
		ufmt.Sprintf("User Registry V1: Change registration price to `%dugnot`", newPrice),
		reason,
		e,
	)
}
