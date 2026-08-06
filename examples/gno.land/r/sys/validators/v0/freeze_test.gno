// freeze_test.gno — Phase A freeze/unfreeze.
//
// These drive the real entry points against a seeded valset
// (testing.SetSysParamStrings) and a seeded valoperCache, rather than
// asserting on hand-mutated trees: the interesting behaviour is what
// gets PUBLISHED to sysparams, and which key it is published under.
//
// State isolation: gno's test runner shares realm globals across
// top-level TestXxx, so every case starts from resetFreezeState.
package validators

import (
	"strings"
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	sysparams "gno.land/r/sys/params"
)

// resetFreezeState clears everything freeze.gno owns. Called by
// resetValset too, so realm state cannot leak into Render tests.
func resetFreezeState() {
	frozenSet = avl.NewTree()
	freezeAdmins = avl.NewTree()
}

// seedThreeValidators seeds valset:current with A/B/C at power 10 and
// gives each a valoperCache profile. Returns the three operators.
func seedThreeValidators(t *testing.T) (opA, opB, opC address) {
	t.Helper()
	resetValset(t)
	resetCache()

	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":10", pubKeyC + ":10",
	})

	opA = testutils.TestAddress("op-A")
	opB = testutils.TestAddress("op-B")
	opC = testutils.TestAddress("op-C")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
		{op: opC, pubKey: pubKeyC, keepRunning: true},
	})
	return opA, opB, opC
}

// powerByPubKey indexes the effective valset by signing pubkey.
func powerByPubKey() map[string]uint64 {
	out := map[string]uint64{}
	for _, v := range sysparams.GetValsetEffective() {
		out[v.PubKey] = v.VotingPower
	}
	return out
}

// ---- freeze ------------------------------------------------------------------

func TestFreeze_RemovesOperatorFromEffectiveValset(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)

	doFreeze(cur, opA, "double sign")

	urequire.True(t, IsFrozen(opA), "opA recorded as frozen")
	powers := powerByPubKey()
	urequire.Equal(t, 2, len(powers), "two validators left live")
	_, stillThere := powers[pubKeyA]
	uassert.False(t, stillThere, "opA's signing key is out of the valset")
	uassert.Equal(t, uint64(10), powers[pubKeyB])
	uassert.Equal(t, uint64(10), powers[pubKeyC])
}

func TestFreeze_RecordsPreFreezePower(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":7", pubKeyB + ":100", pubKeyC + ":100",
	})

	doFreeze(cur, opA, "downtime")

	e := mustFrozen(opA)
	uassert.Equal(t, uint64(7), e.VotingPower, "pre-freeze power captured")
	uassert.Equal(t, mustAddr(t, pubKeyA), e.SigningAddress)
	uassert.Equal(t, "downtime", e.Reason)
}

func TestFreeze_RejectsUnknownOperator(cur realm, t *testing.T) {
	seedThreeValidators(t)
	ghost := testutils.TestAddress("ghost-op")

	uassert.PanicsContains(t, cur, "unknown operator", func() {
		doFreeze(cur, ghost, "nope")
	})
}

func TestFreeze_RejectsOperatorNotInValset(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	// Same roster, but opA's signing key is no longer published.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyB + ":10", pubKeyC + ":10",
	})

	uassert.PanicsContains(t, cur, "not in the effective valset", func() {
		doFreeze(cur, opA, "nope")
	})
}

func TestFreeze_RejectsDoubleFreeze(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)

	doFreeze(cur, opA, "first")
	uassert.PanicsContains(t, cur, "already frozen", func() {
		doFreeze(cur, opA, "second")
	})
}

// ---- liveness floor -----------------------------------------------------------

// The floor is measured against the FULL roster, not against the
// already-shrunken effective set: with three validators, one freeze
// leaves 2/3 (allowed) and a second would leave 1/3 (refused).
func TestFreeze_LivenessFloorCountsHeldOutValidators(cur realm, t *testing.T) {
	opA, opB, _ := seedThreeValidators(t)

	doFreeze(cur, opA, "first")
	urequire.Equal(t, 2, len(sysparams.GetValsetEffective()))

	uassert.PanicsContains(t, cur, "below the 2/3 floor", func() {
		doFreeze(cur, opB, "second")
	})

	// The refused freeze left no trace.
	uassert.False(t, IsFrozen(opB), "opB not recorded")
	uassert.Equal(t, 2, len(sysparams.GetValsetEffective()), "valset unchanged")
}

// A count-only floor would wave this through: three validators, freeze
// one, 2/3 still live. But that one holds 40% of the power, so the
// remaining 60% cannot reach the 2/3 Tendermint needs.
func TestFreeze_PowerFloorCatchesWhatTheCountFloorMisses(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":40", pubKeyB + ":30", pubKeyC + ":30",
	})

	uassert.PanicsContains(t, cur, "would leave 60 of 100 voting power live", func() {
		doFreeze(cur, opA, "heavyweight")
	})
	uassert.False(t, IsFrozen(opA))
	uassert.Equal(t, 3, len(sysparams.GetValsetEffective()), "valset unchanged")
}

func TestFreeze_PowerFloorAllowsALightValidator(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":45", pubKeyC + ":45",
	})

	doFreeze(cur, opA, "lightweight")

	urequire.True(t, IsFrozen(opA))
	uassert.Equal(t, 2, len(sysparams.GetValsetEffective()))
}

func TestFreeze_LivenessFloorMessageReportsRosterNotEffective(cur realm, t *testing.T) {
	opA, opB, _ := seedThreeValidators(t)
	doFreeze(cur, opA, "first")

	// "1 of 3", not "1 of 2": the frozen validator still counts toward
	// the roster the floor protects.
	uassert.PanicsContains(t, cur, "would leave 1 of 3 validators live", func() {
		doFreeze(cur, opB, "second")
	})
}

// ---- unfreeze -----------------------------------------------------------------

func TestUnfreeze_RestoresAtRecordedPower(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":7", pubKeyB + ":100", pubKeyC + ":100",
	})

	doFreeze(cur, opA, "downtime")
	doUnfreeze(cur, opA, "recovered")

	uassert.False(t, IsFrozen(opA), "frozen entry cleared")
	powers := powerByPubKey()
	urequire.Equal(t, 3, len(powers), "back to three validators")
	uassert.Equal(t, uint64(7), powers[pubKeyA], "restored at pre-freeze power")
}

// The regression this whole operator-keyed design exists for: an
// operator who rotates while frozen must come back under the NEW key.
// Restoring the captured signing address would republish a retired
// pubkey to consensus.
func TestUnfreeze_RestoresUnderRotatedKeyNotTheCapturedOne(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)

	doFreeze(cur, opA, "suspected key compromise")
	urequire.Equal(t, mustAddr(t, pubKeyA), mustFrozen(opA).SigningAddress)

	// opA rotates A -> C while frozen. RotateValoperSigningKey is a
	// no-op on the valset (opA is not in it), but the cache moves on.
	// pubKeyC is free here because seedThreeValidators' C is a distinct
	// operator; drop it from the valset first to keep the test honest.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyB + ":10",
	})
	valoperCache.Set(opA.String(), cacheEntry{
		SigningPubKey:  pubKeyC,
		SigningAddress: mustAddr(t, pubKeyC),
		KeepRunning:    true,
	})

	doUnfreeze(cur, opA, "rotated and cleared")

	powers := powerByPubKey()
	_, retired := powers[pubKeyA]
	uassert.False(t, retired, "the RETIRED key must not be republished")
	uassert.Equal(t, uint64(10), powers[pubKeyC], "restored under the rotated key")
}

func TestUnfreeze_RefusesWhenOperatorOptedOut(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	doFreeze(cur, opA, "downtime")

	// Operator sets KeepRunning=false while frozen.
	valoperCache.Set(opA.String(), cacheEntry{
		SigningPubKey:  pubKeyA,
		SigningAddress: mustAddr(t, pubKeyA),
		KeepRunning:    false,
	})

	uassert.PanicsContains(t, cur, "KeepRunning=false; refusing to restore", func() {
		doUnfreeze(cur, opA, "come back")
	})
	uassert.True(t, IsFrozen(opA), "entry survives the refused unfreeze")
}

func TestUnfreeze_RejectsOperatorThatIsNotFrozen(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)

	uassert.PanicsContains(t, cur, "operator is not frozen", func() {
		doUnfreeze(cur, opA, "nothing to do")
	})
}

// ---- drop ----------------------------------------------------------------------

func TestDropFrozen_ForgetsEntryWithoutRestoring(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	doFreeze(cur, opA, "permanent removal pending")

	doDropFrozen(opA, "governance removed them")

	uassert.False(t, IsFrozen(opA), "entry forgotten")
	powers := powerByPubKey()
	uassert.Equal(t, 2, len(powers), "operator stays out of the valset")
	_, back := powers[pubKeyA]
	uassert.False(t, back, "drop must not restore")
}

func TestDropFrozen_RejectsOperatorThatIsNotFrozen(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)

	uassert.PanicsContains(t, cur, "operator is not frozen", func() {
		doDropFrozen(opA, "nothing to drop")
	})
}

// ---- interaction with the valoper proposal path ---------------------------------

// A frozen operator has one owner. A GovDAO power change would put them
// back in the valset while frozenSet still holds the entry.
func TestValoperProposal_RejectsFrozenOperator(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	doFreeze(cur, opA, "frozen")

	uassert.PanicsContains(t, cur, "is frozen; unfreeze or drop", func() {
		_ = NewValidatorProposalRequest(cur,
			[]ValoperChange{{OperatorAddress: opA, Power: 5}},
			"resurrect opA", "")
	})
}

func TestValoperProposal_RejectsOperatorFrozenAfterProposalCreated(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)

	// Proposal created while opA is healthy...
	executor := newValoperChangeExecutor(cur,
		[]ValoperChange{{OperatorAddress: opA, Power: 5}})

	// ...opA is frozen before the vote lands.
	doFreeze(cur, opA, "frozen mid-flight")

	uassert.AbortsContains(t, cur, "is frozen at execution", func() {
		_ = executor.Execute(cross(cur))
	})
}

// ---- admin auth ------------------------------------------------------------------

func TestFreezeAdmins_EmptyByDefault(cur realm, t *testing.T) {
	resetFreezeState()
	uassert.Equal(t, 0, len(GetFreezeAdmins()))
	uassert.False(t, IsFreezeAdmin(testutils.TestAddress("nobody")))
}

func TestFreezeValidator_RejectsNonAdminUser(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	stranger := testutils.TestAddress("stranger")
	testing.SetRealm(testing.NewUserRealm(stranger))

	uassert.AbortsContains(t, cur, "caller is not a freeze admin", func() {
		FreezeValidator(cross(cur), opA, "unauthorised")
	})
	uassert.False(t, IsFrozen(opA))
}

// tx-origin auth would accept this: the admin signed the transaction.
// Requiring the immediate caller to be a user realm does not.
func TestFreezeValidator_RejectsAdminCallLaunderedThroughARealm(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	admin := testutils.TestAddress("freeze-admin")
	freezeAdmins.Set(admin.String(), struct{}{})

	testing.SetOriginCaller(admin)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/attacker"))

	uassert.AbortsContains(t, cur, "must call directly", func() {
		FreezeValidator(cross(cur), opA, "laundered")
	})
	uassert.False(t, IsFrozen(opA), "no freeze applied")
}

func TestFreezeValidator_AcceptsAdminDirectCall(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	admin := testutils.TestAddress("freeze-admin")
	freezeAdmins.Set(admin.String(), struct{}{})

	testing.SetRealm(testing.NewUserRealm(admin))
	FreezeValidator(cross(cur), opA, "admin freeze")

	urequire.True(t, IsFrozen(opA))
	uassert.Equal(t, 2, len(sysparams.GetValsetEffective()))
}

// ---- render --------------------------------------------------------------------

func TestRenderFrozen_ListsFrozenOperatorsAndAdmins(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	admin := testutils.TestAddress("freeze-admin")

	// Untouched chain: Render says nothing about freezing at all.
	out := Render("")
	uassert.False(t, strings.Contains(out, "## Frozen Validators"), "no section when nothing is frozen")
	uassert.False(t, strings.Contains(out, "## Freeze Admins"), "no section when no admin is configured")

	doFreeze(cur, opA, "double sign")

	// Frozen but no admin: the section appears and says so explicitly.
	out = Render("")
	uassert.True(t, strings.Contains(out, "_(none — only GovDAO can freeze/unfreeze)_"), "empty admin list noted")

	freezeAdmins.Set(admin.String(), struct{}{})

	out = Render("")
	uassert.True(t, strings.Contains(out, "## Frozen Validators"), "frozen section present")
	uassert.True(t, strings.Contains(out, opA.String()), "frozen operator listed")
	uassert.True(t, strings.Contains(out, "double sign"), "reason listed")
	uassert.True(t, strings.Contains(out, admin.String()), "admin listed")
}
