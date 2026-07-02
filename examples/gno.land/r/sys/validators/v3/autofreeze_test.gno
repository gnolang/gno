// autofreeze_test.gno — unit tests for Phase B auto-freeze state helpers.
//
// Auth-gated functions (AutoFreezeValidator, RegisterMonitor, etc.) require
// a registered monitor as OriginCaller or a GovDAO call context — both
// unavailable in the test runner. State is mutated directly to test the
// invariant logic and read helpers.
package validators

import (
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/urequire/v0"
)

func resetAutoFreezeState(t *testing.T) {
	t.Helper()
	monitors = avl.NewTree()
	protectedSet = avl.NewTree()
	autoFrozenSet = avl.NewTree()
}

// ---- monitor set ------------------------------------------------------------

func TestMonitors_DefaultEmpty(t *testing.T) {
	resetAutoFreezeState(t)
	urequire.Equal(t, 0, len(GetMonitors()))
	urequire.False(t, IsMonitor(address("g1monitor")))
}

func TestMonitors_ManualAddRemove(t *testing.T) {
	resetAutoFreezeState(t)

	m := address("g1monitor")
	urequire.False(t, IsMonitor(m))

	monitors.Set(m.String(), struct{}{})
	urequire.True(t, IsMonitor(m))
	urequire.Equal(t, 1, len(GetMonitors()))

	monitors.Remove(m.String())
	urequire.False(t, IsMonitor(m))
}

// ---- protected set ----------------------------------------------------------

func TestProtected_DefaultEmpty(t *testing.T) {
	resetAutoFreezeState(t)
	urequire.False(t, IsProtected(address("g1val")))
}

func TestProtected_ManualAddRemove(t *testing.T) {
	resetAutoFreezeState(t)

	v := address("g1val")
	protectedSet.Set(v.String(), struct{}{})
	urequire.True(t, IsProtected(v))

	protectedSet.Remove(v.String())
	urequire.False(t, IsProtected(v))
}

// ---- auto-frozen set --------------------------------------------------------

func TestIsAutoFrozen_Default(t *testing.T) {
	resetAutoFreezeState(t)
	urequire.False(t, IsAutoFrozen(address("g1val")))
	urequire.Equal(t, 0, len(GetAutoFrozenValidators()))
}

func TestIsAutoFrozen_AfterManualInsert(t *testing.T) {
	resetAutoFreezeState(t)

	addr := address("g1autoval")
	autoFrozenSet.Set(addr.String(), autoFrozenEntry{
		Addr:      addr,
		PubKey:    "pk",
		Power:     10,
		Reason:    "downtime",
		Evidence:  "blocks 100-200 missed",
		Monitor:   address("g1monitor"),
		FrozenAt:  50,
		ExpiresAt: 50 + autoFreezeExpiryBlocks,
	})

	urequire.True(t, IsAutoFrozen(addr))
	vals := GetAutoFrozenValidators()
	urequire.Equal(t, 1, len(vals))
	urequire.Equal(t, addr.String(), vals[0].String())
}

// ---- invariant checks (logic only, no crossing calls) -----------------------

func TestPowerFloor_Logic(t *testing.T) {
	// Verify the floor arithmetic: post-freeze live power must be >= 80% of total.
	totalPower := uint64(1000)
	targetPower := uint64(250) // removing 25%

	// livePower before freeze = 900 (100 already auto-frozen elsewhere)
	livePower := uint64(900)
	postFreeze := livePower - targetPower // 650

	// 650/1000 = 65% < 80% — should be rejected.
	urequire.True(t, postFreeze*100 < totalPower*powerFloorPct)

	// With a smaller target (50), post = 850, 85% >= 80% — should pass.
	smallTarget := uint64(50)
	postFreeze2 := livePower - smallTarget // 850
	urequire.False(t, postFreeze2*100 < totalPower*powerFloorPct)
}

func TestAutoFrozenCap_Logic(t *testing.T) {
	// Cap is 12% of total. Verify the arithmetic.
	totalPower := uint64(1000)
	alreadyFrozen := uint64(100) // 10% already auto-frozen
	newTarget := uint64(30)      // adding 3% → 13% total > 12% cap

	urequire.True(t, (alreadyFrozen+newTarget)*100 > totalPower*autoFrozenCapPct)

	// With newTarget = 10: 11% <= 12% — should pass.
	smallTarget := uint64(10)
	urequire.False(t, (alreadyFrozen+smallTarget)*100 > totalPower*autoFrozenCapPct)
}
