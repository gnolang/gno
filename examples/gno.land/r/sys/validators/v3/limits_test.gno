package validators

import (
	"testing"
	"time"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// resetLimits restores the cooldown + trust-level state to its package
// default. Required because gno's test runner does not isolate
// package-level vars across top-level tests — once any test runs a
// successful proposal executor, lastValsetUpdate is set to time.Now()
// and subsequent tests would trip the cooldown check unless reset.
//
// Pair with resetCache() / resetValset(t) at the top of any test that
// drives a valset proposal through to execution.
func resetLimits() {
	lastValsetUpdate = time.Time{}
	valsetUpdateCooldown = 24 * time.Hour
	trustLevelRatio = trustRatio{numerator: 1, denominator: 3}
}

func TestNewValidatorProposalRequest_RejectsWithinCooldown(cur realm, t *testing.T) {
	// Creation-time guard: a proposal cannot even enter GovDAO if the
	// previous valset update was within the cooldown window.
	resetValset(t)
	resetCache()
	resetLimits()
	lastValsetUpdate = time.Now()

	op := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: op, pubKey: pubKeyA, keepRunning: true}})

	uassert.PanicsContains(t, cur, errValsetUpdateCooldown, func() {
		_ = NewValidatorProposalRequest(cur,
			[]ValoperChange{{OperatorAddress: op, Power: 1}},
			"add during cooldown", "",
		)
	})
}

func TestNewValidatorProposalRequest_ExecutorRejectsWithinCooldown(cur realm, t *testing.T) {
	// Execution-time guard (the binding source of truth): a proposal
	// that was created when cooldown was clear can still be rejected
	// at execute-time if another proposal slipped in between.
	resetValset(t)
	resetCache()
	resetLimits()

	testing.SetSysParamStrings(module, submodule, currKey, []string{pubKeyA + ":1"})
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

	exec := newValoperChangeExecutor(cur, []ValoperChange{{OperatorAddress: opB, Power: 1}}, trustLevelRatio)

	// Simulate another proposal having just landed between creation
	// and execution of this one.
	lastValsetUpdate = time.Now()

	uassert.AbortsContains(t, cur, errValsetUpdateCooldown, func() {
		_ = exec.Execute(cross(cur))
	})
}

func TestNewValidatorProposalRequest_AllowsAfterCooldown(cur realm, t *testing.T) {
	// Happy path: cooldown elapsed → proposal applies → lastValsetUpdate
	// is bumped forward by the executor.
	resetValset(t)
	resetCache()
	resetLimits()

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

	lastValsetUpdate = time.Now().Add(-valsetUpdateCooldown - time.Minute)

	exec := newValoperChangeExecutor(cur, []ValoperChange{{OperatorAddress: opA, Power: 11}}, trustLevelRatio)
	urequire.NoError(t, exec.Execute(cross(cur)))

	// Executor should have refreshed the cooldown clock.
	uassert.True(t, time.Since(lastValsetUpdate) < time.Minute,
		"executor must bump lastValsetUpdate on success")
}

func TestNewValidatorProposalRequest_RejectsTrustLevelViolation(cur realm, t *testing.T) {
	// Baseline: A:10, B:1, C:1 (prevTotal=12, trust threshold = 4 with
	// 1/3 ratio). Proposal removes A. Retained = 2 (B+C). 2 < 4 →
	// trust-level violation. Uses only A/B/C so we don't need extra
	// pubkeys.
	resetValset(t)
	resetCache()
	resetLimits()

	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10",
		pubKeyB + ":1",
		pubKeyC + ":1",
	})
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	opC := testutils.TestAddress("op-C")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
		{op: opC, pubKey: pubKeyC, keepRunning: true},
	})

	exec := newValoperChangeExecutor(cur, []ValoperChange{{OperatorAddress: opA, Power: 0}}, trustLevelRatio)

	uassert.AbortsContains(t, cur, errTrustLevelViolated, func() {
		_ = exec.Execute(cross(cur))
	})
}

func TestNewValidatorProposalRequest_RejectsAtTrustLevelBoundary(cur realm, t *testing.T) {
	// Baseline: A:4, B:3, C:2 (prevTotal=9, threshold = 9/3 = 3).
	// Proposal removes A and C, retaining only B. retained (BASELINE
	// VP of survivors) = 3, exactly equal to threshold. 3*3 = 9, 9*1
	// = 9, 9 <= 9 ⇒ true ⇒ aborts.
	//
	// This pins the boundary as STRICT — talliedVotingPower must be
	// strictly greater than the threshold to pass. Matches the
	// CometBFT light-client rule (verifyCommitSingle:
	// "if got <= needed return ErrNotEnoughVotingPowerSigned").
	resetValset(t)
	resetCache()
	resetLimits()

	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":4",
		pubKeyB + ":3",
		pubKeyC + ":2",
	})
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	opC := testutils.TestAddress("op-C")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
		{op: opC, pubKey: pubKeyC, keepRunning: true},
	})

	exec := newValoperChangeExecutor(cur, []ValoperChange{
		{OperatorAddress: opA, Power: 0},
		{OperatorAddress: opC, Power: 0},
	}, trustLevelRatio)

	uassert.AbortsContains(t, cur, errTrustLevelViolated, func() {
		_ = exec.Execute(cross(cur))
	})
}

func TestNewValidatorProposalRequest_AllowsJustAboveTrustLevel(cur realm, t *testing.T) {
	// Companion to RejectsAtTrustLevelBoundary: same baseline, but
	// retained = threshold + 1 ⇒ must pass. Brackets the boundary on
	// the passing side so an off-by-one in the comparison surfaces.
	// Baseline: A:4, B:3, C:2 (prevTotal=9, threshold = 3).
	// Proposal removes B and C, retaining only A. retained = 4.
	// 4*3 = 12, 9*1 = 9, 12 <= 9 ⇒ false ⇒ passes.
	resetValset(t)
	resetCache()
	resetLimits()

	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":4",
		pubKeyB + ":3",
		pubKeyC + ":2",
	})
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	opC := testutils.TestAddress("op-C")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
		{op: opC, pubKey: pubKeyC, keepRunning: true},
	})

	exec := newValoperChangeExecutor(cur, []ValoperChange{
		{OperatorAddress: opB, Power: 0},
		{OperatorAddress: opC, Power: 0},
	}, trustLevelRatio)

	urequire.NoError(t, exec.Execute(cross(cur)))
}

func TestNewValidatorProposalRequest_RejectsJustBelowTrustLevel(cur realm, t *testing.T) {
	// Companion to RejectsAtTrustLevelBoundary: same baseline, but
	// retained = threshold-1 ⇒ must fail. Brackets the boundary on
	// the failing side so an off-by-one in the comparison surfaces.
	// Baseline: A:4, B:3, C:2 (prevTotal=9, threshold = 3).
	// Proposal removes A and B, retaining only C. retained = 2.
	// 2*3 = 6, 9*1 = 9, 6 <= 9 ⇒ true ⇒ aborts.
	resetValset(t)
	resetCache()
	resetLimits()

	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":4",
		pubKeyB + ":3",
		pubKeyC + ":2",
	})
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	opC := testutils.TestAddress("op-C")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
		{op: opC, pubKey: pubKeyC, keepRunning: true},
	})

	exec := newValoperChangeExecutor(cur, []ValoperChange{
		{OperatorAddress: opA, Power: 0},
		{OperatorAddress: opB, Power: 0},
	}, trustLevelRatio)

	uassert.AbortsContains(t, cur, errTrustLevelViolated, func() {
		_ = exec.Execute(cross(cur))
	})
}

func TestNewValidatorProposalRequest_ExecutorUsesSnapshotTrustLevel(cur realm, t *testing.T) {
	// Snapshot semantics: the trust-level rule the executor enforces
	// is the ratio that was in effect at proposal-creation time, NOT
	// the package-level trustLevelRatio at execute-time. This closes
	// the same-block ordering attack: a trust-level-loosening
	// proposal executed first can't relax the rule for a valset-
	// change proposal that was created under the stricter rule.
	//
	// Baseline: A:1, B:1, C:1 (prevTotal=3). Proposal removes A.
	// retained (BASELINE VP of survivors) = 2 (B+C).
	//   under 1/3 ratio: 2*3=6, 3*1=3, 6 <= 3 ⇒ false ⇒ would pass
	//   under 2/3 ratio: 2*3=6, 3*2=6, 6 <= 6 ⇒ true  ⇒ would fail
	// We snapshot at 2/3, then drop the package ratio to 1/3, then
	// execute — the snapshot wins.
	resetValset(t)
	resetCache()
	resetLimits()

	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":1",
		pubKeyB + ":1",
		pubKeyC + ":1",
	})
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	opC := testutils.TestAddress("op-C")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
		{op: opC, pubKey: pubKeyC, keepRunning: true},
	})

	// Build the executor under the stricter ratio.
	stricter := trustRatio{numerator: 2, denominator: 3}
	exec := newValoperChangeExecutor(cur, []ValoperChange{
		{OperatorAddress: opA, Power: 0},
	}, stricter)

	// Loosen the package-level ratio AFTER proposal creation. If the
	// executor read the package var, this would let the change slip
	// through.
	trustLevelRatio = trustRatio{numerator: 1, denominator: 3}

	uassert.AbortsContains(t, cur, errTrustLevelViolated, func() {
		_ = exec.Execute(cross(cur))
	})
}

func TestNewValidatorProposalRequest_RejectsBaselineWipeViaUpsert(cur realm, t *testing.T) {
	// Regression test for the OLD-vs-NEW weights bug: a proposal that
	// removes most baseline validators and inflates the remaining
	// survivor's voting power must be rejected, because an IBC light
	// client trusting the baseline would refuse to verify the
	// resulting header (its tally uses baseline VPs, not the inflated
	// new VP).
	//
	// Baseline: A:1, B:1, C:1 (prevTotal=3, threshold = 1 at 1/3,
	// strict cross-multiply: pass iff retained*3 > 3, i.e. retained > 1).
	// Proposal: remove A, remove B, upsert C → 100.
	// Survivors: {C}. retained = C's BASELINE VP = 1.
	// 1*3 = 3, 3*1 = 3, 3 <= 3 ⇒ aborts. ✓
	//
	// If the executor had used NEW weights (the previous bug), retained
	// would be 100 (C's NEW VP), and the check would have passed —
	// false-positive accept of a proposal that breaks IBC safety.
	resetValset(t)
	resetCache()
	resetLimits()

	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":1",
		pubKeyB + ":1",
		pubKeyC + ":1",
	})
	opA := testutils.TestAddress("op-A")
	opB := testutils.TestAddress("op-B")
	opC := testutils.TestAddress("op-C")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{
		{op: opA, pubKey: pubKeyA, keepRunning: true},
		{op: opB, pubKey: pubKeyB, keepRunning: true},
		{op: opC, pubKey: pubKeyC, keepRunning: true},
	})

	exec := newValoperChangeExecutor(cur, []ValoperChange{
		{OperatorAddress: opA, Power: 0},
		{OperatorAddress: opB, Power: 0},
		{OperatorAddress: opC, Power: 100},
	}, trustLevelRatio)

	uassert.AbortsContains(t, cur, errTrustLevelViolated, func() {
		_ = exec.Execute(cross(cur))
	})
}

func TestNewTrustLevelPropRequest_UpdatesRatio(cur realm, t *testing.T) {
	resetLimits()

	numBefore, denBefore := GetTrustLevel()
	uassert.Equal(t, uint64(1), numBefore)
	uassert.Equal(t, uint64(3), denBefore)

	// NewTrustLevelPropRequest validates at create-time; the executor is
	// re-built here to drive it directly (ProposalRequest doesn't expose
	// its Executor). Going through dao.NewSimpleExecutor would duplicate
	// the factory's branching, so use the private helper instead.
	_ = NewTrustLevelPropRequest(cur, 2, 5, "tighten trust level", "require 40% legacy power")
	exec := newTrustLevelExecutor(cur, trustRatio{numerator: 2, denominator: 5})
	urequire.NoError(t, exec.Execute(cross(cur)))

	numAfter, denAfter := GetTrustLevel()
	uassert.Equal(t, uint64(2), numAfter)
	uassert.Equal(t, uint64(5), denAfter)
}

func TestNewTrustLevelPropRequest_RejectsOutOfBounds(cur realm, t *testing.T) {
	resetLimits()

	// numerator/denominator must yield a ratio in [1/3, 2/3].
	// Out-of-band overflow is caught at execute-time by overflow.Mulu64
	// in the trust-level check, not at create-time — see
	// TestNewValidatorProposalRequest_ExecutorRejectsRatioOverflow.
	cases := []struct{ num, den uint64 }{
		{0, 3}, // 0/3 < 1/3
		{1, 4}, // 1/4 < 1/3
		{3, 4}, // 3/4 > 2/3
		{3, 3}, // 1/1 > 2/3
		{1, 0}, // div-by-zero guard
	}
	for _, c := range cases {
		uassert.PanicsContains(t, cur, errInvalidTrustLevel, func() {
			_ = NewTrustLevelPropRequest(cur, c.num, c.den, "invalid", "")
		})
	}
}

func TestNewCooldownPropRequest_UpdatesCooldown(cur realm, t *testing.T) {
	resetLimits()

	cdBefore := GetCooldown()
	uassert.Equal(t, uint64(24*60*60), cdBefore) // 24h in seconds

	// NewCooldownPropRequest validates at create-time; the executor is
	// re-built here to drive it directly (ProposalRequest doesn't
	// expose its Executor).
	_ = NewCooldownPropRequest(cur, 0, "disable cooldown for testnet", "")
	exec := newCooldownExecutor(cur, 0)
	urequire.NoError(t, exec.Execute(cross(cur)))

	cdAfter := GetCooldown()
	uassert.Equal(t, uint64(0), cdAfter)
}

func TestNewCooldownPropRequest_RejectsOverflowingSeconds(cur realm, t *testing.T) {
	resetLimits()

	// Any input above cooldownMaxSeconds is rejected at create-time.
	// Without this guard, time.Duration(seconds) * time.Second
	// silently overflows int64 past ~9.22e9 seconds (≈292 years) and
	// could wrap to a tiny or negative duration, effectively
	// disabling the cooldown.
	cases := []uint64{
		cooldownMaxSeconds + 1, // just above the cap
		1 << 33,                // ≈272 years, comfortably overflowing
		^uint64(0),             // uint64 max
	}
	for _, s := range cases {
		uassert.PanicsContains(t, cur, errInvalidCooldown, func() {
			_ = NewCooldownPropRequest(cur, s, "too long", "")
		})
	}

	// Cap itself is accepted.
	urequire.NotPanics(t, cur, func() {
		_ = NewCooldownPropRequest(cur, cooldownMaxSeconds, "at cap", "")
	})
}

func TestNewCooldownPropRequest_AllowsBackToBackAfterDisable(cur realm, t *testing.T) {
	// Concrete test of the use case the proposal exists for: after a
	// cooldown=0 proposal executes, subsequent valset proposals are
	// no longer blocked by the cooldown (creation OR execution).
	resetValset(t)
	resetCache()
	resetLimits()

	// Pretend a valset update just happened.
	lastValsetUpdate = time.Now()

	// Disable the cooldown via governance.
	urequire.NoError(t, newCooldownExecutor(cur, 0).Execute(cross(cur)))

	// Now a back-to-back valset proposal must pass cooldown checks.
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

	// Create-time cooldown check uses live cooldown (now 0) — passes.
	pr := NewValidatorProposalRequest(cur,
		[]ValoperChange{{OperatorAddress: opA, Power: 5}},
		"upsert opA after cooldown disabled", "",
	)
	_ = pr

	// Execute path: same live check, also passes.
	exec := newValoperChangeExecutor(cur,
		[]ValoperChange{{OperatorAddress: opA, Power: 5}},
		trustLevelRatio,
	)
	urequire.NoError(t, exec.Execute(cross(cur)))
}

func TestNewValidatorProposalRequest_ExecutorRejectsRatioOverflow(cur realm, t *testing.T) {
	// Defense-in-depth: a maliciously-crafted (but in-bounds) ratio
	// like (2<<40, 3<<40) is accepted by NewTrustLevelPropRequest
	// (ratio = 2/3, within [1/3, 2/3]). If a future baseline grows
	// large enough that baselineTotal * numerator overflows uint64,
	// overflow.Mulu64 in the trust-level check aborts rather than
	// returning a silently-wrong comparison.
	//
	// To trigger overflow at execute-time we need baselineTotal *
	// snapshotTrustLevel.numerator to exceed 2^64. With one validator
	// at VP=2^40, total=2^40, and a snapshot ratio with num=2^40,
	// the product is 2^80 — definite overflow.
	resetValset(t)
	resetCache()
	resetLimits()

	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":1099511627776", // 2^40
	})
	opA := testutils.TestAddress("op-A")
	seedCache(t, []struct {
		op          address
		pubKey      string
		keepRunning bool
	}{{op: opA, pubKey: pubKeyA, keepRunning: true}})

	// Build the executor with a ratio whose numerator and denominator
	// are individually fine, but whose product against a 2^40 baseline
	// overflows uint64.
	bigButValid := trustRatio{numerator: 1 << 40, denominator: 3 << 40}
	exec := newValoperChangeExecutor(cur, []ValoperChange{
		{OperatorAddress: opA, Power: 1 << 40},
	}, bigButValid)

	uassert.AbortsContains(t, cur, errTrustLevelOverflow, func() {
		_ = exec.Execute(cross(cur))
	})
}
