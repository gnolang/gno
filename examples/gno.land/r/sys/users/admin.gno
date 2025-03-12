package users

import (
	"std"

	"gno.land/p/moul/addrset"
	"gno.land/r/gov/dao"
)

const gusersv1 = "gno.land/r/gnoland/users/v1" // preregistered with store write perms

var controllers = addrset.Set{} // caller whitelist

func init() {
	controllers.Add(std.DerivePkgAddr(gusersv1)) // initially whitelisted
}

// ProposeNewController allows GovDAO to add a whitelisted caller
func ProposeNewController(addr std.Address) dao.ProposalRequest {
	cb := func(dao.Metadata) error {
		return addToWhitelist(addr)
	}

	return dao.NewProposalRequestBuilder("Add a new whitelisted caller at sys/users realm").WithCallback(cb).Build()
}

// ProposeControllerRemoval allows GovDAO to add a whitelisted caller
func ProposeControllerRemoval(addr std.Address) dao.ProposalRequest {
	cb := func(dao.Metadata) error {
		return deleteFromwhitelist(addr)
	}

	return dao.NewProposalRequestBuilder("Remove a whitelisted caller from sys/users realm").WithCallback(cb).Build()
}

// Helpers

func deleteFromwhitelist(addr std.Address) error {
	if !controllers.Has(addr) {
		return ErrNotWhitelisted
	}

	if ok := controllers.Remove(addr); !ok {
		panic("failed to remove address from whitelist")
	}

	return nil
}

func addToWhitelist(newCaller std.Address) error {
	if !newCaller.IsValid() {
		return ErrInvalidAddress
	}

	if !controllers.Add(newCaller) {
		return ErrAlreadyWhitelisted
	}

	return nil
}
