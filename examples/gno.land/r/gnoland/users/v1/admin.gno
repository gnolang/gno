package users

import (
	"std"

	"gno.land/r/gov/dao"
	susers "gno.land/r/sys/users"
)

var paused = false // XXX: replace with p/moul/authz

// NewSetPausedExecutor allows GovDAO to pause or unpause this realm
func NewSetPausedExecutor(newPausedValue bool) dao.ProposalRequest {
	cb := func(m dao.Metadata) error {
		paused = newPausedValue
		return nil
	}

	return dao.NewProposalRequestBuilder("Pause users/v1 realm").WithCallback(cb).Build()
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

	cb := func(dao.Metadata) error {
		if err := userData.UpdateName(newName); err != nil {
			return err
		}
		return nil
	}

	return dao.NewProposalRequestBuilder("Propose a new name using users/v1 realm").WithCallback(cb).Build()
}

// ProposeDeleteUser allows GovDAO to propose deletion of a user
// This will make the associated address and names unresolvable.
// WARN: After deletion, the same address WILL NOT be able to register a new name.
func ProposeDeleteUser(addr std.Address, desc string) dao.ProposalRequest {
	userData := susers.ResolveAddress(addr)
	if userData == nil {
		panic(susers.ErrUserNotExistOrDeleted)
	}

	cb := func(dao.Metadata) error {
		if err := userData.Delete(); err != nil {
			return err
		}
		return nil
	}

	return dao.NewProposalRequestBuilder("Propose deleting a name using users/v1 realm").WithDescription(desc).WithCallback(cb).Build()
}

// ProposeNewRegisterPrice allows GovDAO to update the price of registration
func ProposeNewRegisterPrice(newPrice int64) dao.ProposalRequest {
	if newPrice < 0 {
		panic("invalid price")
	}

	cb := func(dao.Metadata) error {
		registerPrice = newPrice
		return nil
	}

	return dao.NewProposalRequestBuilder("Propose change the price for name registration using users/v1 realm").WithCallback(cb).Build()
}
