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
//
// The realm exposes an emergency-halt switch via SetPaused. When paused,
// the verifier rejects EVERY namespace check — PA included — until
// unpaused. This is the "true emergency" semantic; the narrow alternative
// (pause registered-name only, preserve PA) was considered and rejected
// because the threats most likely to justify pausing this realm
// (verifier bug, compromised controller, signature-layer incident) do
// not reliably exempt PA from the same blast radius.
package names

import (
	"chain"
	"chain/runtime"

	"gno.land/r/gov/dao"
	govimpl "gno.land/r/gov/dao/v3/impl"
	memberstore "gno.land/r/gov/dao/v3/memberstore"
	susers "gno.land/r/sys/users"
)

var (
	// admin is the GovDAO T1 multisig address, hardcoded at realm-source
	// commit time. Its only capability is gating Enable() — a one-way,
	// one-shot genesis activation of the namespace verifier. The address
	// has no other authority on this realm; pause/unpause is gated on a
	// separate GovDAO T1 proposal (see ProposeSetPaused), and there is
	// no SetEnabled(false) or SetAdmin path.
	//
	// Hardcoding is acceptable because:
	//   - Enable() is called once, at chain genesis. After that the
	//     address is dead weight — no further capability flows through it.
	//   - The narrow blast radius of "stale admin" is "Enable() can never
	//     be called", which leaves the verifier in pre-Enable bypass mode
	//     (returns true for all checks) — degraded but not exploitable.
	//   - A rotation path would only matter if Enable() needed to be
	//     re-issued. It doesn't; the flag is sticky.
	//
	// If the genesis activation pattern ever needs to change (e.g. a
	// SetEnabled(false) emergency disable is added later), this admin
	// model should be replaced with a GovDAO T1 proposal flow that
	// mirrors ProposeSetPaused. Until then, the hardcoded address is
	// the smallest viable governance surface for the one-shot use case.
	admin   = address("g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh")
	enabled = false
	paused  = false
)

// nameLookupFn returns (addr, ok) for a given registered name. Allows
// the verifier function to be unit-tested without wiring up r/sys/users
// state — production binds resolveCurrentName as the lookup, tests pass
// nil (PA-only) or a fake.
type nameLookupFn func(name string) (addr address, ok bool)

// IsAuthorizedAddressForNamespace checks if the given address can deploy
// to the given namespace. See package doc for the two authorization paths
// and the pause semantic.
//
// Pre-Enable, all checks pass (testing/dev convenience).
func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	return verifier(enabled, paused, address_XXX, namespace, resolveCurrentName)
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

// ProposeSetPaused returns a GovDAO proposal request that, when voted
// through and executed, toggles the chain-wide deploy gate. When the
// realm is paused, the verifier rejects EVERY namespace check — PA
// (personal-address) included — until a subsequent ProposeSetPaused(false)
// proposal executes.
//
// This is an emergency halt. A paused state means NO new MsgAddPackage
// transactions land at any path on the chain. Existing realms continue
// to receive MsgCall traffic normally — pause is scoped to addpkg, not
// to all VM operations. Use cases:
//   - Bug discovered in this realm or r/sys/users that requires a
//     hotfix before further deploys can be trusted.
//   - Wallet/signature-layer incident under investigation.
//
// (A "compromised controller" use case was considered and removed: the
// controller's RegisterUser path is direct into r/sys/users and does
// NOT go through this verifier, so pause does not freeze new
// registrations. To contain a compromised controller, the appropriate
// flow is ProposeControllerRemoval in r/sys/users, not pause here.)
//
// The narrow alternative (pause registered-name path only, preserve
// PA) was considered and rejected. See package doc for rationale.
//
// Gated on GovDAO proposal at T1 tier — not the hardcoded admin used
// by Enable. Pause is consequential enough to warrant a tier-restricted
// governance vote rather than a single-multisig click. T1 filter
// prevents lower-tier members from spamming pause proposals to dilute
// attention. Trade-off is response time: a T1 vote takes hours-to-days;
// if a faster emergency-halt mechanism is needed, that belongs at a
// different layer (e.g. an ante-handler-level chain pause), not here.
//
// Pause is orthogonal to the pre-Enable bypass: before Enable, the
// verifier returns true regardless of paused state. So executing a
// pause proposal before Enable has no effect on deploys, but the
// value persists and applies the moment Enable runs. To avoid this
// staging trap, operators should call Enable BEFORE any pause
// proposals are voted in.
//
// Idempotency: calling ProposeSetPaused(v) when the realm's current
// paused state already equals v panics at proposal-creation time so
// voters never see a proposal whose execution would no-op.
func ProposeSetPaused(v bool) dao.ProposalRequest {
	if paused == v {
		panic("paused state already matches requested value; no-op proposal rejected")
	}
	cb := func(cur realm) error {
		setPaused(v)
		return nil
	}
	title := "Unpause Namespace Verifier"
	desc := "This proposal unpauses `r/sys/names`. After execution, the namespace verifier will resume normal authorization checks (PA + registered-name paths). MsgCall traffic to existing realms is unaffected (pause was scoped to MsgAddPackage); only NEW package deploys gate on the unpaused state."
	if v {
		title = "Pause Namespace Verifier"
		desc = "This proposal pauses `r/sys/names`. After execution, the namespace verifier will reject EVERY new MsgAddPackage on the chain — PA (personal-address) and registered-name namespaces alike — until a subsequent unpause proposal executes. This is an emergency halt scoped to addpkg; MsgCall traffic to existing realms is unaffected."
	}
	return dao.NewProposalRequestWithFilter(
		title,
		desc,
		dao.NewSimpleExecutor(cb, ""),
		govimpl.FilterByTier{Tier: memberstore.T1},
	)
}

// setPaused is the private actuator behind ProposeSetPaused's executor
// callback. It is unexported to ensure no path outside the proposal
// flow can flip the flag — every state change goes through GovDAO.
//
// Emits NamespaceEnforcement{Paused,Unpaused} with the executor's
// realm path for off-chain audit trails.
func setPaused(v bool) {
	paused = v
	if v {
		chain.Emit("NamespaceEnforcementPaused", "by", runtime.PreviousRealm().PkgPath())
	} else {
		chain.Emit("NamespaceEnforcementUnpaused", "by", runtime.PreviousRealm().PkgPath())
	}
}

// IsPaused reports the current value of the pause flag. Note: when
// the realm is pre-Enable, IsPaused may return true but the verifier
// will still pass-through (pre-Enable bypass takes priority).
func IsPaused() bool {
	return paused
}

// verifier checks namespace deployment permissions.
// lookup is the registered-name resolver — pass nil to disable that path
// (used by tests that want to exercise PA-only behavior).
//
// Order of checks (top to bottom, first matching wins):
//  1. !isEnabled → return true   (pre-Enable: testing/dev bypass)
//  2. isPaused   → return false  (emergency halt — INCLUDES PA)
//  3. invalid input → return false
//  4. PA match (addr.String() == namespace) → return true
//  5. Registered-name lookup match → return true
//  6. otherwise → return false
//
// The pause check is intentionally above the PA check. A SetPaused(true)
// halts every deploy regardless of namespace shape.
func verifier(isEnabled, isPaused bool, address_XXX address, namespace string, lookup nameLookupFn) bool {
	if !isEnabled {
		return true // pre-genesis / dev convenience: bypass everything
	}

	if isPaused {
		return false // emergency halt: reject every deploy including PA
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
