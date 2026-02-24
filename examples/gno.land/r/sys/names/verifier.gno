// Package names enforces namespace permissions for package deployment.
// Only address-prefix (PA) namespaces are allowed. Reserved namespaces
// are blocked from all users.
package names

import (
	"gno.land/p/nt/ownable"
)

var (
	Ownable = ownable.NewWithAddress("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // dropped in genesis via Enable

	enabled = false
)

// isReservedNamespace reports whether namespace is reserved for system use.
// Reserved namespaces cannot be claimed by any user.
func isReservedNamespace(namespace string) bool {
	switch namespace {
	case "gnoland", "nt", "sys", "gov":
		return true
	default:
		return false
	}
}

// IsAuthorizedAddressForNamespace checks if the given address can deploy to the given namespace.
// Only the address's own PA namespace is permitted. Reserved namespaces are always denied.
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
// An address matching the namespace (PA namespace) is the only allowed case.
// Reserved namespaces are denied for all users.
func verifier(isEnabled bool, address_XXX address, namespace string) bool {
	if !isEnabled {
		return true // only in pre-genesis cases
	}

	if address_XXX.String() == "" || namespace == "" {
		return false
	}

	if isReservedNamespace(namespace) {
		return false
	}

	// Allow user with their own address as namespace
	// ie gno.land/{p,r}/{ADDRESS}/**
	return address_XXX.String() == namespace
}
