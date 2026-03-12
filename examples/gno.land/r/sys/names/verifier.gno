// Package names enforces namespace permissions for package deployment.
// Only address-prefix (PA) namespaces are allowed.
package names

import "chain/runtime"

var (
	admin = address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh") // govdao t1 multisig
	enabled = false
)

// IsAuthorizedAddressForNamespace checks if the given address can deploy to the given namespace.
// Only the address's own PA namespace is permitted.
func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	return verifier(enabled, address_XXX, namespace)
}

// Enable enables the namespace check for this realm.
// The namespace check is disabled initially to ease txtar and other testing contexts,
// but this function is meant to be called in the genesis of a chain.
func Enable(cur realm) {
	if runtime.PreviousRealm().Address() != admin {
		panic("caller is not admin")
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
