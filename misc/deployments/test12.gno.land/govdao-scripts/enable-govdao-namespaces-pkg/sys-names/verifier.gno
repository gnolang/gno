// Package names/v2 enforces namespace permissions for package deployment.
// Supports PA (address-prefix) namespaces and DAO-registered custom namespaces.
package names

import "gno.land/p/nt/avl/v0"

// namespaceStore maps custom namespace string → owner address.
var namespaceStore = avl.NewTree()

// IsAuthorizedAddressForNamespace checks if the given address can deploy to the given namespace.
// Allows PA (own address) namespaces and DAO-registered custom namespaces.
func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	if namespace == "" || !address_XXX.IsValid() {
		return false
	}
	if address_XXX.String() == namespace {
		return true // PA namespace
	}
	owner, ok := namespaceStore.Get(namespace)
	if !ok {
		return false
	}
	return owner.(address) == address_XXX
}
