package params

// Delegated management of vm:p:run_submitters, the allowlist of addresses
// permitted to send MsgRun. See delegate.gno for why this is a named slot.
//
// The delegate may add addresses, and remove only those it added, never the
// last one: an empty list turns the gate off. GovDAO keeps full control through
// ProposeSetRunSubmitters.
//
// The ante handler reads this key on every transaction, before gas metering, so
// its length is capped by maxAddressListLen in gno.land/pkg/sdk/vm/params.go.
// UpdateSysParamStrings re-sets the whole list, so the cap binds the delegate.

import (
	"chain"
	prms "sys/params"

	"gno.land/p/moul/addrset/v0"

	"gno.land/r/gov/dao"
)

const (
	vmModulePrefix    = "vm"
	vmParamsSubmodule = "p"

	runSubmittersKey = "run_submitters"
)

// runSubmittersMgr is the package path authorized to manage run_submitters,
// empty when not delegated.
var runSubmittersMgr string

// runSubmittersGrants records which addresses the current delegate added, so
// it may remove only those. An address written by any other route has no grant
// and stays out of the delegate's reach. Cleared when the delegation changes.
var runSubmittersGrants = addrset.Set{}

// RunSubmittersManager returns the package path currently authorized to manage
// run_submitters, or "" when the capability is not delegated.
func RunSubmittersManager() string {
	return runSubmittersMgr
}

// RunSubmittersGrantedBy reports whether the current delegate added addr, and
// so may remove it. Lets a delegate check before acting rather than panic.
func RunSubmittersGrantedBy(addr address) bool {
	return runSubmittersGrants.Has(addr)
}

// IsRunSubmittersDelegate reports whether pkgpath currently holds the
// capability.
func IsRunSubmittersDelegate(pkgpath string) bool {
	return pkgpath != "" && pkgpath == runSubmittersMgr
}

// ProposeSetRunSubmittersManager creates a GovDAO proposal handing management of
// run_submitters to pkgpath.
//
// pkgpath may be a sub-realm identity such as
// "gno.land/r/nt/commondao/v0#dao/42". Matching is exact.
func ProposeSetRunSubmittersManager(cur realm, pkgpath string) dao.ProposalRequest {
	assertDelegatePath(pkgpath)
	if pkgpath == runSubmittersMgr {
		panic("no-op proposal rejected: " + pkgpath + " already manages " + runSubmittersKey)
	}

	// desc shows the manager at creation; the event reads it at execution.
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
			runSubmittersGrants = addrset.Set{}
			return prev
		})
}

// ProposeClearRunSubmittersManager creates a GovDAO proposal revoking the
// delegation.
//
// Addresses the delegate added stay listed; ProposeSetRunSubmitters resets the
// list.
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
// Callable only by the delegated manager. Addresses already present are a
// no-op.
func AddRunSubmitters(cur realm, addrs []string) {
	assertDelegate(0, cur, runSubmittersMgr, "the "+runSubmittersKey+" allowlist")
	if len(addrs) == 0 {
		return
	}

	// An empty list means the gate is off. Arming it is a GovDAO vote: a list
	// that shuts the members out cannot be undone, since proposing needs MsgRun.
	if len(GetRunSubmitters()) == 0 {
		panic("refusing to arm the " + runSubmittersKey +
			" allowlist: it is empty, so the gate is off and anyone may MsgRun. " +
			"Turning it on is a GovDAO vote, not a delegated edit")
	}

	// Grant only what this call added: re-adding a listed address must not hand
	// the delegate removal rights over it.
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
// Callable only by the delegated manager, and only for addresses it added. Any
// other address panics rather than being skipped.
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

	// The delegate may never empty the list, which would turn the gate off.
	// Counted against the parameter, since the argument may repeat or name
	// unlisted addresses.
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
// The only way to set the list by vote; the generic factories refuse the key.
//
// A non-empty list must name the proposer, who must be the signing user and a
// GovDAO member. A list naming nobody who can propose cannot be undone, since
// proposing needs MsgRun. Checked at creation, where a refusal can still be
// acted on; a panic in the executor would leave the proposal stuck.
func ProposeSetRunSubmitters(cur realm, addrs []string) dao.ProposalRequest {
	// IsUser, not IsUserCall: the MsgRun path runs in the signer's own
	// /e/<addr>/run realm, which IsUserCall would refuse.
	prev := cur.Previous()
	if !prev.IsUser() {
		panic("refusing to propose " + runSubmittersKey +
			" through a relaying realm: only the signing user holds a key")
	}
	proposer := prev.Address().String()
	// Asked of the current implementation, not a fixed member store, so
	// the floor follows a DAO upgrade.
	if !dao.IsMember(prev.Address()) {
		panic("refusing to build a " + runSubmittersKey + " proposal for " +
			proposer + ": the builder is not a GovDAO member, and a list " +
			"naming only a non-member cannot create the proposal that changes it back")
	}
	// Snapshot: the caller can still write the array it passed.
	addrs = copyStrings(addrs)
	listed := false
	for _, a := range addrs {
		if a == proposer {
			listed = true
			break
		}
	}
	if len(addrs) > 0 && !listed {
		panic("refusing to propose a " + runSubmittersKey + " allowlist that omits " +
			"the proposer " + proposer + ": a non-empty list that names nobody " +
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
		chain.Emit("SetRunSubmitters", "key", runSubmittersKey, "proposer", proposer)
		return nil
	}
	return dao.NewProposalRequest(title, desc,
		dao.NewSimpleExecutor(0, cur, callback, ""))
}

// setRunSubmitters replaces the whole allowlist. Set, not Update: an append
// could never remove an address or empty the list.
func setRunSubmitters(addrs []string) {
	prms.SetSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey, addrs)
}
