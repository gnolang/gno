package config

import (
	"errors"
	"std"

	"gno.land/p/moul/addrset"
	"gno.land/p/moul/authz"
)

var (
	Authority, privConfig = authz.NewWithOrigin("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	SafeConfig            = privConfig.Safe()
)

func init() {
	// Configure permissions for management actions
	privConfig.Add("add_manager", "g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	privConfig.Add("remove_manager", "g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	privConfig.Add("transfer_management", "g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	privConfig.Add("call_privileged_action", "g1manfred47kzduec920z88wfr64ylksmdcedlf5")
}

func crossing() {
	// This function can be used for logging or debugging
	// when crossing permission boundaries
}

func PrivilegedAction() {
	crossing()
	Authority.AssertPreviousCan("call_privileged_action")
	std.Emit("PrivilegedAction called",
		"by", std.PreviousRealm().Address().String(),
	)
}

// AddManager adds a new address to the list of authorized managers.
func AddManager(addr std.Address) error {
	Authority.AssertPreviousCan("add_manager")
	
	// Add the new manager to the admin membership
	privConfig.Add(authz.Admin, addr)
	privConfig.Add("add_manager", addr)
	privConfig.Add("remove_manager", addr)
	privConfig.Add("transfer_management", addr)
	privConfig.Add("call_privileged_action", addr)
	
	return nil
}

// RemoveManager removes an address from the list of authorized managers.
func RemoveManager(addr std.Address) error {
	Authority.AssertPreviousCan("remove_manager")
	
	// Note: In the new authz system, we can't remove from composite memberships easily
	// This would require a more sophisticated implementation with explicit manager tracking
	// For now, return an error indicating this needs to be implemented
	return errors.New("remove manager not yet implemented in new authz system")
}

// TransferManagement transfers the authority to manage keys to a new authority.
func TransferManagement(newAuthority authz.Membership) error {
	Authority.AssertPreviousCan("transfer_management")
	
	if newAuthority == nil {
		return errors.New("new authority cannot be nil")
	}
	
	// Transfer all management permissions to the new authority
	privConfig.Set(authz.Admin, newAuthority)
	privConfig.Set("add_manager", newAuthority)
	privConfig.Set("remove_manager", newAuthority)
	privConfig.Set("transfer_management", newAuthority)
	privConfig.Set("call_privileged_action", newAuthority)
	
	return nil
}

// ListManagers returns a slice of all managed keys.
func ListManagers() []std.Address {
	// In the new authz system, we'd need to track managers explicitly
	// For now, return the original admin
	return []std.Address{"g1manfred47kzduec920z88wfr64ylksmdcedlf5"}
}

func HasManager(addr std.Address) bool {
	return Authority.Can(addr, authz.Admin)
}
