// Package names provides namespace authorization for package deployments.
// Only address-prefix (PA) namespaces are allowed.
package names

// IsAuthorizedAddressForNamespace checks whether the given address owns the given namespace.
// The VM keeper only calls this function when the `sysnames_pkgpath` param is non-empty.
func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	return verifier(address_XXX, namespace)
}

// verifier checks namespace deployment permissions.
// An address matching the namespace is the only allowed case.
func verifier(address_XXX address, namespace string) bool {
	if address_XXX.String() == "" || namespace == "" {
		return false
	}

	// Allow deploying under your own address, e.g. gno.land/{p,r}/{ADDRESS}/**
	return address_XXX.String() == namespace
}
