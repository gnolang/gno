// Package names provides functionality for checking if package deployments
// by users registered in r/sys/users are done to proper namespaces.
package names

import (
	"std"
	"strings"

	"gno.land/p/demo/dao"

	"gno.land/r/gov/dao/bridge"
	"gno.land/r/sys/users"
)

type verifierFunc func(address std.Address, name string) bool

var checkFunc = defaultVerifier // Callback for the namespace check

const VerifyFuncUpdatedEvent = "VerifyFuncUpdated"

// IsAuthorizedAddressForName ensures that the given address has ownership of the given name.
func IsAuthorizedAddressForName(address std.Address, name string) bool {
	return checkFunc(address, name)
}

// defaultVerifyFunction checks the store to see that the
// user has properly registered the given name.
// This function considers as valid an `address` that matches the `name`.
func defaultVerifier(address std.Address, name string) bool {
	if strings.TrimSpace(address.String()) == "" || strings.TrimSpace(name) == "" {
		return false
	}

	// Allow user with their own address as name
	// This enables pseudo-anon namespaces
	if address.String() == name {
		return true
	}

	// Can be a registered name or an alias
	userData, _ := users.ResolveName(name)
	if userData == nil {
		return false
	}

	/// XXX: add check for r/sys/teams down the line

	return userData.Addr() == address
}

// NewVerifyCallExecutor allows updating the verifier function via a GovDAO proposal
func NewVerifyCallExecutor(newVerifyCall func(address std.Address, name string) bool) dao.Executor {
	callback := func() error {
		checkFunc = newVerifyCall

		std.Emit(VerifyFuncUpdatedEvent)
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(callback)
}
