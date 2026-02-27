// Package names enforces namespace permissions for package deployment.
// Only address-prefix (PA) namespaces are allowed.
package names

import (
	"gno.land/p/nt/ownable"
)

var (
	Ownable = ownable.NewWithAddress("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // dropped in genesis via Enable

	enabled = false
)

// IsAuthorizedAddressForNamespace checks if the given address can deploy to the given namespace.
// Only the address's own PA namespace is permitted.
func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	return verifier(enabled, address_XXX, namespace)
}

// Enable enables the namespace check and drops centralized ownership of this realm.
// The namespace check is disabled initially to ease txtar and other testing contexts,
// but this function is meant to be called in the genesis of a chain.
func Enable(cur realm) {
	if err := Ownable.DropOwnershipByPrevious(); err != nil {
		panic(err)
	}
	enabled = true
}

func IsEnabled() bool {
	return enabled
}

// verifier checks namespace deployment permissions.
// An address matching the namespace is the only allowed case.
func verifier(isEnabled bool, address_XXX address, namespace string) bool {
	if !isEnabled {
		return true // only in pre-genesis cases
	}

	if namespace == "" || !address_XXX.IsValid() {
		return false
	}

	// Allow user with their own address as namespace
	// ie gno.land/{p,r}/{ADDRESS}/**
	return address_XXX.String() == namespace
}
