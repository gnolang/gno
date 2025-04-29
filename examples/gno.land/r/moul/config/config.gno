package config

import (
	"errors"
	"std"

	"gno.land/p/moul/authz"
	// Import ufmt if needed for more complex string formatting in Render, otherwise manager.String() is fine.
	// "gno.land/p/demo/ufmt"
)

var (
	// manager controls the configuration authority.
	manager *authz.Authorizer
	// Define the initial manager address for clarity and reuse.
	initialManager = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
)

func init() {
	// Initialize the manager with a specific MemberAuthority.
	initialAuthority := authz.NewMemberAuthority(
		initialManager,
		// Add other initial members here if needed.
	)
	// Wrap the initial authority in an Authorizer.
	manager = authz.NewWithAuthority(initialAuthority)
}

// Render displays the current managing authority's configuration.
func Render() string {
	if manager == nil {
		return "Error: Manager not initialized"
	}
	// The String() method of the underlying authority provides the content.
	return "Current Authority:\n" + manager.String()
}

// AddMember adds a member to the authority, requires authorization.
// This only works if the current authority is a MemberAuthority.
func AddMember(addr std.Address) error {
	if manager == nil {
		return errors.New("manager not initialized")
	}
	memberAuth, ok := manager.Current().(*authz.MemberAuthority)
	if !ok {
		return errors.New("current authority is not a MemberAuthority, cannot add member directly")
	}
	// The AddMember method on MemberAuthority performs its own authorization check.
	return memberAuth.AddMember(addr)
}

// RemoveMember removes a member from the authority, requires authorization.
// This only works if the current authority is a MemberAuthority.
func RemoveMember(addr std.Address) error {
	if manager == nil {
		return errors.New("manager not initialized")
	}
	memberAuth, ok := manager.Current().(*authz.MemberAuthority)
	if !ok {
		return errors.New("current authority is not a MemberAuthority, cannot remove member directly")
	}
	// The RemoveMember method on MemberAuthority performs its own authorization check.
	return memberAuth.RemoveMember(addr)
}

// TransferAuthority changes the managing authority, requires authorization by the current authority.
func TransferAuthority(newAuthority authz.Authority) error {
	if manager == nil {
		return errors.New("manager not initialized")
	}
	if newAuthority == nil {
		return errors.New("new authority cannot be nil")
	}
	// The Transfer method on Authorizer performs the authorization check using the current authority.
	return manager.Transfer(newAuthority)
}
