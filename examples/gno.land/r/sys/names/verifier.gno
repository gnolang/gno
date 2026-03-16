// Package names provides namespace authorization for package deployments,
// ensuring that users registered in r/sys/users can only deploy to namespaces they own.
package names

import (
	"gno.land/r/sys/users"
)

// IsAuthorizedAddressForNamespace checks whether the given address owns the given namespace.
// The VM keeper only calls this function when the `sysnames_pkgpath` param is non-empty.
func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	return verifier(address_XXX, namespace)
}

// verifier checks whether the address owns the namespace by looking up r/sys/users.
// An address always owns the namespace matching its own string representation (PA namespaces).
func verifier(address_XXX address, namespace string) bool {
	if address_XXX.String() == "" || namespace == "" {
		return false
	}

	// Allow deploying under your own address, e.g. gno.land/{p,r}/{ADDRESS}/**
	if address_XXX.String() == namespace {
		return true
	}

	// Check if the namespace is registered to this address
	userData, _ := users.ResolveName(namespace)
	if userData == nil || userData.IsDeleted() {
		return false
	}

	return userData.Addr() == address_XXX
}
