package names

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestDefaultVerifier(t *testing.T) {
	// Disabled: any case is true (regardless of paused or lookup).
	uassert.True(t, verifier(false, false, alice, alice.String(), nil))
	uassert.True(t, verifier(false, false, "", alice.String(), nil))
	uassert.True(t, verifier(false, false, alice, "somerandomusername", nil))

	// Pre-Enable bypass takes priority over paused — even when paused=true,
	// !isEnabled returns true. This is the documented "stage pause before
	// Enable" behavior.
	uassert.True(t, verifier(false, true, alice, alice.String(), nil))
	uassert.True(t, verifier(false, true, alice, "anything", nil))

	// Enabled: PA namespace check.
	uassert.True(t, verifier(true, false, alice, alice.String(), nil))

	// Enabled: non-PA namespaces denied when no registered-name lookup.
	uassert.False(t, verifier(true, false, alice, "notregistered", nil))
	uassert.False(t, verifier(true, false, alice, "alice", nil))

	// Enabled: empty name/address.
	uassert.False(t, verifier(true, false, address(""), "", nil))
	uassert.False(t, verifier(true, false, alice, "", nil))
	uassert.False(t, verifier(true, false, address(""), "something", nil))
}

// TestRegisteredNameVerifier exercises the second authorization path —
// the registered-name lookup. Uses a fake lookup so this test doesn't
// depend on r/sys/users state.
func TestRegisteredNameVerifier(t *testing.T) {
	// Lookup that says "gnobody" is registered to alice.
	aliceOwnsGnobody := func(name string) (address, bool) {
		if name == "gnobody" {
			return alice, true
		}
		return "", false
	}

	// alice can deploy under "gnobody" (her registered name).
	uassert.True(t, verifier(true, false, alice, "gnobody", aliceOwnsGnobody))

	// bob cannot deploy under "gnobody" (registered to alice, not bob).
	uassert.False(t, verifier(true, false, bob, "gnobody", aliceOwnsGnobody))

	// Unknown name still rejects (lookup returns ok=false).
	uassert.False(t, verifier(true, false, alice, "notregistered", aliceOwnsGnobody))

	// PA path still works alongside registered-name lookup. Tried first.
	uassert.True(t, verifier(true, false, alice, alice.String(), aliceOwnsGnobody))

	// Disabled bypasses both paths entirely (returns true regardless).
	uassert.True(t, verifier(false, false, bob, "gnobody", aliceOwnsGnobody))

	// Empty namespace + non-nil lookup: still rejects (early-out before lookup).
	uassert.False(t, verifier(true, false, alice, "", aliceOwnsGnobody))

	// Lookup returns ok=true for current-but-mismatched address.
	bobOwnsGnobody := func(name string) (address, bool) {
		if name == "gnobody" {
			return bob, true
		}
		return "", false
	}
	uassert.False(t, verifier(true, false, alice, "gnobody", bobOwnsGnobody))
}

// TestPausedVerifier covers the emergency-halt semantic. When paused
// AND enabled, the verifier rejects EVERY namespace check including PA.
func TestPausedVerifier(t *testing.T) {
	// Lookup that says "gnobody" is registered to alice. Used to confirm
	// the registered-name path is also halted by pause.
	aliceOwnsGnobody := func(name string) (address, bool) {
		if name == "gnobody" {
			return alice, true
		}
		return "", false
	}

	t.Run("paused blocks PA namespaces", func(t *testing.T) {
		// PA would normally succeed (addr.String() == namespace), but
		// pause short-circuits before the PA check.
		uassert.False(t, verifier(true, true, alice, alice.String(), nil))
		uassert.False(t, verifier(true, true, alice, alice.String(), aliceOwnsGnobody))
	})

	t.Run("paused blocks registered-name namespaces", func(t *testing.T) {
		// Registered-name path would normally succeed, but pause
		// short-circuits before the lookup is consulted.
		uassert.False(t, verifier(true, true, alice, "gnobody", aliceOwnsGnobody))
	})

	t.Run("paused blocks any other namespace too", func(t *testing.T) {
		uassert.False(t, verifier(true, true, alice, "anything", aliceOwnsGnobody))
		uassert.False(t, verifier(true, true, alice, "", aliceOwnsGnobody))
		uassert.False(t, verifier(true, true, address(""), "anything", aliceOwnsGnobody))
	})

	t.Run("pre-Enable bypass takes priority over paused", func(t *testing.T) {
		// !enabled is checked BEFORE paused in the verifier — the
		// pre-Enable shortcut returns true regardless of pause state.
		// This is the "pause is staged, applies on Enable" semantic.
		uassert.True(t, verifier(false, true, alice, alice.String(), nil))
		uassert.True(t, verifier(false, true, alice, "gnobody", aliceOwnsGnobody))
		uassert.True(t, verifier(false, true, alice, "anything", nil))
	})

	t.Run("paused=false with enabled=true behaves normally", func(t *testing.T) {
		// Sanity: just confirms the false case doesn't accidentally do
		// something different from the no-pause baseline.
		uassert.True(t, verifier(true, false, alice, alice.String(), nil))
		uassert.True(t, verifier(true, false, alice, "gnobody", aliceOwnsGnobody))
		uassert.False(t, verifier(true, false, bob, "gnobody", aliceOwnsGnobody))
	})
}

// TestProposeSetPaused_idempotency verifies the proposal-creation guard:
// requesting a pause state that already matches the current state panics
// at proposal-creation time, so voters never see a no-op proposal.
//
// NOTE: this test mutates the package-level `paused` var directly to
// avoid driving the full DAO flow. Restores the value via defer.
func TestProposeSetPaused_idempotency(t *testing.T) {
	saved := paused
	defer func() { paused = saved }()

	t.Run("propose pause when already paused panics", func(t *testing.T) {
		paused = true
		urequire.PanicsWithMessage(t,
			"paused state already matches requested value; no-op proposal rejected",
			func() { ProposeSetPaused(true) })
	})

	t.Run("propose unpause when already unpaused panics", func(t *testing.T) {
		paused = false
		urequire.PanicsWithMessage(t,
			"paused state already matches requested value; no-op proposal rejected",
			func() { ProposeSetPaused(false) })
	})

	t.Run("propose pause when unpaused succeeds", func(t *testing.T) {
		paused = false
		urequire.NotPanics(t, func() { ProposeSetPaused(true) })
		// The proposal-create call doesn't actually mutate paused —
		// only its execution would. Confirm.
		uassert.False(t, paused)
	})

	t.Run("propose unpause when paused succeeds", func(t *testing.T) {
		paused = true
		urequire.NotPanics(t, func() { ProposeSetPaused(false) })
		// Same: proposal-create does not mutate.
		uassert.True(t, paused)
	})
}

// TestSetPaused_internal verifies the private setPaused actuator
// directly. Bypasses the proposal flow to confirm the toggle and the
// event-emission path independently.
func TestSetPaused_internal(t *testing.T) {
	saved := paused
	defer func() { paused = saved }()

	paused = false
	setPaused(true)
	uassert.True(t, paused)

	setPaused(false)
	uassert.False(t, paused)
}

func TestIsPaused(t *testing.T) {
	saved := paused
	defer func() { paused = saved }()

	paused = false
	uassert.False(t, IsPaused())

	paused = true
	uassert.True(t, IsPaused())
}

func TestEnable(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("random")))
	uassert.AbortsWithMessage(t, "caller is not admin", func() {
		Enable(cross)
	})
	uassert.False(t, IsEnabled())

	testing.SetRealm(testing.NewUserRealm(admin))
	uassert.NotPanics(t, func() {
		Enable(cross)
	})
	uassert.True(t, IsEnabled())
}
