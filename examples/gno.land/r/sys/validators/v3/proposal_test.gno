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
	// Each operator may appear at most once per proposal; any shape
	// that mentions the same op twice must panic at create-time.
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
	// KeepRunning=false is binding: no proposal shape can keep an
	// opted-out operator in the active set. The dedupe rejection
	// fires before the KR check is even reached.
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
	// RotateValoperSigningKey publishes valset:proposed before any
	// subsequent executor reads, so baseline always reflects the
	// post-rotation state. A later power-update upserts at NEW only;
	// no OLD ghost.
	resetValset(t)
	resetCache()

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

	testing.SetRealm(testing.NewCodeRealm(valopersRealmPath))
	RotateValoperSigningKey(cross, opA, pubKeyA, pubKeyC)
	NotifyValoperChanged(cross, opA, pubKeyC, mustAddr(t, pubKeyC), true)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))
	urequire.NoError(t, newValoperChangeExecutor(
		[]ValoperChange{{OperatorAddress: opA, Power: 2}},
	).Execute(cross))

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

func TestNewValidatorProposalRequest_PhantomBaselineDocumentsUnreachableState(t *testing.T) {
	// Pin executor behavior on a phantom state (cache=NEW,
	// valset:current=OLD, dirty=false) — unreachable via natural
	// flow because RotateValoperSigningKey publishes proposed
	// before NotifyValoperChanged updates the cache. If a future
	// code path ever updates the cache without going through
	// Rotate, the asymmetry would be a real bug; this test pins
	// the current behavior so the divergence surfaces.
	resetValset(t)
	resetCache()

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
		{op: opA, pubKey: pubKeyC, keepRunning: true},
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
	uassert.Equal(t, uint64(1), powerOf[pubKeyA], "phantom OLD lingers from baseline")
	uassert.Equal(t, uint64(2), powerOf[pubKeyC], "executor upserts at NEW")
	uassert.Equal(t, uint64(5), powerOf[pubKeyB], "opB unchanged")
	urequire.Equal(t, 3, len(effective),
		"three entries — phantom-state ghost; this state is unreachable via natural flow")
}

func TestNewValidatorProposalRequest_SameBlockExecuteThenRotate(t *testing.T) {
	// Same-block ordering: proposal-execute writes proposed
	// (dirty=true); a subsequent rotation reads proposed-when-dirty
	// and accumulates the prior power change rather than clobbering
	// back to current.
	resetValset(t)
	resetCache()

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

	urequire.NoError(t, newValoperChangeExecutor(
		[]ValoperChange{{OperatorAddress: opA, Power: 3}},
	).Execute(cross))

	testing.SetRealm(testing.NewCodeRealm(valopersRealmPath))
	RotateValoperSigningKey(cross, opA, pubKeyA, pubKeyC)
	NotifyValoperChanged(cross, opA, pubKeyC, mustAddr(t, pubKeyC), true)

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

// TestNewValidatorProposalRequest_ExecutorVanishedCacheEntryPanics pins
// the "operator vanished from valoperCache between propose and execute"
// branch. No production path deletes from valoperCache today (only Set
// is called by NotifyValoperChanged), but bptree.BPTree exposes Remove
// so the underlying data structure does support deletion. This test
// simulates that hypothetical state directly via the package-private
// cache var to confirm the executor panics with the documented message
// rather than silently mis-publishing an empty/wrong valset.
//
// If a future commit ever introduces a public cache-delete path, this
// test still passes — it's a contract-pinning test for the panic itself.
// If the team decides the branch is unreachable enough to drop, this
// test is the first thing to break.
func TestNewValidatorProposalRequest_ExecutorVanishedCacheEntryPanics(t *testing.T) {
	resetValset(t)
	resetCache()

	op := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyA, keepRunning: true}})

	exec := newValoperChangeExecutor(
		[]ValoperChange{{OperatorAddress: op, Power: 5}},
	)

	// Simulate the unreachable-today state: cache entry deleted between
	// proposal-create and proposal-execute. Direct package-private mutation
	// (NOT a public API) — production code has no path here today.
	_, removed := valoperCache.Remove(op.String())
	urequire.True(t, removed, "fixture: cache entry must have been present before Remove")

	uassert.AbortsContains(t, "operator vanished from valoperCache between propose and execute", func() {
		_ = exec.Execute(cross)
	})
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
