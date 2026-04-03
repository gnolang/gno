package users

import (
	"chain"
	"chain/runtime"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/gov/dao"
)

const initControllerPath = "gno.land/r/sys/users/init"

var controllers = addrset.Set{} // caller whitelist

func init() {
	// auto-whitelist the init controller for bootstrapping for testing chain.
	if chainID := runtime.ChainID(); chainID == "dev" {
		controllers.Add(chain.PackageAddress(initControllerPath))
	}
}

// AddControllerAtGenesis allows adding a controller during chain genesis (height 0).
// This is mostly useful for testing.
func AddControllerAtGenesis(_ realm, addr address) {
	height := runtime.ChainHeight()
	if height > 0 {
		panic("AddControllerAtGenesis can only be called at genesis (height 0)")
	}

	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	controllers.Add(addr)
}

// ProposeNewController allows GovDAO to add a whitelisted caller
func ProposeNewController(addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		return addToWhitelist(addr)
	}

	desc := "This proposal adds " + addr.String() + " to `sys/users` realm's callers whitelist."
	return dao.NewProposalRequest("Add Whitelisted Caller to \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// ProposeControllerRemoval allows GovDAO to add a whitelisted caller
func ProposeControllerRemoval(addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		return deleteFromWhitelist(addr)
	}

	desc := "This proposal removes " + addr.String() + " from `sys/users` realm's callers whitelist."
	return dao.NewProposalRequest("Remove Whitelisted Caller From \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
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

	desc := ufmt.Sprint(
		"This proposal adds %s and removes %s from `sys/users` realm's callers whitelist.",
		toAdd,
		toRemove,
	)
	return dao.NewProposalRequest("Add and Remove Whitelisted Callers From \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// ProposeRegisterUser allows GovDAO to register a name without checking controllers
func ProposeRegisterUser(name string, addr address) dao.ProposalRequest {
	// Validate the name and address now, even though registerUser will validate again
	if err := validateName(name); err != nil {
		panic(err.Error())
	}
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		return registerUser(cur, name, addr)
	}

	desc := "This proposal registers " + name + " with address " + addr.String() + " in `sys/users`."
	return dao.NewProposalRequest("Register User to \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// ProposeUpdateName allows GovDAO to update a name with an alias without checking controllers
func ProposeUpdateName(addr address, newName string) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}
	if err := validateName(newName); err != nil {
		panic(err.Error())
	}

	cb := func(cur realm) error {
		data := ResolveAddress(addr)
		if data == nil {
			return ErrUserNotExistOrDeleted
		}
		return data.updateName(newName)
	}

	desc := "This proposal updates address " + addr.String() + " with alias " + newName + " in `sys/users`."
	return dao.NewProposalRequest("Update Name Alias in \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
}

// ProposeDeleteUser allows GovDAO to delete a user without checking controllers
func ProposeDeleteUser(addr address) dao.ProposalRequest {
	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	cb := func(cur realm) error {
		data := ResolveAddress(addr)
		if data == nil {
			return ErrUserNotExistOrDeleted
		}
		return data.delete()
	}

	desc := "This proposal deletes the user with address " + addr.String() + " in `sys/users`."
	return dao.NewProposalRequest("Delete User in \"sys/users\" Realm", desc, dao.NewSimpleExecutor(cb, ""))
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
