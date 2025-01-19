// Package names provides functionality for checking of package deployments
// by users registered in r/sys/users are done to proper namespaces.
package names

import (
	"std"
	"strings"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ownable"

	"gno.land/r/gov/dao/bridge"
	"gno.land/r/sys/users"
)

type verifierFunc func(address std.Address, name string) bool

var (
	Ownable = ownable.NewWithAddress("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul > dropped in genesis via Enable

	checkFunc = defaultVerifier // Callback for the namespace check
	enabled   = false           // set to true in genesis
)

const VerifyFuncUpdatedEvent = "VerifyFuncUpdated"

// IsAuthorizedAddressForName ensures that the given address has ownership of the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return checkFunc(enabled, address, name)
}

// Enable enables the namespace check and drops centralized ownership of this realm.
// The namespace check is disabled initially to ease txtar and other testing contexts,
// but this function is meant to be called in the genesis of a chain.
// This way, only GovDAO will be able to modify the namespace checking function.
func Enable() {
	Ownable.AssertCallerIsOwner()
	enabled = true
	Ownable.DropOwnership()
}

func IsEnabled() bool {
	return enabled
}

// defaultVerifyFunction checks the store to see that the
// user has properly registered the given name.
// This function considers as valid an `address` that matches the `name`.
func defaultVerifier(enabled bool, address std.Address, name string) bool {
	if !enabled {
		return true // only in pre-genesis cases
	}

	if strings.TrimSpace(address.String()) == "" || strings.TrimSpace(name) == "" {
		return false
	}

	// Allow user with their own address as name
	// This enables pseudo-anon namespaces
	if address.String() == name {
		return true
	}

	// Can be a registered name or an alias
	userData, _ := users.ResolveName(name)
	if userData == nil || userData.IsDeleted() {
		return false
	}

	/// XXX: add check for r/sys/teams down the line

	return userData.Addr() == address
}

// NewVerifyCallExecutor allows updating the verifier function via a GovDAO proposal
func NewVerifyCallExecutor(newVerifyCall func(enabled bool, address std.Address, name string) bool) dao.Executor {
	callback := func() error {
		checkFunc = newVerifyCall

		std.Emit(VerifyFuncUpdatedEvent)
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(callback)
}
