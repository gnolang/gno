package namereg

import (
	"chain"
	"strconv"
	"strings"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
	govimpl "gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
	susers "gno.land/r/sys/users"
)

var paused = false // XXX: replace with p/moul/authz

//----------------------------------------
// Privileged mutators.

func setPaused(cur realm, newPausedValue bool) {
	paused = newPausedValue
	if paused {
		chain.Emit("NameRegistrationPaused", "by", cur.Previous().PkgPath())
	} else {
		chain.Emit("NameRegistrationUnpaused", "by", cur.Previous().PkgPath())
	}
}

func updateUsername(cur realm, userData *susers.UserData, newName string) error {
	// UpdateName must be called from this realm.
	// If the canonical name exists it's re-mapped to the new name.
	return userData.UpdateNameIgnoreCanonical(0, cur, newName)
}

func deleteUserdata(cur realm, userData *susers.UserData) error {
	// Delete must be called from this realm.
	return userData.Delete(0, cur)
}

func setRegisterPrice(cur realm, newPrice int64) {
	registerPrice = newPrice
	chain.Emit(
		"NameRegistrationPriceChanged",
		"price", strconv.FormatInt(newPrice, 10),
		"by", cur.Previous().PkgPath(),
	)
}

//----------------------------------------
// Public API

// NewSetPausedExecutor allows GovDAO to pause or unpause this realm
func NewSetPausedExecutor(cur realm, newPausedValue bool) dao.ProposalRequest {
	if paused == newPausedValue {
		panic("paused value is already " + strconv.FormatBool(newPausedValue))
	}

	cb := func(cur realm) error {
		setPaused(cur, newPausedValue)
		return nil
	}

	e := dao.NewSimpleExecutor(0, cur, cb, "")
	filter := govimpl.NewFilterByTier(memberstore.T1)

	if newPausedValue {
		return dao.NewProposalRequestWithFilter("User Registry V1: Pause", "", e, filter)
	}

	return dao.NewProposalRequestWithFilter("User Registry V1: Unpause", "", e, filter)
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
//
// The new name overrides any canonical name collision, so if a prior canonical
// entry exists when the new name is applied it's silently overwritten, mapping
// the canonical name to the new name.
func ProposeNewName(cur realm, addr address, newName string) dao.ProposalRequest {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	// Check that name currently is not and has been registered to someone else
	// and if so panic to avoid creating a proposal that will panic on execution
	// because username is already taken, so it can't be used anymore.
	if susers.IsNameTaken(newName) {
		panic(susers.ErrNameTaken)
	}

	cb := func(cur realm) error {
		err := updateUsername(cur, userData, newName)
		return err
	}

	e := dao.NewSimpleExecutor(0, cur, cb, "")

	// Use user's address in the title because username could change before
	// the proposal is executed, which could be misleading when voting.
	title := ufmt.Sprintf(
		"User Registry V1: Rename name for %s to `%s`",
		userData.Addr(),
		newName,
	)
	desc := ufmt.Sprintf(
		"Renames [%s](/r/sys/namereg/v1:%s) to `%s`", // Allows checking current name
		userData.Addr(),
		userData.Addr(),
		newName,
	)

	return dao.NewProposalRequest(title, desc, e)
}

// ProposeDeleteUser allows GovDAO to propose deletion of a user
// This will make the associated address and names unresolvable.
// WARN: After deletion, the same address WILL NOT be able to register a new name.
func ProposeDeleteUser(cur realm, addr address, reason string) dao.ProposalRequest {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func(cur realm) error {
		return deleteUserdata(cur, userData)
	}

	e := dao.NewSimpleExecutor(0, cur, cb, "")

	desc := ufmt.Sprintf(
		"After deletion [%s](/r/sys/namereg/v1:%s) WILL NOT be able to register a new name.",
		userData.Addr(),
		userData.Addr(),
	)
	reason = strings.TrimSpace(reason)
	if reason != "" {
		desc += "\n\n" + reason
	}

	// Use user's address in the title because username could change before
	// the proposal is executed, which could be misleading when voting.
	title := ufmt.Sprintf("User Registry V1: Delete user for %s", userData.Addr())

	return dao.NewProposalRequest(title, desc, e)
}

// ProposeNewRegisterPrice allows GovDAO to update the price of registration.
// Rejects prices below MinRegisterPrice (currently 0) at proposal-creation
// time. (audit finding #14: original code only rejected negative values,
// which would have been arithmetically nonsensical.)
func ProposeNewRegisterPrice(cur realm, newPrice int64) dao.ProposalRequest {
	if newPrice < MinRegisterPrice {
		panic(ufmt.Sprintf("price below floor: %d ugnot < %d ugnot (MinRegisterPrice)",
			newPrice, MinRegisterPrice))
	}

	cb := func(cur realm) error {
		setRegisterPrice(cur, newPrice)
		return nil
	}

	e := dao.NewSimpleExecutor(0, cur, cb, "")

	return dao.NewProposalRequest(
		ufmt.Sprintf("User Registry V1: Update registration price to `%d`", newPrice),
		"",
		e,
	)
}
