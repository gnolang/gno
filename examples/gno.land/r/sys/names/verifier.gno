package names

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	"std"
	"strings"

	susers "gno.land/r/sys/users"
)

const (
	VerifyFuncUpdatedEvent = "VerifyFuncUpdated"
	adminAddr              = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul
)

type verifierFunc func(enabled bool, address std.Address, name string) bool

var (
	checkFunc  = defaultVerifier // Checking namespace callback
	AdminStore = ownable.NewWithAddress(adminAddr)
)

// IsAuthorizedAddressForName ensures that the given address has ownership of the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return checkFunc(susers.IsPaused(), address, name)
}

// UpdateVerifyCall updates the method that verifies the namespace.
func UpdateVerifyCall(check verifierFunc) error {
	if err := AdminStore.CallerIsOwner(); err != nil {
		return err
	}

	checkFunc = check
	std.Emit(VerifyFuncUpdatedEvent)

	return nil
}

// defaultVerifyFunction checks the store to see that the
// user has properly registered the given name.
// This function considers as valid an `address` that matches the `name`.
func defaultVerifier(enabled bool, address std.Address, name string) bool {
	if !enabled {
		return true
	}

	if strings.TrimSpace(address) == "" || strings.TrimSpace(name) == "" {
		return false // XXX: possibly redundant?
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
