// Package names enforces namespace permissions for package deployment.
//
// Two namespace shapes grant deploy authority when enforcement is enabled:
//
//  1. PA (personal-address) namespaces — gno.land/{r,p}/<addr>/* — the
//     deployer's address string equals the namespace literal. Anyone can
//     deploy under their own address.
//
//  2. Registered-name namespaces — gno.land/{r,p}/<name>/* — r/sys/users
//     has a (name → addr) mapping where the resolved address equals the
//     deployer AND the name is the user's CURRENT name (not a historical
//     alias from a rename chain). This is the bridge that lets
//     r/sys/namereg/v1 (or any other DAO-whitelisted controller) grant
//     deploy authority via name registration.
//
// Authority is unscoped: a registered name owns BOTH r/<name>/* and
// p/<name>/* paths. There is no sub-prefix isolation (e.g. r/u/<name>/*).
package names

import (
	"chain/runtime"

	susers "gno.land/r/sys/users"
)

var (
	admin   = address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh") // govdao t1 multisig
	enabled = false
)

// nameLookupFn returns (addr, ok) for a given registered name. Allows
// the verifier function to be unit-tested without wiring up r/sys/users
// state — production binds resolveCurrentName as the lookup, tests pass
// nil (PA-only) or a fake.
type nameLookupFn func(name string) (addr address, ok bool)

// IsAuthorizedAddressForNamespace checks if the given address can deploy
// to the given namespace. See package doc for the two authorization paths.
//
// Pre-Enable, all checks pass (testing/dev convenience).
func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	return verifier(enabled, address_XXX, namespace, resolveCurrentName)
}

// resolveCurrentName is the production nameLookupFn, backed by r/sys/users.
// Returns ok=true only if the name resolves to a non-deleted user AND the
// queried name is that user's CURRENT name (the most recent UpdateName).
//
// Restricting to the current name has two consequences worth knowing:
//
//  1. After a UpdateName from "alice" to "alice2", the user keeps deploy
//     authority over r/alice2/* but LOSES it for r/alice/*. Already-
//     deployed packages at r/alice/* keep working — deploy-time
//     authorization doesn't unwind the past — but no NEW deploys can
//     land there.
//
//  2. The old name "alice" is also unregisterable by anyone else: when
//     r/sys/users.UpdateName runs, it inserts the new name into nameStore
//     but does not remove the old one. r/sys/users.RegisterUser then
//     rejects re-registration of "alice" with ErrNameTaken. Net effect:
//     a rename permanently removes the old name from circulation.
//
// The alternative (allow historical aliases to retain authority) was
// rejected because it lets a single user register one cheap name, then
// rename N times to claim authority over N distinct namespaces — a
// stealth namespace acquisition vector worse than the current
// burn-on-rename behavior.
func resolveCurrentName(name string) (address, bool) {
	data, isCurrent := susers.ResolveName(name)
	if data == nil || !isCurrent {
		return "", false
	}
	return data.Addr(), true
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
// lookup is the registered-name resolver — pass nil to disable that path
// (used by tests that want to exercise PA-only behavior).
func verifier(isEnabled bool, address_XXX address, namespace string, lookup nameLookupFn) bool {
	if !isEnabled {
		return true // pre-genesis / dev convenience: bypass both paths
	}

	if namespace == "" || !address_XXX.IsValid() {
		return false
	}

	// Path 1: PA (personal-address) namespace.
	// gno.land/{p,r}/{ADDRESS}/**
	if address_XXX.String() == namespace {
		return true
	}

	// Path 2: registered-name namespace via r/sys/users.
	if lookup != nil {
		if owner, ok := lookup(namespace); ok && owner == address_XXX {
			return true
		}
	}

	return false
}
