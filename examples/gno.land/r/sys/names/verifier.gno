package names

import (
	"std"
	"strings"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/pausable"

	"gno.land/r/sys/users"
)

const (
	VerifyFuncUpdatedEvent = "VerifyFuncUpdated"
	adminAddr              = "g1manfred47kzduec920z88wfr64ylksmdcedlf5" // @moul
)

type verifierFunc func(enabled bool, address std.Address, name string) bool

var (
	checkFunc = defaultVerifier // Callback for the namespace check

	// Safe objects

	// Ownable manages the administration of this realm
	Ownable = ownable.NewWithAddress(adminAddr)

	// Pausable manages whether namespace registration is enabled on-chain or not
	Pausable = pausable.NewFromOwnable(Ownable)
)

// IsAuthorizedAddressForName ensures that the given address has ownership of the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return checkFunc(Pausable.IsPaused(), address, name)
}

// defaultVerifyFunction checks the store to see that the
// user has properly registered the given name.
// This function considers as valid an `address` that matches the `name`.
func defaultVerifier(enabled bool, address std.Address, name string) bool {
	if !enabled {
		return true
	}

	if strings.TrimSpace(address.String()) == "" || strings.TrimSpace(name) == "" {
		return false // XXX: possibly redundant?
	}

	// Allow user with their own address as name
	// This enables pseudo-anon namespaces
	if address.String() == name {
		return true
	}

	/// XXX: add check for r/sys/teams down the line

	// Can be a registered name or an alias
	lookup := users.ResolveName(name)
	return lookup.Addr == address
}

// UpdateVerifyCall updates the method that verifies the namespace.
func UpdateVerifyCall(check verifierFunc) error {
	if !Ownable.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	checkFunc = check
	std.Emit(VerifyFuncUpdatedEvent)

	return nil
}
