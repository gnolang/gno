package names

import (
	"std"

	susers "gno.land/r/sys/users"
)

const VerifyFuncUpdatedEvent = "VerifyFuncUpdated"

type verifierFunc func(enabled bool, address std.Address, name string) bool

var (
	checkFunc = defaultVerifier // Checking namespace callback
)

// IsAuthorizedAddressForName ensures that the given address has ownership of the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return checkFunc(susers.IsPaused(), address, name)
}

// defaultVerifyFunction checks the store to see that the
// user has properly registered the given name.
// This function considers as valid an `address` that matches the `name`.
func defaultVerifier(enabled bool, address std.Address, name string) bool {
	if !enabled {
		return true
	}

	// Allow user with their own address as name
	// This enables pseudo-anon namespaces
	if address.String() == name {
		return true
	}

	// Can be a registered name or an alias
	lookup := susers.ResolveName(name)
	return lookup.Addr == address
}
