// Package names provides functionality for checking of package deployments
// by users registered in r/sys/users are done to proper namespaces.
package names

import (
	"std"
	"strings"

	"gno.land/p/demo/ownable"

	"gno.land/r/sys/users"
)

var (
	Ownable = ownable.NewWithAddress("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // dropped in genesis via Enable

	enabled = false
)

// IsAuthorizedAddressForName ensures that the given address has ownership of the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return verifier(enabled, address, name)
}

// Enable enables the namespace check and drops centralized ownership of this realm.
// The namespace check is disabled initially to ease txtar and other testing contexts,
// but this function is meant to be called in the genesis of a chain.
func Enable() {
	Ownable.AssertCallerIsOwner()
	enabled = true
	Ownable.DropOwnership()
}

func IsEnabled() bool {
	return enabled
}

// verifier checks the store to see that the
// user has properly registered a given name.
// This function considers as valid an `address` that matches the `name` (PA namespaces)
func verifier(enabled bool, address std.Address, name string) bool {
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
