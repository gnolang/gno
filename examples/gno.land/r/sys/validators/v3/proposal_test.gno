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
