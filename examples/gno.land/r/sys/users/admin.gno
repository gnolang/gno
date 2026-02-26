package users

import (
	"chain"
	"errors"

	"gno.land/p/moul/addrset"
	"gno.land/r/gov/dao"
)

const gUsersV1Path = "gno.land/r/gnoland/users/v1" // preregistered with store write perms

var controllers = addrset.Set{} // caller whitelist

func init() {
	controllers.Add(chain.PackageAddress(gUsersV1Path)) // initially whitelisted
}

// ProposeNewController allows GovDAO to add a whitelisted caller
func ProposeNewController(addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		return addToWhitelist(addr)
	}

	return dao.NewProposalRequest("Add a new whitelisted caller at sys/users realm", "", dao.NewSimpleExecutor(cb, ""))
}

// ProposeControllerRemoval allows GovDAO to add a whitelisted caller
func ProposeControllerRemoval(addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		return deleteFromWhitelist(addr)
	}

	return dao.NewProposalRequest("Remove a whitelisted caller from sys/users realm", "", dao.NewSimpleExecutor(cb, ""))
}

// ProposeControllerAdditionAndRemoval allows GovDAO to add a new caller and remove an old caller in the same proposal.
func ProposeControllerAdditionAndRemoval(toAdd, toRemove address) dao.ProposalRequest {
	if !toAdd.IsValid() || !toRemove.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		err := addToWhitelist(toAdd)
		if err != nil {
			return err
		}

		return deleteFromWhitelist(toRemove)
	}

	return dao.NewProposalRequest("Add and Remove whitelisted caller from sys/users realm", "", dao.NewSimpleExecutor(cb, ""))
}

// Helpers

func deleteFromWhitelist(addr address) error {
	if !controllers.Has(addr) {
		return NewErrNotWhitelisted()
	}

	if ok := controllers.Remove(addr); !ok {
		return ErrWhitelistRemoveFailed
	}

	return nil
}

func addToWhitelist(newCaller address) error {
	if !controllers.Add(newCaller) {
		return ErrAlreadyWhitelisted
	}

	return nil
}
