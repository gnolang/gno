package namereg

import (
	"gno.land/p/nt/ufmt/v0"
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

// NewSetPausedExecutor allows GovDAO to pause or unpause this realm
func NewSetPausedExecutor(newPausedValue bool) dao.ProposalRequest {
	cb := func(cur realm) error {
		setPaused(cur, newPausedValue)
		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	if newPausedValue {
		return dao.NewProposalRequest("User Registry V1: Pause", "", e)
	}

	return dao.NewProposalRequest("User Registry V1: Unpause", "", e)
}

// ProposeNewName allows GovDAO to propose a new name for an existing user.
// The associated address and all previous names of a user that changes a
// name are preserved, and all resolve to the new name.
//
// Governance renames bypass the Open Nym Tier `nym-...\d{3}` format intentionally.
// The DAO is the trust root for this realm; if voters approve a rename to
// `vitalik` (e.g. for trademark dispute resolution or system reservations),
// the validation here imposes no further opinion. The new name is still
// subject to the base shape enforced by `r/sys/users.validateName`
// (`^[a-z][a-z0-9]*([_-][a-z0-9]+)*$`, max 64 chars), which runs inside
// the updateUsername callback below.
func ProposeNewName(addr address, newName string) dao.ProposalRequest {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func(cur realm) error {
		err := updateUsername(cur, userData, newName)
		return err
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(
		ufmt.Sprintf("User Registry V1: Rename user `%s` to `%s`", userData.Name(), newName),
		"",
		e,
	)
}

// ProposeDeleteUser allows GovDAO to propose deletion of a user
// This will make the associated address and names unresolvable.
// WARN: After deletion, the same address WILL NOT be able to register a new name.
func ProposeDeleteUser(addr address, reason string) dao.ProposalRequest {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func(cur realm) error {
		return deleteUserdata(cur, userData)
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(
		ufmt.Sprintf("User Registry V1: Delete user `%s`", userData.Name()),
		reason,
		e,
	)
}

// ProposeNewRegisterPrice allows GovDAO to update the price of registration.
// Rejects prices below MinRegisterPrice (currently 0) at proposal-creation
// time. (audit finding #14: original code only rejected negative values,
// which would have been arithmetically nonsensical.)
func ProposeNewRegisterPrice(newPrice int64) dao.ProposalRequest {
	if newPrice < MinRegisterPrice {
		panic(ufmt.Sprintf("price below floor: %d ugnot < %d ugnot (MinRegisterPrice)",
			newPrice, MinRegisterPrice))
	}

	cb := func(cur realm) error {
		setRegisterPrice(cur, newPrice)
		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(
		ufmt.Sprintf("User Registry V1: Update registration price to `%d`", newPrice),
		"",
		e,
	)
}
