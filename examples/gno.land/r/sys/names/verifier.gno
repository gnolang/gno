// Package names provides functionality for checking of package deployments
// by users registered in r/sys/users are done to proper namespaces.
package names

import (
	"strings"

	"chain/runtime"

	"gno.land/r/sys/users"
)

var (
	admin = address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh") // govdao t1 multisig

	enabled = false
)

// IsAuthorizedAddressForNamespace ensures that the given address has ownership of the given name.
// A user's name found in r/sys/users is equivalent to their namespace.
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

// verifier checks the store to see that the
// user has properly registered a given name/namespace.
// This function considers as valid an `address` that matches the `namespace` (PA namespaces)
func verifier(enabled bool, address_XXX address, namespace string) bool {
	if !enabled {
		return true // only in pre-genesis cases
	}

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
