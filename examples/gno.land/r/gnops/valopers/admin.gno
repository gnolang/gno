package valopers

import (
	"gno.land/p/moul/authz/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
)

var auth *authz.Authorizer

// Auth returns a read-only description of the realm's current governance
// authority (for rendering / inspection). It renders as
//
//	contract_authority[contract=gno.land/r/gov/dao,proposer=contract-identity]
//
// i.e. "GovDAO governs this realm, and only GovDAO may drive it" — a
// statement an on-chain reader can act on. The proposer half is
// load-bearing: without it the string is byte-identical whether the
// authority is gated or wide open (see ContractAuthority.String).
//
// It deliberately does NOT return the live *authz.Authorizer: an
// exported handle to the authority would let any realm reach its
// mutators (Transfer, DoByPrevious, ...) directly. Returning a
// description instead of the handle keeps that surface unreachable from
// here in the first place. Valoper.AuthOwner follows the same rule for
// per-operator auth lists.
func Auth() string {
	return auth.String()
}

// updateInstructions is the realm's only privileged write. It authorizes
// as the realm that CROSSED INTO valopers, not as valopers itself.
//
// Why DoByPrevious and not DoByCurrent: own-path + DoByCurrent is a
// tautology (see NewContractAuthority's godoc for the mechanism, once).
// `rlm.Previous()` is the caller — on the legitimate path r/gov/dao's
// executor frame, which is what init.gno's authority is pointed at.
//
// What it buys: it is the code-level guard against this realm's most
// likely regression, a maintainer later adding an exported crossing
// entrypoint that reaches here. Under DoByCurrent such an entrypoint
// authorizes for any caller; under DoByPrevious the caller's own address
// is the principal and the gate rejects it. Pinned by
// filetests/z_govdao_only_principal_filetest.gno.
//
// Two things it does NOT cover on its own, both closed elsewhere:
// re-exporting the privileged closure (guarded by the seal in
// NewInstructionsProposalRequest), and exporting a function whose
// signature is assignable to dao's `func(realm) error` callback type,
// which the proxy would invoke on an attacker's behalf. The second was a
// live bypass of this gate until SimpleExecutor.Execute began rejecting
// invocation from outside r/gov/dao (r/gov/dao/types.gno); keep BOTH
// conditions in mind before adding an exported symbol here.
//
// Returns the authorization error rather than panicking on it. That matters
// on the governance path: impl.ExecuteProposal turns an executor ERROR into
// status Denied plus a DeniedReason, but it cannot see a panic — a panic
// aborts the whole ExecuteProposal transaction, burning its gas and leaving
// the proposal stuck Accepted, retryable forever and never deniable. The
// refusal is currently unreachable on the approved route (Previous() there
// is always r/gov/dao), but init.gno's handler is explicitly offered as the
// seam for an intent check (timelock, quorum, audit trail), and the first
// such check to return an error would hit exactly that.
func updateInstructions(_ int, rlm realm, newInstructions string) error {
	return auth.DoByPrevious(0, rlm, "update-instructions", func() error {
		instructions = newInstructions
		return nil
	})
}

// NewInstructionsProposalRequest builds the GovDAO proposal that rewrites
// this realm's instructions. It replaces the previously exported
// NewInstructionsProposalCallback.
//
// SECURITY: the privileged closure must never leave this package.
//
// The VM mints a crossing frame's `cur` from the CALLEE's declaring
// package, so a closure declared here always runs with valopers'
// identity no matter who invokes it — and whoever holds the closure
// chooses its Previous() by wrapping it in an executor of their own.
// Handing such a closure to a caller therefore hands out the ability to
// satisfy this realm's gate whichever principal it is pointed at: the
// handler in init.gno runs the action and `instructions` is rewritten
// with no proposal and no vote. That is what the exported callback did.
//
// Returning a dao.ProposalRequest instead seals the capability: the
// executor is an unexported field with no accessor (dao.ProposalRequest
// exposes only Title/Description/Filter), so the only way to reach the
// closure is for GovDAO to execute the proposal that contains it.
//
// Note this is deliberately stronger than returning a dao.Executor —
// even a dao.NewSafeExecutor-wrapped one. Executor.Execute is an
// exported method, so a returned Executor stays directly invocable by
// its holder, and SafeExecutor's only gate (InAllowedDAOs) FAILS OPEN
// while the allowedDAOs list is empty, which is the documented
// bootstrap state (see r/gov/dao/loader/v0). A sealed request has no
// such conditional.
//
// The rule for this realm: no exported function may return a crossing
// closure, a dao.Executor, or anything else that carries valopers'
// frame identity — and no exported function may itself BE such a thing,
// i.e. have a signature assignable to dao's `func(realm) error` callback
// type, since the proxy would then invoke it on an attacker's behalf.
// filetests/z_foreign_realm_capability_filetest.gno pins the first;
// filetests/z_govdao_only_principal_filetest.gno pins the second.
func NewInstructionsProposalRequest(cur realm, newInstructions string) dao.ProposalRequest {
	cb := func(cur realm) error {
		return updateInstructions(0, cur, newInstructions)
	}

	title := instructionsProposalTitle
	description := ufmt.Sprintf("Update the instructions to: \n\n%s", newInstructions)

	return dao.NewProposalRequest(title, description, dao.NewSimpleExecutor(0, cur, cb, ""))
}

// instructionsProposalTitle is the on-chain title of the instructions
// proposal, named as a constant so a test can pin it: it is rendered to
// GovDAO voters and is the string off-chain tooling matches on. It was
// "/p/gnops/valopers: ..." before this realm's proposal construction moved
// here; the "/p/" was simply wrong (valopers is an /r/ realm).
const instructionsProposalTitle = "/r/gnops/valopers: Update instructions"

// rotateAuthority replaces this realm's governance authority.
//
// Unexported, and reachable only through the sealed request built by
// NewAuthorityRotationProposalRequest — same discipline as
// updateInstructions, for the same reason.
//
// authz.Authorizer.Transfer derives its principal as
// rlm.Previous().Address(), so inside a GovDAO-executed callback declared
// here Previous() is r/gov/dao — exactly the principal init.gno's authority
// asserts. No same-package cross() trick is needed.
func rotateAuthority(_ int, rlm realm, newContractPath string) error {
	return auth.Transfer(0, rlm, authz.NewContractAuthority(
		newContractPath,
		func(_ string, action authz.PrivilegedAction) error {
			return action()
		},
	))
}

// NewAuthorityRotationProposalRequest builds the GovDAO proposal that
// re-points this realm's governance authority at `newContractPath`.
//
// WHY THIS EXISTS. Without it the authority installed by init.gno is
// permanently frozen: nothing else calls auth.Transfer, Auth() returns a
// description rather than the live Authorizer, and Transfer routes through
// the authority's own gate — so if the asserted principal ever stops being
// presentable, `instructions` becomes unwritable with no on-chain recovery.
// That is not hypothetical: the r/gov/dao proxy path is a governance
// artifact that can be superseded, and a well-formed but wrong path (say
// the impl path instead of the proxy) is accepted at construction and dead
// forever. Redeploying instead is expensive — r/sys/validators/v0/cache.gno
// hardcodes `const valopersRealmPath`, so a new pkgpath means editing a
// second genesis realm and re-registering every valoper. It must exist
// before deploy, because code cannot be added to a deployed realm.
//
// WHY A PATH AND NOT AN authz.Authority. Taking an Authority would make
// this an open-interface input (Class-3 impl-substitution): a proposal
// could install an always-approve or always-deny authority, and readers of
// Auth() could not tell from the rendered description alone. Taking a path
// keeps the installed authority canonical by construction — always a
// ContractAuthority with the pass-through handler and the contract-identity
// gate — so the only thing governance can change is WHICH principal is
// asserted. The path is validated by authz.NewContractAuthority, which
// panics on a malformed one, aborting the proposal execution rather than
// bricking the realm.
//
// This remains a privileged surface: a majority that can pass this proposal
// can point the authority at a principal it controls. That is the same
// power a majority already has over `instructions`, and it is the price of
// being recoverable at all.
func NewAuthorityRotationProposalRequest(cur realm, newContractPath string) dao.ProposalRequest {
	cb := func(cur realm) error {
		return rotateAuthority(0, cur, newContractPath)
	}

	title := "/r/gnops/valopers: Rotate governance authority"
	description := ufmt.Sprintf(
		"Re-point the valopers governance authority at: \n\n%s\n\n"+
			"The authority stays a contract-identity ContractAuthority; only the asserted principal changes.",
		newContractPath,
	)

	return dao.NewProposalRequest(title, description, dao.NewSimpleExecutor(0, cur, cb, ""))
}

// NewInstructionsProposalCallback was removed, not deprecated: it was the
// capability leak described above, so an inert shim would be pointless and
// a working one would reopen the hole. Precedent + replay check: the
// min-fee callback below was already removed the same way even though it
// HAD been used on gnoland1 (GovDAO Prop #19), whereas the instructions
// callback has no on-chain trace at all — the live realm's instructions
// are byte-identical to the genesis default and none of gnoland1's 26
// GovDAO proposals is an "Update instructions" one.
//
// The min-fee callback was removed: the fee now lives in sysparams
// under node:valoper:register_fee, and
// proposal.ProposeNewMinFeeProposalRequest delegates to
// sys/params.NewSysParamUint64PropRequest. Removing avoids the
// forward-compat hazard of a no-op shim with no caller-auth gating.
