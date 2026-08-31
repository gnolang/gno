package params

// Delegated management of vm:p:run_submitters, the allowlist of addresses
// permitted to send MsgRun.
//
// See delegate.gno for why this is a named slot rather than a registry.
//
// What the delegate can and cannot do, and why the asymmetry is this way round:
//
//   - It may ADD addresses. This is the routine work the delegation exists for.
//   - It may REMOVE only addresses it added itself, and never the last one.
//     De-listing its own mistake is a core part of managing an allowlist, so
//     add-only would be a strange capability to hand out. Two bounds keep that
//     safe. Grant-scoping stops it removing an address that predates the
//     delegation. The non-empty floor in RemoveRunSubmitters stops it reaching
//     zero by any route — an empty list means the gate is OFF and anyone may
//     MsgRun, so emptying it would let the delegate revoke the entire
//     restriction GovDAO voted for, which is the one thing this capability must
//     not be able to do.
//   - GovDAO retains everything, through ProposeSetRunSubmitters below -- the
//     generic factories no longer accept this key. That is the bounded reset:
//     one proposal returns the key to a known-good list regardless of what the
//     delegate did, and the proposal shows voters the exact resulting list.
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

	// The delegate may curate a list that is already in force. It may not put
	// one into force.
	//
	// An empty run_submitters means the allowlist is OFF and anyone may MsgRun.
	// So the first add is not curation: it switches a chain-wide restriction on
	// and picks who it admits. A delegate adding one address to an empty list
	// leaves that address the only one on the chain that may run code.
	//
	// Refused rather than discouraged because it is unrepairable. Creating a
	// GovDAO proposal needs MsgRun -- a ProposalRequest carries an Executor,
	// which MsgCall cannot build from string arguments -- so once the gate is
	// armed against the members they cannot propose the vote that would undo
	// it, and the floor in RemoveRunSubmitters stops the delegate undoing it
	// either.
	if len(GetRunSubmitters()) == 0 {
		panic("refusing to arm the " + runSubmittersKey +
			" allowlist: it is empty, so the gate is off and anyone may MsgRun. " +
			"Turning it on is a GovDAO vote, not a delegated edit")
	}

	// Record a grant only for an address this call actually ADDED.
	//
	// UpdateSysParamStrings dedupes on add, so passing an address already on the
	// list leaves the parameter unchanged. Recording a grant for it anyway would
	// let the delegate launder authority over entries it never granted: read the
	// list, re-add all of it (a no-op on the parameter, but every address now
	// recorded as its own), then remove all of it -- including whatever predated
	// the delegation. The floor in RemoveRunSubmitters would refuse the last of
	// those removals, but only the last: everything up to it would still go
	// through, leaving the delegate holding the only listed address and so the
	// sole authority over who may MsgRun.
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

	// The delegate may not empty the list, whatever it granted.
	//
	// An empty run_submitters means the gate is OFF -- anyone on the chain may
	// send MsgRun. So emptying it is not a smaller version of removing one
	// address, it is the opposite of what the delegation is for: it would let a
	// delegate authorized to curate a list unilaterally revoke the whole
	// restriction GovDAO voted for.
	//
	// Grant-scoping alone does not prevent this. It holds only while at least
	// one entry the delegate did not grant survives, and GovDAO replacing the
	// list wholesale can remove its own entries without touching the grant
	// record. A floor makes the invariant structural instead of emergent.
	//
	// Counted against the parameter, not the argument: the caller may name
	// addresses that are not listed, or name one twice, and neither shrinks the
	// list. Only GovDAO can go to zero, through the whole-list setter, where the
	// resulting list is on the ballot.
	removing := make(map[string]bool, len(addrs))
	for _, a := range addrs {
		removing[a] = true
	}
	remaining := 0
	for _, a := range GetRunSubmitters() {
		if !removing[a] {
			remaining++
		}
	}
	if remaining == 0 {
		panic("refusing to empty the " + runSubmittersKey +
			" allowlist: an empty list disables the gate entirely, so only " +
			"GovDAO may do it")
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

// ProposeSetRunSubmitters creates a GovDAO proposal replacing the whole
// run_submitters allowlist.
//
// This is the only way to set the list by vote: the generic factories refuse
// the key (see assertNotRunSubmittersKey), so every whole-list write comes
// through here and carries the rule below.
//
// The proposer must be on the list they propose.
//
// An empty run_submitters means the gate is off and anyone may MsgRun. A
// non-empty one therefore decides who may run code at all -- and a list naming
// nobody who can create a GovDAO proposal cannot be undone, because creating a
// proposal needs MsgRun: a ProposalRequest carries an Executor, and MsgCall
// cannot build one from string arguments. The vote would end governance.
//
// Requiring the proposer's own address is a cheap way to prove the list is
// usable rather than merely plausible. GovDAO refuses a proposal from a
// non-member (PreCreateProposal, "only members can create new proposals"), so
// if this proposal exists at all its author is a member -- and they just signed
// the transaction that created it, so the address demonstrably holds a key. A
// list that merely NAMES a member proves neither: the address may belong to
// nobody, since any member can enroll an arbitrary address.
//
// Checked here, at proposal creation, rather than inside the executor. Nothing
// in r/gov/dao recovers from an executor panic, so a check that fires at
// execution turns a passed proposal into one that can never be executed. Here
// the refusal reaches a person who can still fix the list and propose again.
//
// This is a floor, not an invariant. The proposer may resign from GovDAO later,
// and the list is not re-checked when they do. It rules out arriving at a dead
// list in one vote; it cannot rule out drifting into one.
func ProposeSetRunSubmitters(cur realm, addrs []string) dao.ProposalRequest {
	// IsCurrent before Previous, as AGENTS.md requires and assertDelegate does.
	//
	// Belt and braces here rather than load-bearing: the compiler already
	// refuses anything but `cur` or `cross(rlm)` as the first argument to a
	// crossing function, so this cur is live by construction and the check
	// cannot fire. It earns its place if this ever becomes a helper taking an
	// ordinary realm parameter, which is the shape assertDelegate has and where
	// a stashed value really can be threaded in.
	if !cur.IsCurrent() {
		panic("unauthorized: cur is not the caller's live realm")
	}
	proposer := cur.Previous().Address()
	listed := false
	for _, a := range addrs {
		if a == proposer.String() {
			listed = true
			break
		}
	}
	if len(addrs) > 0 && !listed {
		panic("refusing to propose a " + runSubmittersKey + " allowlist that omits " +
			"the proposer " + proposer.String() + ": a non-empty list that names nobody " +
			"who can create a proposal cannot be changed back")
	}

	title := "Set the " + runSubmittersKey + " allowlist"
	desc := "Replace the " + runSubmittersKey + " allowlist, which gates who may " +
		"send MsgRun."
	if len(addrs) == 0 {
		desc += " This empties the list, which switches the gate OFF: anyone may " +
			"send MsgRun."
	} else {
		desc += " Only these addresses may send MsgRun. Creating a GovDAO proposal " +
			"needs MsgRun, so this list also decides who can govern."
		for _, a := range addrs {
			desc += "\n- " + a
		}
	}

	callback := func(cur realm) error {
		setRunSubmitters(addrs)
		chain.Emit("SetRunSubmitters", "key", runSubmittersKey, "proposer", proposer.String())
		return nil
	}
	return dao.NewProposalRequest(title, desc,
		dao.NewSimpleExecutor(0, cur, callback, ""))
}

// setRunSubmitters replaces the whole allowlist.
//
// SetSysParamStrings, not UpdateSysParamStrings: Update with add=true appends
// non-duplicates onto what is already there, so it can never remove an address
// or empty the list. This realm reserves the key from the generic factories, so
// this is the only route by vote -- if it appended, the parameter would be
// append-only chain-wide, a compromised address could never be de-listed, and
// the gate could never be turned back off.
//
// A named function rather than the executor's body inline, so a test can reach
// it: a closure held in a ProposalRequest cannot be called from outside.
func setRunSubmitters(addrs []string) {
	prms.SetSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey, addrs)
}
