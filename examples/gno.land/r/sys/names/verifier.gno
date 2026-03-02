// Package names provides functionality for checking of package deployments
// by users registered in r/sys/users are done to proper namespaces.
package names

import (
	"strings"

	"gno.land/r/sys/users"
)

// IsAuthorizedAddressForNamespace ensures that the given address has ownership of the given name.
// A user's name found in r/sys/users is equivalent to their namespace.
// Note: Whether namespace enforcement is enabled or not is controlled by the
// VM param `sysnames_enabled`. The keeper only calls this function when enforcement is on.
func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	return verifier(address_XXX, namespace)
}

// verifier checks the store to see that the
// user has properly registered a given name/namespace.
// This function considers as valid an `address` that matches the `namespace` (PA namespaces)
func verifier(address_XXX address, namespace string) bool {
	if strings.TrimSpace(address_XXX.String()) == "" || strings.TrimSpace(namespace) == "" {
		return false
	}

	// Allow user with their own address as namespace
	// This enables pseudo-anon namespaces
	// ie gno.land/{p,r}/{ADDRESS}/**
	if address_XXX.String() == namespace {
		return true
	}

	// Can be a registered namespace or an alias
	userData, _ := users.ResolveName(namespace)
	if userData == nil || userData.IsDeleted() {
		return false
	}

	/// XXX: add check for r/sys/teams down the line

	return userData.Addr() == address_XXX
}
