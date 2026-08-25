package params

import (
	"testing"

	prms "sys/params"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

const (
	testDelegate = "gno.land/r/test/delegate"
	testOther    = "gno.land/r/test/other"
	// A sub-realm identity, the form a single DAO hosted by a multi-tenant
	// realm presents. "#" cannot occur in a real package path.
	testSubDelegate = "gno.land/r/nt/commondao/v0#dao/42"
)

// resetDelegation returns the slot to undelegated. The realm's tests share
// package state, so anything touching the slot must put it back.
func resetDelegation() {
	runSubmittersMgr = ""
	runSubmittersGrants = addrset.Set{}
}

// armRunSubmitters puts addresses on the list the way GovDAO does: by writing
// the parameter, not through the delegate.
//
// Tests need it because the delegate may no longer arm an empty list. An empty
// run_submitters means the gate is off, so the first add would switch a
// chain-wide restriction on rather than curate one, and that is a vote.
func armRunSubmitters(addrs ...string) {
	prms.UpdateSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey, addrs, true)
}

// unarmRunSubmitters drops addresses the way a GovDAO vote does, bypassing the
// delegate's grant-scoping. Used to build a list whose every remaining entry was
// granted by the delegate, which is the only state where the non-empty floor is
// the binding constraint.
func unarmRunSubmitters(addrs ...string) {
	prms.UpdateSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey, addrs, false)
}

// TestRunSubmittersUndelegatedDeniesEveryone is the single most important test
// here: with no delegate configured, the slot is "" and a direct user call also
// presents "". A gate that compared the two without checking for empty first
// would admit every account on the chain.
func TestRunSubmittersUndelegatedDeniesEveryone(cur realm, t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	uassert.Equal(t, "", RunSubmittersManager())

	// A user account: PkgPath() is empty, matching the empty slot.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("nobody")))
	uassert.AbortsContains(t, cur, "no delegate is configured", func() {
		AddRunSubmitters(cross(cur), []string{testutils.TestAddress("victim").String()})
	})

	// And a code realm, for completeness.
	testing.SetRealm(testing.NewCodeRealm(testOther))
	uassert.AbortsContains(t, cur, "no delegate is configured", func() {
		AddRunSubmitters(cross(cur), []string{testutils.TestAddress("victim").String()})
	})
}

// TestRunSubmittersOnlyTheDelegateMayWrite pins that authorization is by exact
// package path.
func TestRunSubmittersOnlyTheDelegateMayWrite(cur realm, t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	runSubmittersMgr = testDelegate
	armRunSubmitters(testutils.TestAddress("seeded-by-vote").String())
	addr := testutils.TestAddress("granted").String()

	// A different realm is refused.
	testing.SetRealm(testing.NewCodeRealm(testOther))
	uassert.AbortsContains(t, cur, "unauthorized", func() {
		AddRunSubmitters(cross(cur), []string{addr})
	})

	// A user account is refused: an empty path must not match a set slot.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("nobody")))
	uassert.AbortsContains(t, cur, "unauthorized", func() {
		AddRunSubmitters(cross(cur), []string{addr})
	})

	// The delegate itself succeeds.
	testing.SetRealm(testing.NewCodeRealm(testDelegate))
	AddRunSubmitters(cross(cur), []string{addr})
	uassert.True(t, contains(GetRunSubmitters(), addr),
		"the delegate's addition must reach the parameter")
}

// TestRunSubmittersSubRealmIdentityIsExact pins that delegate matching is exact
// string equality, which is what makes a per-DAO delegation safe.
//
// A sub-realm identity minted by cur.Sub(subpath) presents "host#subpath", so a
// single DAO hosted by a multi-tenant realm can hold the capability. Matching by
// prefix instead would hand it to every DAO that host serves -- and CommonDAO
// membership, while invite-gated, is unlimited once invited.
//
// Asserted through the pure predicate rather than by crossing: constructing a
// live sub-realm cur is not something the test harness can do (NewCodeRealm
// rejects "#", and MakeRealm does not satisfy IsCurrent), and the comparison
// under test is the same one assertDelegate performs.
func TestRunSubmittersSubRealmIdentityIsExact(t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	runSubmittersMgr = testSubDelegate

	uassert.True(t, IsRunSubmittersDelegate(testSubDelegate),
		"the named DAO holds the capability")
	uassert.False(t, IsRunSubmittersDelegate("gno.land/r/nt/commondao/v0"),
		"the host realm must not inherit its sub-identity's authority")
	uassert.False(t, IsRunSubmittersDelegate("gno.land/r/nt/commondao/v0#dao/43"),
		"a sibling DAO of the same host must not match")
	uassert.False(t, IsRunSubmittersDelegate("gno.land/r/nt/commondao/v0#dao/4"),
		"a prefix of the subpath must not match")
	uassert.False(t, IsRunSubmittersDelegate(""),
		"an empty path must never match a configured delegate")
}

// TestRunSubmittersRemoveIsScopedToOwnGrants pins that the delegate cannot
// remove an address it did not add, so a list GovDAO curated survives a rogue
// delegate. The companion bound -- that it can never take the list to zero by
// any route -- is TestRunSubmittersCannotEmptyTheList.
func TestRunSubmittersRemoveIsScopedToOwnGrants(cur realm, t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	preexisting := testutils.TestAddress("breakglass").String()
	testing.SetSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey,
		[]string{preexisting})

	runSubmittersMgr = testDelegate
	armRunSubmitters(testutils.TestAddress("seeded-by-vote").String())
	own := testutils.TestAddress("ownadd").String()

	testing.SetRealm(testing.NewCodeRealm(testDelegate))
	AddRunSubmitters(cross(cur), []string{own})

	// It may remove what it granted.
	uassert.True(t, RunSubmittersGrantedBy(address(own)))
	RemoveRunSubmitters(cross(cur), []string{own})
	uassert.False(t, contains(GetRunSubmitters(), own))

	// It may NOT remove the address that predated the delegation.
	uassert.False(t, RunSubmittersGrantedBy(address(preexisting)))
	uassert.AbortsContains(t, cur, "only GovDAO may remove it", func() {
		RemoveRunSubmitters(cross(cur), []string{preexisting})
	})
	uassert.True(t, contains(GetRunSubmitters(), preexisting),
		"the pre-existing entry must survive a refused removal")
}

// TestRunSubmittersCannotEmptyTheList pins the non-empty floor.
//
// An empty run_submitters means the gate is OFF: anyone on the chain may
// MsgRun. So emptying the list is not a smaller version of removing one
// address, it is the unilateral revocation of the whole restriction GovDAO
// voted for -- the one thing this capability must not be able to do.
//
// Grant-scoping alone does not prevent it. It holds only while an entry the
// delegate did not grant survives, and GovDAO replacing the list wholesale can
// remove its own entries without touching the grant record, which is exactly
// the state set up below.
func TestRunSubmittersCannotEmptyTheList(cur realm, t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	runSubmittersMgr = testDelegate
	seed := testutils.TestAddress("seeded-by-vote").String()
	armRunSubmitters(seed)
	a := testutils.TestAddress("granted-a").String()
	b := testutils.TestAddress("granted-b").String()

	testing.SetRealm(testing.NewCodeRealm(testDelegate))
	AddRunSubmitters(cross(cur), []string{a, b})

	// GovDAO then drops its own seed by vote, which grant-scoping does not
	// constrain. This is the only way to reach a list whose every entry was
	// granted by the delegate, now that the delegate cannot arm an empty one --
	// and it is the state where the floor is the last thing standing.
	unarmRunSubmitters(seed)
	uassert.Equal(t, 2, len(GetRunSubmitters()))

	// Every listed address is now one the delegate granted, so grant-scoping
	// permits removing all of them. The floor is the only thing left.
	uassert.True(t, RunSubmittersGrantedBy(address(a)))
	uassert.True(t, RunSubmittersGrantedBy(address(b)))

	// Down to one is fine: shrinking the list is the delegate's job.
	RemoveRunSubmitters(cross(cur), []string{a})
	uassert.Equal(t, 1, len(GetRunSubmitters()))

	// The last one is not.
	uassert.AbortsContains(t, cur, "refusing to empty", func() {
		RemoveRunSubmitters(cross(cur), []string{b})
	})
	uassert.True(t, contains(GetRunSubmitters(), b),
		"the last entry must survive a refused removal")

	// Nor in one call, and nor by naming addresses that are not listed: the
	// count is taken against the parameter, not the argument.
	uassert.AbortsContains(t, cur, "refusing to empty", func() {
		RemoveRunSubmitters(cross(cur), []string{b, b})
	})
	uassert.True(t, contains(GetRunSubmitters(), b))

	// And the delegate is not locked out of its ordinary work: adding still
	// works, and once there are two again it may remove one.
	AddRunSubmitters(cross(cur), []string{a})
	RemoveRunSubmitters(cross(cur), []string{b})
	uassert.True(t, contains(GetRunSubmitters(), a))
	uassert.False(t, contains(GetRunSubmitters(), b))
}

// TestRunSubmittersRevocationDropsGrantRecord pins that a new holder does not
// inherit removal authority over its predecessor's grants.
func TestRunSubmittersRevocationDropsGrantRecord(cur realm, t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	runSubmittersMgr = testDelegate
	armRunSubmitters(testutils.TestAddress("seeded-by-vote").String())
	addr := testutils.TestAddress("byfirst").String()
	testing.SetRealm(testing.NewCodeRealm(testDelegate))
	AddRunSubmitters(cross(cur), []string{addr})
	uassert.True(t, RunSubmittersGrantedBy(address(addr)))

	// Hand the capability over, as the proposal executor would.
	runSubmittersMgr = testOther
	runSubmittersGrants = addrset.Set{}

	testing.SetRealm(testing.NewCodeRealm(testOther))
	uassert.False(t, RunSubmittersGrantedBy(address(addr)))
	uassert.AbortsContains(t, cur, "only GovDAO may remove it", func() {
		RemoveRunSubmitters(cross(cur), []string{addr})
	})
}

// TestDelegatePathValidation pins what may be stored in a slot. An ephemeral
// `maketx run` realm has a NON-empty path, so a check for "is this code" would
// let one through; only the gno.land/r/ requirement excludes it.
func TestDelegatePathValidation(t *testing.T) {
	for _, bad := range []string{
		"",
		"gno.land/p/nt/avl/v0",
		"gno.land/e/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/run",
		"gno.land/r/",
		// Paths no caller could ever present. Storing one would leave GovDAO
		// believing it delegated while the delegate is refused on every call.
		"gno.land/r/test/delegate ",                // trailing space
		" gno.land/r/test/delegate",                // leading space
		"gno.land/r/nt/commondao/v0#DAO/42",        // uppercase subpath
		"gno.land/r/nt/commondao/v0#dao/42#dao/43", // two separators
	} {
		// PanicsContains, not AbortsContains: this is a same-realm call, so
		// the panic never crosses a realm boundary and is not an abort.
		uassert.PanicsContains(t, cur, "invalid delegate", func() {
			assertDelegatePath(bad)
		})
	}
	// Sound paths, including a sub-realm identity.
	assertDelegatePath(testDelegate)
	assertDelegatePath(testSubDelegate)
}

// TestValsetGateRejectsForeignRealm closes a pre-existing coverage gap found
// while factoring the shared gate: assertValsetCaller had no test in this realm
// at all, so nothing asserted that a realm other than r/sys/validators/v3 is
// refused.
func TestValsetGateRejectsForeignRealm(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(testOther))
	uassert.AbortsContains(t, cur, "unauthorized", func() {
		SetValsetProposal(cross(cur), []string{"somepubkey:1"})
	})

	// A user account must be refused identically -- the empty-path case.
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("nobody")))
	uassert.AbortsContains(t, cur, "unauthorized", func() {
		SetValsetProposal(cross(cur), []string{"somepubkey:1"})
	})
}

// TestAssertNotValsetKeyRejectsGenericFactory closes the other gap: the guard
// whose own comment says it stops "any GovDAO supermajority" from writing
// validator-set state through the generic param factory had no test.
func TestAssertNotValsetKeyRejectsGenericFactory(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("proposer")))
	uassert.AbortsContains(t, cur, "reserved for", func() {
		NewSysParamStringsPropRequest(cross(cur), "node", "valset", "proposed",
			[]string{"somepubkey:1"})
	})
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// TestRunSubmittersGrantLaunderingIsRefused reproduces a real hole in the first
// version of this code, found by audit.
//
// UpdateSysParamStrings dedupes on add, so re-adding an address already on the
// list is a no-op on the parameter. The first implementation recorded a grant
// for every argument regardless, which let the delegate launder authority over
// entries it never granted:
//
//	read the list -> re-add all of it -> remove all of it
//
// The parameter never changed on the middle step, but every address became
// "granted by me", so the removal was permitted and the allowlist ended empty --
// including the entry that predated the delegation. Since GovDAO proposal
// creation is MsgRun-only, an empty run_submitters means no proposal can be
// created to revoke the delegate or restore the list. That is a chain brick
// recoverable only by relaunch.
func TestRunSubmittersGrantLaunderingIsRefused(cur realm, t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	breakglass := testutils.TestAddress("breakglass2").String()
	testing.SetSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey,
		[]string{breakglass})

	runSubmittersMgr = testDelegate
	armRunSubmitters(testutils.TestAddress("seeded-by-vote").String())
	testing.SetRealm(testing.NewCodeRealm(testDelegate))

	// Step 1: re-add what is already there. A no-op on the parameter, and it
	// must NOT create a grant.
	AddRunSubmitters(cross(cur), GetRunSubmitters())
	uassert.False(t, RunSubmittersGrantedBy(address(breakglass)),
		"re-adding an existing address must not record a grant for it")

	// Step 2: the removal must therefore still be refused.
	uassert.AbortsContains(t, cur, "only GovDAO may remove it", func() {
		RemoveRunSubmitters(cross(cur), []string{breakglass})
	})
	uassert.True(t, contains(GetRunSubmitters(), breakglass),
		"the pre-existing entry must survive")

	// And the whole-list version of the same attack, which is how it would
	// actually be run.
	AddRunSubmitters(cross(cur), GetRunSubmitters())
	uassert.AbortsContains(t, cur, "only GovDAO may remove it", func() {
		RemoveRunSubmitters(cross(cur), GetRunSubmitters())
	})
	uassert.True(t, len(GetRunSubmitters()) > 0,
		"the allowlist must never end up empty at a delegate's hand")

	// A genuinely new address still works normally.
	fresh := testutils.TestAddress("freshgrant").String()
	AddRunSubmitters(cross(cur), []string{fresh})
	uassert.True(t, RunSubmittersGrantedBy(address(fresh)))
	RemoveRunSubmitters(cross(cur), []string{fresh})
	uassert.False(t, contains(GetRunSubmitters(), fresh))
}

// TestRenderShowsDelegationState pins that the page actually reports the state
// it exists to report.
//
// The value of a Render here is that someone auditing the chain can see whether
// a parameter is delegated without knowing to ask. A page that renders the same
// text whether or not a delegation exists would defeat that, so both states are
// checked.
func TestRenderShowsDelegationState(t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	addr := testutils.TestAddress("rendered").String()
	testing.SetSysParamStrings(vmModulePrefix, vmParamsSubmodule, runSubmittersKey,
		[]string{addr})

	// Undelegated: it must say so rather than leave the reader guessing.
	out := Render("")
	uassert.True(t, contains2(out, "nobody"),
		"an undelegated parameter must be reported as such")
	uassert.False(t, contains2(out, testDelegate),
		"no delegate should be named when none is set")

	// Delegated: the holder is named.
	runSubmittersMgr = testDelegate
	armRunSubmitters(testutils.TestAddress("seeded-by-vote").String())
	out = Render("")
	uassert.True(t, contains2(out, testDelegate),
		"the delegate holding the capability must be named")
	uassert.True(t, contains2(out, addr),
		"the addresses the parameter currently allows must be listed")

	// The valset writer is fixed in source, not delegated, and the page should
	// not blur the two.
	uassert.True(t, contains2(out, valsetAuthorizedRealm),
		"the realm that writes valset params must be shown too")
}

// contains2 reports whether s contains sub.
func contains2(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TestRunSubmittersDelegateCannotArmTheGate covers the direction the non-empty
// floor does not: turning the allowlist ON.
//
// An empty run_submitters means the gate is off and anyone may MsgRun. So a
// delegate adding the first address is not curating a list, it is switching a
// chain-wide restriction on and choosing who it admits -- leaving that address
// the only one on the chain that may run code.
//
// That is unrepairable in band. Creating a GovDAO proposal needs MsgRun, so
// once the gate is armed against the members they cannot propose the vote that
// would reset the list or revoke the delegation, and the floor in
// RemoveRunSubmitters stops the delegate from undoing it either.
func TestRunSubmittersDelegateCannotArmTheGate(cur realm, t *testing.T) {
	resetDelegation()
	defer resetDelegation()

	runSubmittersMgr = testDelegate
	attacker := testutils.TestAddress("attacker").String()

	// The gate starts off.
	uassert.Equal(t, 0, len(GetRunSubmitters()))

	testing.SetRealm(testing.NewCodeRealm(testDelegate))
	uassert.AbortsContains(t, cur, "refusing to arm", func() {
		AddRunSubmitters(cross(cur), []string{attacker})
	})
	uassert.Equal(t, 0, len(GetRunSubmitters()),
		"a refused arm must leave the gate off")

	// Once GovDAO has armed it, the delegate may curate as before.
	seeded := testutils.TestAddress("seeded-by-vote").String()
	armRunSubmitters(seeded)
	testing.SetRealm(testing.NewCodeRealm(testDelegate))
	AddRunSubmitters(cross(cur), []string{attacker})
	uassert.Equal(t, 2, len(GetRunSubmitters()),
		"curating an armed list is still the delegate's job")
}

// TestRunSubmittersKeyIsReservedFromGenericFactories pins that the whole-list
// path is the dedicated constructor and nothing else.
//
// The generic factories take module, submodule and name as arguments, so
// without this any of the nine could write run_submitters and walk past the
// proposer rule in ProposeSetRunSubmitters. Same shape as the valset
// reservation next to it.
func TestRunSubmittersKeyIsReservedFromGenericFactories(cur realm, t *testing.T) {
	addr := testutils.TestAddress("someone").String()

	uassert.AbortsContains(t, cur, "reserved for ProposeSetRunSubmitters", func() {
		NewSysParamStringsPropRequest(cross(cur), "vm", "p", runSubmittersKey, []string{addr})
	})
	uassert.AbortsContains(t, cur, "reserved for ProposeSetRunSubmitters", func() {
		NewSysParamStringsPropRequestWithTitle(cross(cur), "vm", "p", runSubmittersKey, "t", []string{addr})
	})
	uassert.AbortsContains(t, cur, "reserved for ProposeSetRunSubmitters", func() {
		NewSysParamStringsPropRequestAddWithTitle(cross(cur), "vm", "p", runSubmittersKey, "t", []string{addr})
	})
	uassert.AbortsContains(t, cur, "reserved for ProposeSetRunSubmitters", func() {
		NewSysParamStringsPropRequestRemoveWithTitle(cross(cur), "vm", "p", runSubmittersKey, "t", []string{addr})
	})
	uassert.AbortsContains(t, cur, "reserved for ProposeSetRunSubmitters", func() {
		NewSysParamStringPropRequest(cross(cur), "vm", "p", runSubmittersKey, addr)
	})

	// A different key in the same module is unaffected.
	uassert.NotAborts(t, cur, func() {
		NewSysParamStringsPropRequest(cross(cur), "vm", "p", "code_submitters", []string{addr})
	})
}

// TestProposeSetRunSubmittersRequiresProposerOnTheList covers the rule that a
// non-empty allowlist must include whoever proposed it.
//
// A list naming nobody who can create a proposal cannot be undone, because
// creating one needs MsgRun. Requiring the proposer's own address proves the
// list is usable: GovDAO refuses a proposal from a non-member, and the proposer
// just signed the transaction, so that address demonstrably holds a key.
func TestProposeSetRunSubmittersRequiresProposerOnTheList(cur realm, t *testing.T) {
	other := testutils.TestAddress("someone-else").String()

	// A list without the proposer is refused.
	uassert.AbortsContains(t, cur, "omits the proposer", func() {
		ProposeSetRunSubmitters(cross(cur), []string{other})
	})

	// Emptying the list is always allowed: that switches the gate off, which
	// cannot lock anyone out.
	uassert.NotAborts(t, cur, func() {
		ProposeSetRunSubmitters(cross(cur), []string{})
	})

	// A list that does include the proposer goes through.
	proposer := testutils.TestAddress("proposer")
	testing.SetRealm(testing.NewUserRealm(proposer))
	uassert.NotAborts(t, cur, func() {
		ProposeSetRunSubmitters(cross(cur), []string{other, proposer.String()})
	})
}
