package users

import (
	"chain"
	"chain/runtime"

	"gno.land/p/moul/addrset"
	"gno.land/r/gov/dao"
)

const testControllerPath = "gno.land/r/sys/users/init"

var controllers = addrset.Set{} // caller whitelist

func init() {
	// On dev/test chains, auto-whitelist the init controller for bootstrapping.
	chainID := runtime.ChainID()
	if chainID == "dev" || chainID == "tendermint_test" {
		controllers.Add(chain.PackageAddress(testControllerPath))
	}
}

// ProposeNewController allows GovDAO to add a whitelisted caller
func ProposeNewController(addr address) dao.ProposalRequest {
	cb := func(cur realm) error {
		return addToWhitelist(addr)
	}

	return dao.NewProposalRequest("Add a new whitelisted caller at sys/users realm", "", dao.NewSimpleExecutor(cb, ""))
}

// ProposeControllerRemoval allows GovDAO to add a whitelisted caller
func ProposeControllerRemoval(addr address) dao.ProposalRequest {
	cb := func(cur realm) error {
		return deleteFromwhitelist(addr)
	}

	return dao.NewProposalRequest("Remove a whitelisted caller from sys/users realm", "", dao.NewSimpleExecutor(cb, ""))
}

// ProposeControllerAdditionAndRemoval allows GovDAO to add a new caller and remove an old caller in the same proposal.
func ProposeControllerAdditionAndRemoval(toAdd, toRemove address) dao.ProposalRequest {
	cb := func(cur realm) error {
		err := addToWhitelist(toAdd)
		if err != nil {
			return err
		}

		return deleteFromwhitelist(toRemove)
	}

	return dao.NewProposalRequest("Add and Remove whitelisted caller from sys/users realm", "", dao.NewSimpleExecutor(cb, ""))
}

// Helpers

func deleteFromwhitelist(addr address) error {
	if !controllers.Has(addr) {
		return NewErrNotWhitelisted()
	}

	if ok := controllers.Remove(addr); !ok {
		panic("failed to remove address from whitelist")
	}

	return nil
}

func addToWhitelist(newCaller address) error {
	if !newCaller.IsValid() {
		return ErrInvalidAddress
	}

	if !controllers.Add(newCaller) {
		return ErrAlreadyWhitelisted
	}

	return nil
}
