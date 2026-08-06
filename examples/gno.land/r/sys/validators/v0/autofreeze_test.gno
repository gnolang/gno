// autofreeze_test.gno — Phase B monitor-driven freeze.
//
// These drive AutoFreezeValidator and ExpireAutoFrozen for real,
// against a seeded valset and valoperCache, with the monitor's realm
// frame set the way a live transaction would set it. The invariants are
// what this file is about, so each one gets a case that trips it and a
// case that does not.
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

// resetAutoFreezeState clears everything autofreeze.gno owns. Wired
// into resetFreezeState so no test can inherit a monitor or an
// auto-frozen entry.
func resetAutoFreezeState() {
	monitors = avl.NewTree()
	protectedSet = avl.NewTree()
	autoFrozenSet = avl.NewTree()
}

// registerMonitor registers a monitor and returns its address. Callers
// must follow up with testing.SetRealm in the test body itself —
// SetRealm applies to the calling frame, so doing it in here would not
// survive the return.
func registerMonitor(t *testing.T) address {
	t.Helper()
	m := testutils.TestAddress("monitor-1")
	monitors.Set(m.String(), struct{}{})
	return m
}

// ---- happy path -------------------------------------------------------------------

func TestAutoFreeze_RemovesOperatorAndRecordsExpiry(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	// A light validator: 10 of 210 is under both the 20% floor headroom
	// and the 12% cap.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})
	m := registerMonitor(t)
	testing.SetRealm(testing.NewUserRealm(m))

	AutoFreezeValidator(cross(cur), opA, "missed 200 blocks", "heights 100-300")

	urequire.True(t, IsAutoFrozen(opA), "recorded as auto-frozen")
	uassert.False(t, IsFrozen(opA), "manual freeze set untouched")

	powers := powerByPubKey()
	urequire.Equal(t, 2, len(powers), "operator pulled from the valset")
	_, still := powers[pubKeyA]
	uassert.False(t, still)

	raw := autoFrozenSet.Get(opA.String())
	urequire.NotNil(t, raw)
	e := raw.(autoFrozenEntry)
	uassert.Equal(t, m, e.Monitor, "monitor recorded")
	uassert.Equal(t, uint64(10), e.Power)
	uassert.Equal(t, "heights 100-300", e.Evidence)
	uassert.Equal(t, e.FrozenAt+autoFreezeExpiryBlocks, e.ExpiresAt, "expiry stamped")
}

// ---- invariant 1: power floor -------------------------------------------------------

func TestAutoFreeze_RejectsBelowPowerFloor(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	// 30 of 100 live power: freezing it leaves 70%, under the 80% floor.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":30", pubKeyB + ":35", pubKeyC + ":35",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))

	uassert.AbortsContains(t, cur, "below the 80% floor", func() {
		AutoFreezeValidator(cross(cur), opA, "too heavy", "evidence")
	})
	uassert.False(t, IsAutoFrozen(opA))
	uassert.Equal(t, 3, len(sysparams.GetValsetEffective()), "valset unchanged")
}

// The floor is measured against the whole roster, so a validator held
// out by a MANUAL freeze still counts in the denominator. Measuring
// against live power alone would let each freeze shrink the base and
// erode the floor one validator at a time.
func TestAutoFreeze_FloorDenominatorIncludesManuallyFrozenPower(cur realm, t *testing.T) {
	opA, opB, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":20", pubKeyC + ":100",
	})

	// Manually freeze opB: live power drops to 110, the roster stays 130.
	doFreeze(cur, opB, "manual")
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))

	// Freezing opA leaves 100 live. Against the roster that is
	// 100/130 = 77%, under the floor. Against live power alone it would
	// read 100/110 = 91% and sail through — which is exactly how a
	// shrinking denominator erodes the guarantee.
	uassert.AbortsContains(t, cur, "would leave 100 of 130 voting power live", func() {
		AutoFreezeValidator(cross(cur), opA, "check denominator", "evidence")
	})
	uassert.False(t, IsAutoFrozen(opA))
}

// ---- invariant 2: auto-frozen cap -----------------------------------------------------

func TestAutoFreeze_RejectsAboveAutoFrozenCap(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	// 13 of 100 exceeds the 12% cap, while 87% live clears the floor.
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":13", pubKeyB + ":44", pubKeyC + ":43",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))

	uassert.AbortsContains(t, cur, "above the 12% cap", func() {
		AutoFreezeValidator(cross(cur), opA, "over cap", "evidence")
	})
	uassert.False(t, IsAutoFrozen(opA))
}

// The cap is cumulative: a monitor cannot walk past it one small
// validator at a time.
func TestAutoFreeze_CapAccumulatesAcrossFreezes(cur realm, t *testing.T) {
	opA, opB, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":7", pubKeyB + ":7", pubKeyC + ":86",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))

	// 7% — fits.
	AutoFreezeValidator(cross(cur), opA, "first", "evidence")
	urequire.True(t, IsAutoFrozen(opA))

	// 7% + 7% = 14% — over the 12% cap, even though each alone fits.
	uassert.AbortsContains(t, cur, "above the 12% cap", func() {
		AutoFreezeValidator(cross(cur), opB, "second", "evidence")
	})
	uassert.False(t, IsAutoFrozen(opB))
}

// ---- invariant 3: protected set --------------------------------------------------------

func TestAutoFreeze_RejectsProtectedOperator(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})
	protectedSet.Set(opA.String(), struct{}{})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))

	uassert.AbortsContains(t, cur, "is protected and cannot be auto-frozen", func() {
		AutoFreezeValidator(cross(cur), opA, "protected", "evidence")
	})
	uassert.False(t, IsAutoFrozen(opA))
}

// Protection binds monitors only; the human lever still works.
func TestProtected_DoesNotBlockManualFreeze(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})
	protectedSet.Set(opA.String(), struct{}{})

	doFreeze(cur, opA, "governance override")
	uassert.True(t, IsFrozen(opA))
}

// ---- invariant 4: expiry -----------------------------------------------------------------

func TestExpireAutoFrozen_NoopBeforeExpiry(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))
	AutoFreezeValidator(cross(cur), opA, "downtime", "evidence")

	uassert.Equal(t, 0, ExpireAutoFrozen(cross(cur)), "nothing due yet")
	uassert.True(t, IsAutoFrozen(opA), "entry survives")
	uassert.Equal(t, 2, len(sysparams.GetValsetEffective()))
}

func TestExpireAutoFrozen_RestoresWhenDue(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))
	AutoFreezeValidator(cross(cur), opA, "downtime", "evidence")

	expireNow(t, opA)

	uassert.Equal(t, 1, ExpireAutoFrozen(cross(cur)), "one operator restored")
	uassert.False(t, IsAutoFrozen(opA), "entry cleared")
	powers := powerByPubKey()
	urequire.Equal(t, 3, len(powers))
	uassert.Equal(t, uint64(10), powers[pubKeyA], "restored at its recorded power")
}

// Same regression as the manual path: an operator who rotates while
// auto-frozen must come back under the NEW key.
func TestExpireAutoFrozen_RestoresUnderRotatedKey(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))
	AutoFreezeValidator(cross(cur), opA, "suspected compromise", "evidence")

	valoperCache.Set(opA.String(), cacheEntry{
		SigningPubKey:  pubKeyC,
		SigningAddress: mustAddr(t, pubKeyC),
		KeepRunning:    true,
	})
	expireNow(t, opA)

	urequire.Equal(t, 1, ExpireAutoFrozen(cross(cur)))
	powers := powerByPubKey()
	_, retired := powers[pubKeyA]
	uassert.False(t, retired, "the RETIRED key must not be republished")
	uassert.Equal(t, uint64(10), powers[pubKeyC], "restored under the rotated key")
}

// A stuck operator must not block the sweep for everyone else.
func TestExpireAutoFrozen_SkipsOptedOutOperatorWithoutBlockingTheSweep(cur realm, t *testing.T) {
	opA, opB, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":5", pubKeyB + ":5", pubKeyC + ":100",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))
	AutoFreezeValidator(cross(cur), opA, "downtime", "evidence")
	AutoFreezeValidator(cross(cur), opB, "downtime", "evidence")

	// opA opts out while frozen; opB is fine.
	valoperCache.Set(opA.String(), cacheEntry{
		SigningPubKey:  pubKeyA,
		SigningAddress: mustAddr(t, pubKeyA),
		KeepRunning:    false,
	})
	expireNow(t, opA)
	expireNow(t, opB)

	uassert.Equal(t, 1, ExpireAutoFrozen(cross(cur)), "only opB restored")
	uassert.False(t, IsAutoFrozen(opA), "opA's entry dropped, not left dangling")
	uassert.False(t, IsAutoFrozen(opB))

	powers := powerByPubKey()
	_, backA := powers[pubKeyA]
	uassert.False(t, backA, "opted-out operator not republished")
	uassert.Equal(t, uint64(5), powers[pubKeyB], "opB back in")
}

// ---- interaction with the manual path ------------------------------------------------------

func TestAutoFreeze_RejectsAlreadyManuallyFrozenOperator(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})
	doFreeze(cur, opA, "manual first")
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))

	uassert.AbortsContains(t, cur, "already manually frozen", func() {
		AutoFreezeValidator(cross(cur), opA, "again", "evidence")
	})
}

func TestAutoFreeze_RejectsDoubleAutoFreeze(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))
	AutoFreezeValidator(cross(cur), opA, "first", "evidence")

	uassert.AbortsContains(t, cur, "already auto-frozen", func() {
		AutoFreezeValidator(cross(cur), opA, "second", "evidence")
	})
}

// An auto-frozen operator is held out, so the manual count floor must
// see them in the roster too.
func TestManualFreeze_FloorCountsAutoFrozenValidators(cur realm, t *testing.T) {
	opA, opB, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":5", pubKeyB + ":50", pubKeyC + ":45",
	})
	testing.SetRealm(testing.NewUserRealm(registerMonitor(t)))
	AutoFreezeValidator(cross(cur), opA, "auto", "evidence")

	// Roster is still 3, so a manual freeze would leave 1 of 3.
	uassert.PanicsContains(t, cur, "would leave 1 of 3 validators live", func() {
		doFreeze(cur, opB, "manual")
	})
}

// ---- auth ---------------------------------------------------------------------------------

func TestAutoFreeze_RejectsUnregisteredCaller(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("stranger")))

	uassert.AbortsContains(t, cur, "caller is not a registered monitor", func() {
		AutoFreezeValidator(cross(cur), opA, "unauthorised", "evidence")
	})
	uassert.False(t, IsAutoFrozen(opA))
}

// tx-origin auth would accept this: the monitor signed the transaction.
// A monitor is an unattended hot key, so laundering matters more here
// than anywhere else in the realm.
func TestAutoFreeze_RejectsMonitorCallLaunderedThroughARealm(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})
	m := testutils.TestAddress("monitor-1")
	monitors.Set(m.String(), struct{}{})

	testing.SetOriginCaller(m)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/attacker"))

	uassert.AbortsContains(t, cur, "must call directly", func() {
		AutoFreezeValidator(cross(cur), opA, "laundered", "evidence")
	})
	uassert.False(t, IsAutoFrozen(opA))
}

func TestMonitors_EmptyByDefault(cur realm, t *testing.T) {
	seedThreeValidators(t)
	uassert.Equal(t, 0, len(GetMonitors()))
	uassert.False(t, IsMonitor(testutils.TestAddress("nobody")))
}

// ---- render ----------------------------------------------------------------------------------

func TestRenderAutoFreeze_SilentUntilUsed(cur realm, t *testing.T) {
	opA, _, _ := seedThreeValidators(t)
	testing.SetSysParamStrings(module, submodule, currKey, []string{
		pubKeyA + ":10", pubKeyB + ":100", pubKeyC + ":100",
	})

	out := Render("")
	uassert.False(t, strings.Contains(out, "## Monitors"), "silent with no monitors")

	m := registerMonitor(t)
	testing.SetRealm(testing.NewUserRealm(m))
	AutoFreezeValidator(cross(cur), opA, "missed blocks", "evidence")

	out = Render("")
	uassert.True(t, strings.Contains(out, "## Monitors"), "monitor section present")
	uassert.True(t, strings.Contains(out, m.String()), "monitor listed")
	uassert.True(t, strings.Contains(out, "## Auto-Frozen Validators"), "auto-frozen section present")
	uassert.True(t, strings.Contains(out, "missed blocks"), "reason listed")
}

// expireNow rewinds an auto-frozen entry's ExpiresAt to the current
// height, so ExpireAutoFrozen sees it as due. The test runner has no
// way to advance the chain height far enough to wait it out.
func expireNow(t *testing.T, op address) {
	t.Helper()
	raw := autoFrozenSet.Get(op.String())
	urequire.NotNil(t, raw, "operator must be auto-frozen")
	e := raw.(autoFrozenEntry)
	e.ExpiresAt = 0
	autoFrozenSet.Set(op.String(), e)
}
