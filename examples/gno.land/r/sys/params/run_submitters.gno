package params

// Delegated management of vm:p:run_submitters, the allowlist of addresses
// permitted to send MsgRun.
//
// See delegate.gno for why this is a named slot rather than a registry.
//
// What the delegate can and cannot do, and why the asymmetry is this way round:
//
//   - It may ADD addresses. This is the routine work the delegation exists for.
//   - It may REMOVE only addresses it added itself. De-listing its own mistake
//     is a core part of managing an allowlist, so add-only would be a strange
//     capability to hand out. Scoping removal to its own grants is what keeps it
//     safe: the delegate can never remove an address that was present before the
//     delegation, which is what stops it emptying the list. That matters more
//     than it looks — GovDAO proposal creation is MsgRun-only, so a delegate
//     that could empty this list could stop anyone from creating the proposal
//     that revokes it. Permanent brick.
//   - GovDAO retains everything, including the existing whole-list setter. That
//     is the bounded reset: one proposal returns the key to a known-good list
//     regardless of what the delegate did, and the proposal shows voters the
//     exact resulting list.
//
// A cost to state plainly rather than bury: this key is read by the ante handler
// on EVERY transaction, before the per-tx gas meter exists, so its length is an
// unmetered per-transaction constant for the whole chain. The delegate therefore
// holds a knob on that constant, bounded only by maxAddressListLen in
// gno.land/pkg/sdk/vm/params.go. That bound is enforced on this path for free:
// UpdateSysParamStrings re-sets the whole list, which re-enters WillSetParam and
// Params.Validate, so both the cap and bech32 validation apply to a delegate's
// additions. A key that is not read on the ante path would be a cheaper first
// delegation; this one is the one that was asked for.

import (
	"chain"
	prms "sys/params"

	"gno.land/p/moul/addrset"

	"gno.land/r/gov/dao"
)

const (
	vmModulePrefix    = "vm"
	vmParamsSubmodule = "p"

	runSubmittersKey = "run_submitters"
)

// runSubmittersMgr is the package path authorized to manage run_submitters.
// Empty means the capability is not delegated, and empty must deny — see
// assertDelegate.
var runSubmittersMgr string

// runSubmittersGrants records which addresses the current delegate added, so
// removal can be scoped to its own grants.
//
// The parameter is the source of truth and this is a side table, so the two can
// disagree — genesis, `gnogenesis params set`, or any future direct keeper write
// produces entries with no grant recorded. That direction is safe: an unrecorded
// address is simply not removable by the delegate, which is the conservative
// answer. Cleared whenever the delegation changes hands, so a new delegate never
// inherits authority over its predecessor's grants.
var runSubmittersGrants = addrset.Set{}

// RunSubmittersManager returns the package path currently authorized to manage
// run_submitters, or "" when the capability is not delegated.
func RunSubmittersManager() string {
	return runSubmittersMgr
}

// RunSubmittersGrantedBy reports whether the current delegate added addr, i.e.
// whether it may remove it.
//
// Exposed so a delegate can check before acting. A delegate that discovers a
// refusal by panicking mid-proposal-execution is in a bad place: the panic
// aborts the transaction, and for a DAO whose proposal has already passed, every
// retry aborts the same way.
func RunSubmittersGrantedBy(addr address) bool {
	return runSubmittersGrants.Has(addr)
}

// IsRunSubmittersDelegate reports whether pkgpath currently holds the
// capability. Pure predicate, for a caller that wants to fail cleanly rather
// than be panicked at.
func IsRunSubmittersDelegate(pkgpath string) bool {
	return pkgpath != "" && pkgpath == runSubmittersMgr
}

// ProposeSetRunSubmittersManager creates a GovDAO proposal handing management of
// run_submitters to pkgpath.
//
// pkgpath may be a sub-realm identity such as
// "gno.land/r/nt/commondao/v0#dao/42", which is how a single DAO hosted by a
// multi-tenant realm is named. Matching is exact, so naming the bare host would
// authorize the host itself and none of its DAOs.
func ProposeSetRunSubmittersManager(cur realm, pkgpath string) dao.ProposalRequest {
	assertDelegatePath(pkgpath)
	if pkgpath == runSubmittersMgr {
		panic("no-op proposal rejected: " + pkgpath + " already manages " + runSubmittersKey)
	}

	// desc uses the manager as of proposal creation, which is the honest thing
	// to show a voter. The EVENT reads it again inside the executor, because two
	// proposals created while the slot is empty and executed in sequence would
	// otherwise both report from="" while the second actually replaced the first.
	from := runSubmittersMgr
	desc := "Authorize " + pkgpath + " to add addresses to the " + runSubmittersKey +
		" allowlist, which gates who may send MsgRun. It may remove only " +
		"addresses it added itself. GovDAO retains full control, including " +
		"replacing the whole list."
	if from != "" {
		desc += " This replaces the current manager, " + from +
			", and discards the record of which addresses it granted."
	}

	return newDelegateProposal(cur, DelegateSetEvent, runSubmittersKey, pkgpath,
		"Delegate "+runSubmittersKey+" management", desc,
		func() string {
			prev := runSubmittersMgr
			runSubmittersMgr = pkgpath
			// A new holder must not inherit removal authority over addresses
			// the previous one granted.
			runSubmittersGrants = addrset.Set{}
			return prev
		})
}

// ProposeClearRunSubmittersManager creates a GovDAO proposal revoking the
// delegation.
//
// Revocation is immediate on execution because the slot is consulted on every
// call. It deliberately does NOT remove addresses the delegate added: sweeping
// them would make the executed effect invisible at vote time, and would silently
// remove nothing whenever the grant record had drifted. Use the existing
// whole-list setter to reset the list to a reviewed value.
func ProposeClearRunSubmittersManager(cur realm) dao.ProposalRequest {
	if runSubmittersMgr == "" {
		panic("no-op proposal rejected: " + runSubmittersKey + " is not delegated")
	}

	from := runSubmittersMgr
	return newDelegateProposal(cur, DelegateClearedEvent, runSubmittersKey, "",
		"Revoke "+runSubmittersKey+" management",
		"Revoke "+from+"'s authority to manage the "+runSubmittersKey+
			" allowlist. Addresses it already added REMAIN on the list; reset the "+
			"list explicitly if that is not wanted.",
		func() string {
			prev := runSubmittersMgr
			runSubmittersMgr = ""
			runSubmittersGrants = addrset.Set{}
			return prev
		})
}

// AddRunSubmitters adds addresses to the run_submitters allowlist.
//
// Callable only by the delegated manager. Addresses already present are a no-op
// (UpdateSysParamStrings dedupes), and the chain still validates every entry and
// enforces the list-length cap, because the update re-sets the whole list.
func AddRunSubmitters(cur realm, addrs []string) {
	assertDelegate(0, cur, runSubmittersMgr, "the "+runSubmittersKey+" allowlist")
	if len(addrs) == 0 {
		return
	}

	// Record a grant only for an address this call actually ADDED.
	//
	// UpdateSysParamStrings dedupes on add, so passing an address already on the
	// list leaves the parameter unchanged. Recording a grant for it anyway would
	// let the delegate launder authority over entries it never granted: read the
	// list, re-add all of it (a no-op on the parameter, but every address now
	// recorded as its own), then remove all of it. That empties the allowlist,
	// including whatever predated the delegation -- and since GovDAO proposal
	// creation is MsgRun-only, an empty allowlist means no proposal can be
	// created to revoke the delegate or restore the list. Chain brick, recoverable
	// only by relaunch.
	//
	// So the grant record has to follow the parameter, not the argument.
	present := make(map[string]bool)
	for _, a := range GetRunSubmitters() {
		present[a] = true
	}

	prms.UpdateSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey, addrs, true)
	for _, a := range addrs {
		if !present[a] {
			runSubmittersGrants.Add(address(a))
		}
	}
	chain.Emit(DelegateWriteEvent,
		"key", runSubmittersKey, "realm", runSubmittersMgr, "op", "add")
}

// RemoveRunSubmitters removes addresses from the run_submitters allowlist.
//
// Callable only by the delegated manager, and only for addresses that manager
// added. Refusing rather than silently skipping is deliberate: a partial removal
// that reported success would leave the caller believing an address was
// de-listed when it was not.
func RemoveRunSubmitters(cur realm, addrs []string) {
	assertDelegate(0, cur, runSubmittersMgr, "the "+runSubmittersKey+" allowlist")
	if len(addrs) == 0 {
		return
	}
	for _, a := range addrs {
		if !runSubmittersGrants.Has(address(a)) {
			panic("cannot remove " + a + ": not granted by " + runSubmittersMgr +
				", only GovDAO may remove it")
		}
	}

	prms.UpdateSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey, addrs, false)
	for _, a := range addrs {
		runSubmittersGrants.Remove(address(a))
	}
	chain.Emit(DelegateWriteEvent,
		"key", runSubmittersKey, "realm", runSubmittersMgr, "op", "remove")
}

// GetRunSubmitters returns the current allowlist.
func GetRunSubmitters() []string {
	vals, _ := prms.GetSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey)
	return vals
}
