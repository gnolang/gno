package validators

import (
	"strconv"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	sysparams "gno.land/r/sys/params"
)

// seedCache populates valoperCache with the given (op, pubkey, kr)
// tuples — used in tests to satisfy NewValidatorProposalRequest's
// creation-time membership check without going through valopers.
func seedCache(t *testing.T, entries []struct {
	op          address
	pubKey      string
	keepRunning bool
}) {
	t.Helper()
	for _, e := range entries {
		signingAddr := mustAddr(t, e.pubKey)
		valoperCache.Set(e.op.String(), cacheEntry{
			SigningPubKey:  e.pubKey,
			SigningAddress: signingAddr,
			KeepRunning:    e.keepRunning,
		})
	}
}

func TestNewValidatorProposalRequest_RejectsUnknownOperator(t *testing.T) {
	resetCache()

	op := testutils.TestAddress("ghost-op")

	uassert.PanicsContains(t, "unknown operator", func() {
		_ = NewValidatorProposalRequest(
			[]ValoperChange{{OperatorAddress: op, Power: 1}},
			"add ghost",
			"",
		)
	})
}

func TestNewValidatorProposalRequest_RejectsEmptyChanges(t *testing.T) {
	resetCache()

	uassert.PanicsContains(t, errNoValoperChanges, func() {
		_ = NewValidatorProposalRequest(nil, "title", "")
	})
}

func TestNewValidatorProposalRequest_RejectsEmptyTitle(t *testing.T) {
	resetCache()
	op := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyA, keepRunning: true}})

	uassert.PanicsContains(t, "proposal title is empty", func() {
		_ = NewValidatorProposalRequest(
			[]ValoperChange{{OperatorAddress: op, Power: 1}},
			"   ",
			"",
		)
	})
}

func TestNewValidatorProposalRequest_RejectsTooManyChanges(t *testing.T) {
	resetCache()

	// Seed 41 cache entries so the membership check passes; the
	// length cap should fire before the per-entry validation.
	changes := make([]ValoperChange, 41)
	pubkeys := []string{pubKeyA, pubKeyB, pubKeyC}
	for i := 0; i < 41; i++ {
		op := testutils.TestAddress("op-" + strconv.Itoa(i))
		pk := pubkeys[i%3]
		valoperCache.Set(op.String(), cacheEntry{
			SigningPubKey:  pk,
			SigningAddress: mustAddr(t, pk),
			KeepRunning:    true,
		})
		changes[i] = ValoperChange{OperatorAddress: op, Power: 1}
	}

	uassert.PanicsContains(t, "max number of allowed validators per proposal is 40", func() {
		_ = NewValidatorProposalRequest(changes, "too many", "")
	})
}

func TestNewValidatorProposalRequest_DescriptionRendering(t *testing.T) {
	resetCache()
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	pr := NewValidatorProposalRequest(
		[]ValoperChange{
			{OperatorAddress: opA, Power: 5},
			{OperatorAddress: opB, Power: 0},
		},
		"mixed changes",
		"context line",
	)

	desc := pr.Description()
	urequire.True(t, len(desc) > 0)
	uassert.True(t, contains(desc, "context line"))
	uassert.True(t, contains(desc, "## Validator Updates"))
	uassert.True(t, contains(desc, opA.String()+": add (power 5)"))
	uassert.True(t, contains(desc, opB.String()+": remove"))
}

func TestNewValidatorProposalRequest_ExecutorReResolvesPubkey(t *testing.T) {
	// Creation-time captured changes: ValoperChange refers to opA.
	// Cache for opA points to pubKeyA at creation. Before execution,
	// opA's cache entry is updated to pubKeyB. Executor must publish
	// the NEW pubkey, not the creation-time one.
	resetValset(t)
	resetCache()

	opA := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: opA, pubKey: pubKeyA, keepRunning: true}})

	changes := []ValoperChange{{OperatorAddress: opA, Power: 7}}

	// Build the executor; it captures `changes` by reference (slice
	// of structs) but resolves SigningPubKey at run-time via cache.
	exec := newValoperChangeExecutor(changes)

	// Simulate operator rotation: opA's cache entry now points to
	// pubKeyB. Captured changes slice is unchanged.
	valoperCache.Set(opA.String(), cacheEntry{
		SigningPubKey:  pubKeyB,
		SigningAddress: mustAddr(t, pubKeyB),
		KeepRunning:    true,
	})

	urequire.NoError(t, exec.Execute(cross))

	// Effective valset should contain pubKeyB (post-rotation), not
	// pubKeyA (creation-time).
	effective := sysparams.GetValsetEffective()
	urequire.Equal(t, 1, len(effective))
	uassert.Equal(t, pubKeyB, effective[0].PubKey)
	uassert.Equal(t, uint64(7), effective[0].VotingPower)
}

func TestNewValidatorProposalRequest_RemoveOperator(t *testing.T) {
	resetValset(t)
	resetCache()

	// Seed valset with opA already signing under pubKeyA.
	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":10"})

	opA := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: opA, pubKey: pubKeyA, keepRunning: false}})

	// Liveness floor: removing the only validator empties the set.
	// Executor runs inside a crossing dao.Executor.Execute call, so
	// the panic surfaces as an abort, not a regular panic.
	uassert.AbortsContains(t, "would empty the validator set", func() {
		_ = newValoperChangeExecutor([]ValoperChange{{OperatorAddress: opA, Power: 0}}).Execute(cross)
	})
}

func TestNewValidatorProposalRequest_RemoveLeavesOthers(t *testing.T) {
	resetValset(t)
	resetCache()

	// Seed valset with two validators.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10",
		pubKeyB + ":5",
	})

	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	changes := []ValoperChange{{OperatorAddress: opA, Power: 0}}
	// Build the executor directly (private function, same package).
	urequire.NoError(t, newValoperChangeExecutor(changes).Execute(cross))

	// Effective set: only opB / pubKeyB remains.
	effective := sysparams.GetValsetEffective()
	urequire.Equal(t, 1, len(effective))
	uassert.Equal(t, pubKeyB, effective[0].PubKey)
}

func TestNewValidatorProposalRequest_RejectsKeepRunningFalseAtCreation(t *testing.T) {
	resetCache()
	op := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyA, keepRunning: false}})

	uassert.PanicsContains(t, "KeepRunning=false", func() {
		_ = NewValidatorProposalRequest(
			[]ValoperChange{{OperatorAddress: op, Power: 1}},
			"add opted-out", "",
		)
	})
}

func TestNewValidatorProposalRequest_AllowsRemoveOfKeepRunningFalse(t *testing.T) {
	// KeepRunning=false is the operator's opt-out signal; removing
	// such an operator must still be allowed (it's the standard
	// exit path). Only adds are gated.
	resetValset(t)
	resetCache()

	// Seed the valset with two operators so removing one doesn't
	// trip the empty-valset liveness floor.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10",
		pubKeyB + ":5",
	})

	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: false}, // opted out
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	// Build proposal succeeds (remove path: Power=0 ignores KeepRunning).
	pr := NewValidatorProposalRequest(
		[]ValoperChange{{OperatorAddress: opA, Power: 0}},
		"remove opted-out opA", "",
	)
	_ = pr

	// Executor also succeeds.
	urequire.NoError(t, newValoperChangeExecutor([]ValoperChange{{OperatorAddress: opA, Power: 0}}).Execute(cross))
}

func TestNewValidatorProposalRequest_RejectsDuplicateOp(t *testing.T) {
	// Dedupe: each operator may appear at most once per proposal.
	// The legacy [{op,0},{op,N}] power-update pair is now a
	// duplicate and must panic at create-time. So is any other
	// shape that mentions the same op twice ([{op,N},{op,M}],
	// [{op,0},{op,0}]).
	resetCache()
	op := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyA, keepRunning: true}})

	cases := [][]ValoperChange{
		{{OperatorAddress: op, Power: 0}, {OperatorAddress: op, Power: 7}}, // remove + re-add
		{{OperatorAddress: op, Power: 7}, {OperatorAddress: op, Power: 8}}, // double add
		{{OperatorAddress: op, Power: 0}, {OperatorAddress: op, Power: 0}}, // double remove
	}
	for _, changes := range cases {
		uassert.PanicsContains(t, "duplicate operator in proposal", func() {
			_ = NewValidatorProposalRequest(changes, "dup", "")
		})
	}
}

func TestNewValidatorProposalRequest_RejectsPowerUpdatePairForOptedOutOp(t *testing.T) {
	// KeepRunning=false is a BINDING opt-out: no proposal shape can
	// keep an opted-out operator in the active set. The legacy
	// [{op,0},{op,N}] pair previously bypassed the KR check via the
	// position-aware exemption; that bypass is now closed at the
	// dedupe layer.
	resetCache()
	op := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyA, keepRunning: false}})

	uassert.PanicsContains(t, "duplicate operator in proposal", func() {
		_ = NewValidatorProposalRequest(
			[]ValoperChange{
				{OperatorAddress: op, Power: 0},
				{OperatorAddress: op, Power: 7},
			},
			"bypass attempt", "",
		)
	})
}

func TestNewValidatorProposalRequest_UpsertExistingValidator(t *testing.T) {
	// Single-entry {op, newPower} against an op already in the
	// effective valset must upsert: the existing entry's power is
	// overwritten, no remove/re-add ceremony required.
	resetValset(t)
	resetCache()

	// Seed valset with two validators; we upsert opA's power.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":1",
		pubKeyB + ":1",
	})

	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	changes := []ValoperChange{{OperatorAddress: opA, Power: 9}}
	urequire.NoError(t, newValoperChangeExecutor(changes).Execute(cross))

	effective := sysparams.GetValsetEffective()
	powerOf := map[string]uint64{}
	for _, v := range effective {
		powerOf[v.PubKey] = v.VotingPower
	}
	uassert.Equal(t, uint64(9), powerOf[pubKeyA], "opA power upserted from 1 to 9")
	uassert.Equal(t, uint64(1), powerOf[pubKeyB], "opB unchanged")
}

func TestNewValidatorProposalRequest_ExecutorRejectsRaceFlippedKeepRunning(t *testing.T) {
	// KeepRunning=true at proposal-create time; operator flips to
	// false BEFORE the executor runs. Race-safety check rejects.
	resetValset(t)
	resetCache()

	op := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyA, keepRunning: true}})

	changes := []ValoperChange{{OperatorAddress: op, Power: 5}}
	exec := newValoperChangeExecutor(changes)

	// Operator flips KeepRunning=false BEFORE the executor runs.
	valoperCache.Set(op.String(), cacheEntry{
		SigningPubKey:  pubKeyA,
		SigningAddress: mustAddr(t, pubKeyA),
		KeepRunning:    false,
	})

	uassert.AbortsContains(t, "KeepRunning=false at execution", func() {
		_ = exec.Execute(cross)
	})
}

func TestNewValidatorProposalRequest_NaturalRotationFlow_NoGhost(t *testing.T) {
	// Regression: under the natural rotation-then-power-update flow,
	// the executor must not leave a stale OLD signing addr in the
	// published valset. The flow is:
	//
	//   1. opA in valset under OLD (pubKeyA) at power=1.
	//   2. RotateValoperSigningKey(opA, OLD, NEW) — this is what
	//      valopers.UpdateSigningKey calls. The rotate function
	//      reads GetValsetEffective, deletes set[OLD], inserts
	//      set[NEW]@1, and publishes via SetValsetProposal. Sets
	//      dirty=true. valset:proposed now has NEW (no OLD).
	//   3. NotifyValoperChanged updates cache to NEW.
	//   4. Later, a proposal {opA, power=2} executes. Its baseline
	//      = GetValsetEffective() returns proposed (since
	//      dirty=true), reflecting the post-rotation NEW signing
	//      addr — NOT OLD. Executor upserts set[NEW]=NEW@2. No
	//      OLD ghost.
	//
	// The pre-existing "rotation-crossover" concern was speculation
	// that baseline could carry OLD while cache carries NEW; under
	// the natural flow that state never arises because rotation
	// publishes BEFORE any subsequent executor reads.
	resetValset(t)
	resetCache()

	// Step 1: seed valset:current with opA at OLD@1 and a peer.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":1",
		pubKeyB + ":5",
	})
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	// Step 2: rotate opA OLD -> NEW (pubKeyA -> pubKeyC). This is
	// the gate-protected v3 entrypoint; auth requires PreviousRealm
	// == valopers.
	testing.SetRealm(testing.NewCodeRealm(valopersRealmPath))
	RotateValoperSigningKey(cross, opA, pubKeyA, pubKeyC)

	// Step 3: simulate NotifyValoperChanged updating the cache —
	// in the real flow this is the next call from valopers.
	NotifyValoperChanged(cross, opA, pubKeyC, mustAddr(t, pubKeyC), true)

	// Step 4: proposal {opA, power=2} now executes. Reset realm
	// to a non-restricted context so SetValsetProposal can run via
	// the executor's internal cross-call.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))
	urequire.NoError(t, newValoperChangeExecutor(
		[]ValoperChange{{OperatorAddress: opA, Power: 2}},
	).Execute(cross))

	// Verify: final published valset has opA at NEW@2 and opB
	// unchanged. The OLD signing addr (pubKeyA) MUST NOT appear.
	effective := sysparams.GetValsetEffective()
	powerOf := map[string]uint64{}
	for _, v := range effective {
		powerOf[v.PubKey] = v.VotingPower
	}
	uassert.Equal(t, uint64(2), powerOf[pubKeyC], "opA published at NEW power=2")
	uassert.Equal(t, uint64(5), powerOf[pubKeyB], "opB unchanged")
	_, ghost := powerOf[pubKeyA]
	uassert.False(t, ghost, "OLD signing key (pubKeyA) must not linger in valset")
	urequire.Equal(t, 2, len(effective), "exactly two entries — opA(NEW), opB")
}

func TestNewValidatorProposalRequest_PhantomBaselineWouldGhost(t *testing.T) {
	// Adversarial / documentation-of-risk: directly inject a
	// phantom state where the cache has been updated to NEW but
	// the valset still holds OLD (and dirty=false, so baseline
	// reads current). Confirm what the executor does — and
	// document that the natural flow cannot produce this state
	// (RotateValoperSigningKey publishes proposed before
	// NotifyValoperChanged updates the cache).
	//
	// If this test ever ASSERTS NO GHOST while the natural-flow
	// invariant breaks (e.g., if a future code path updates cache
	// without going through Rotate), the asymmetry would be a
	// real bug.
	resetValset(t)
	resetCache()

	// Phantom: valset:current still has OLD; proposed empty, dirty=false.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":1",
		pubKeyB + ":5",
	})

	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	// Cache: opA points to NEW (pubKeyC) without rotate ever publishing.
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyC, keepRunning: true}, // phantom NEW
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	urequire.NoError(t, newValoperChangeExecutor(
		[]ValoperChange{{OperatorAddress: opA, Power: 2}},
	).Execute(cross))

	effective := sysparams.GetValsetEffective()
	powerOf := map[string]uint64{}
	for _, v := range effective {
		powerOf[v.PubKey] = v.VotingPower
	}

	// Under the phantom state the executor DOES leave a ghost:
	// it upserts at addr_NEW (cache) while addr_OLD lingers from
	// baseline. Document this with assertions so a future
	// behavioral change is caught.
	uassert.Equal(t, uint64(1), powerOf[pubKeyA], "phantom OLD lingers from baseline")
	uassert.Equal(t, uint64(2), powerOf[pubKeyC], "executor upserts at NEW")
	uassert.Equal(t, uint64(5), powerOf[pubKeyB], "opB unchanged")
	urequire.Equal(t, 3, len(effective),
		"three entries — phantom-state ghost; this state is unreachable via natural flow")
}

func TestNewValidatorProposalRequest_SameBlockExecuteThenRotate(t *testing.T) {
	// Pin GetValsetEffective semantics under same-block ordering:
	// proposal-execute first writes valset:proposed (dirty=true);
	// the next rotation (also same-block) reads proposed-when-dirty
	// as its baseline and accumulates instead of clobbering. The
	// final published valset must reflect BOTH changes — power
	// upserted on the operator AND the operator's signing key
	// rotated to NEW.
	resetValset(t)
	resetCache()

	// Seed valset with two operators: opA at OLD (pubKeyA) power=1,
	// opB at pubKeyB power=5.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":1",
		pubKeyB + ":5",
	})

	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
	})

	// Phase 1 (same block): proposal-execute upserts opA's power
	// from 1 to 3. baseline=current, set[addrA]=power=3, publish
	// proposed=[pubKeyA:3, pubKeyB:5], dirty=true.
	urequire.NoError(t, newValoperChangeExecutor(
		[]ValoperChange{{OperatorAddress: opA, Power: 3}},
	).Execute(cross))

	// Phase 2 (same block): rotate opA OLD -> NEW (pubKeyA -> pubKeyC).
	// RotateValoperSigningKey reads GetValsetEffective which returns
	// proposed-when-dirty = [pubKeyA:3, pubKeyB:5]. It deletes
	// addrA, inserts addrC carrying the SAME power (3) preserved
	// from the previous proposal — accumulation, not clobber.
	testing.SetRealm(testing.NewCodeRealm(valopersRealmPath))
	RotateValoperSigningKey(cross, opA, pubKeyA, pubKeyC)
	NotifyValoperChanged(cross, opA, pubKeyC, mustAddr(t, pubKeyC), true)

	// Final published valset must have pubKeyC at power=3 (both the
	// power update AND the rotation applied) and pubKeyB unchanged.
	// Critically, pubKeyA must NOT appear (no clobber back to old
	// state) and pubKeyC must NOT be at the old power=1 (no missed
	// upsert).
	effective := sysparams.GetValsetEffective()
	powerOf := map[string]uint64{}
	for _, v := range effective {
		powerOf[v.PubKey] = v.VotingPower
	}
	uassert.Equal(t, uint64(3), powerOf[pubKeyC], "rotation accumulates with prior upsert; final power=3")
	uassert.Equal(t, uint64(5), powerOf[pubKeyB], "opB unchanged")
	_, gotOld := powerOf[pubKeyA]
	uassert.False(t, gotOld, "OLD signing key (pubKeyA) must not appear after rotation")
	urequire.Equal(t, 2, len(effective), "exactly two entries — opA(NEW) and opB")
}

// contains is a tiny strings.Contains shim so tests don't import a
// new package.
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
