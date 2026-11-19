package names

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestDefaultVerifier(t *testing.T) {
	// Disabled: any case is true (regardless of paused or lookup).
	uassert.True(t, verifier(false, false, alice, alice.String(), nil, nil))
	uassert.True(t, verifier(false, false, "", alice.String(), nil, nil))
	uassert.True(t, verifier(false, false, alice, "somerandomusername", nil, nil))

	// Pre-Enable bypass takes priority over paused — even when paused=true,
	// !isEnabled returns true. This is the documented "stage pause before
	// Enable" behavior.
	uassert.True(t, verifier(false, true, alice, alice.String(), nil, nil))
	uassert.True(t, verifier(false, true, alice, "anything", nil, nil))

	// Enabled: PA namespace check.
	uassert.True(t, verifier(true, false, alice, alice.String(), nil, nil))

	// Enabled: non-PA namespaces denied when no registered-name lookup.
	uassert.False(t, verifier(true, false, alice, "notregistered", nil, nil))
	uassert.False(t, verifier(true, false, alice, "alice", nil, nil))

	// Enabled: empty name/address.
	uassert.False(t, verifier(true, false, address(""), "", nil, nil))
	uassert.False(t, verifier(true, false, alice, "", nil, nil))
	uassert.False(t, verifier(true, false, address(""), "something", nil, nil))
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
	uassert.True(t, verifier(true, false, alice, "gnobody", aliceOwnsGnobody, nil))

	// bob cannot deploy under "gnobody" (registered to alice, not bob).
	uassert.False(t, verifier(true, false, bob, "gnobody", aliceOwnsGnobody, nil))

	// Unknown name still rejects (lookup returns ok=false).
	uassert.False(t, verifier(true, false, alice, "notregistered", aliceOwnsGnobody, nil))

	// PA path still works alongside registered-name lookup. Tried first.
	uassert.True(t, verifier(true, false, alice, alice.String(), aliceOwnsGnobody, nil))

	// Disabled bypasses both paths entirely (returns true regardless).
	uassert.True(t, verifier(false, false, bob, "gnobody", aliceOwnsGnobody, nil))

	// Empty namespace + non-nil lookup: still rejects (early-out before lookup).
	uassert.False(t, verifier(true, false, alice, "", aliceOwnsGnobody, nil))

	// Lookup returns ok=true for current-but-mismatched address.
	bobOwnsGnobody := func(name string) (address, bool) {
		if name == "gnobody" {
			return bob, true
		}
		return "", false
	}
	uassert.False(t, verifier(true, false, alice, "gnobody", bobOwnsGnobody, nil))
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
		uassert.False(t, verifier(true, true, alice, alice.String(), nil, nil))
		uassert.False(t, verifier(true, true, alice, alice.String(), aliceOwnsGnobody, nil))
	})

	t.Run("paused blocks registered-name namespaces", func(t *testing.T) {
		// Registered-name path would normally succeed, but pause
		// short-circuits before the lookup is consulted.
		uassert.False(t, verifier(true, true, alice, "gnobody", aliceOwnsGnobody, nil))
	})

	t.Run("paused blocks any other namespace too", func(t *testing.T) {
		uassert.False(t, verifier(true, true, alice, "anything", aliceOwnsGnobody, nil))
		uassert.False(t, verifier(true, true, alice, "", aliceOwnsGnobody, nil))
		uassert.False(t, verifier(true, true, address(""), "anything", aliceOwnsGnobody, nil))
	})

	t.Run("pre-Enable bypass takes priority over paused", func(t *testing.T) {
		// !enabled is checked BEFORE paused in the verifier — the
		// pre-Enable shortcut returns true regardless of pause state.
		// This is the "pause is staged, applies on Enable" semantic.
		uassert.True(t, verifier(false, true, alice, alice.String(), nil, nil))
		uassert.True(t, verifier(false, true, alice, "gnobody", aliceOwnsGnobody, nil))
		uassert.True(t, verifier(false, true, alice, "anything", nil, nil))
	})

	t.Run("paused=false with enabled=true behaves normally", func(t *testing.T) {
		// Sanity: just confirms the false case doesn't accidentally do
		// something different from the no-pause baseline.
		uassert.True(t, verifier(true, false, alice, alice.String(), nil, nil))
		uassert.True(t, verifier(true, false, alice, "gnobody", aliceOwnsGnobody, nil))
		uassert.False(t, verifier(true, false, bob, "gnobody", aliceOwnsGnobody, nil))
	})
}

// TestProposeSetPaused_idempotency verifies the proposal-creation guard:
// requesting a pause state that already matches the current state panics
// at proposal-creation time, so voters never see a no-op proposal.
//
// NOTE: this test mutates the package-level `paused` var directly to
// avoid driving the full DAO flow. Restores the value via defer.
func TestProposeSetPaused_idempotency(cur realm, t *testing.T) {
	saved := paused
	defer func() { paused = saved }()

	t.Run("propose pause when already paused panics", func(t *testing.T) {
		paused = true
		urequire.PanicsWithMessage(t, cur,
			"paused state already matches requested value; no-op proposal rejected",
			func() { ProposeSetPaused(cur, true) })
	})

	t.Run("propose unpause when already unpaused panics", func(t *testing.T) {
		paused = false
		urequire.PanicsWithMessage(t, cur,
			"paused state already matches requested value; no-op proposal rejected",
			func() { ProposeSetPaused(cur, false) })
	})

	t.Run("propose pause when unpaused succeeds", func(t *testing.T) {
		paused = false
		urequire.NotPanics(t, cur, func() { ProposeSetPaused(cur, true) })
		// The proposal-create call doesn't actually mutate paused —
		// only its execution would. Confirm.
		uassert.False(t, paused)
	})

	t.Run("propose unpause when paused succeeds", func(t *testing.T) {
		paused = true
		urequire.NotPanics(t, cur, func() { ProposeSetPaused(cur, false) })
		// Same: proposal-create does not mutate.
		uassert.True(t, paused)
	})
}

// TestSetPaused_internal verifies the private setPaused actuator
// directly. Bypasses the proposal flow to confirm the toggle and the
// event-emission path independently.
func TestSetPaused_internal(cur realm, t *testing.T) {
	saved := paused
	defer func() { paused = saved }()

	paused = false
	setPaused(0, cur, true)
	uassert.True(t, paused)

	setPaused(0, cur, false)
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

func TestEnable(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("random")))
	uassert.AbortsWithMessage(t, cur, "caller is not admin", func() {
		Enable(cross(cur))
	})
	uassert.False(t, IsEnabled())

	testing.SetRealm(testing.NewUserRealm(admin))
	uassert.NotPanics(t, cur, func() {
		Enable(cross(cur))
	})
	uassert.True(t, IsEnabled())
}

// --- path 3: teams ----------------------------------------------------------

// aliceOwnsTeam is a teamLookupFn where alice is in "shared"'s team and the
// namespace has NOT retired its registered-name path.
func aliceOwnsTeam(namespace string, addr address) (bool, bool) {
	return namespace == "shared" && addr == alice, false
}

// aliceOwnsRetiredTeam is the same team, after the namespace retired path 2.
func aliceOwnsRetiredTeam(namespace string, addr address) (bool, bool) {
	if namespace != "shared" {
		return false, false
	}
	return addr == alice, true
}

func bobOwnsShared(name string) (address, bool) {
	if name == "shared" {
		return bob, true
	}
	return "", false
}

func TestTeamVerifier(t *testing.T) {
	// A team member deploys under the namespace.
	uassert.True(t, verifier(true, false, alice, "shared", nil, aliceOwnsTeam))

	// Someone who is not in the team does not.
	uassert.False(t, verifier(true, false, bob, "shared", nil, aliceOwnsTeam))

	// The team only covers its own namespace.
	uassert.False(t, verifier(true, false, alice, "other", nil, aliceOwnsTeam))

	// Paths 2 and 3 are OR'd while the namespace has not retired: bob holds the
	// registered name, alice is in the team, and both may deploy. This is what
	// lets a migration publish home before giving up the original key.
	uassert.True(t, verifier(true, false, bob, "shared", bobOwnsShared, aliceOwnsTeam))
	uassert.True(t, verifier(true, false, alice, "shared", bobOwnsShared, aliceOwnsTeam))
}

func TestRetiredNamespaceDropsTheRegisteredNamePath(t *testing.T) {
	// Before retiring, the registered-name holder may deploy.
	uassert.True(t, verifier(true, false, bob, "shared", bobOwnsShared, aliceOwnsTeam))

	// After retiring, the same holder may not, and the team still may. This is
	// the entire point of the mechanism: the name has not moved, and the
	// original key has stopped being able to act.
	uassert.False(t, verifier(true, false, bob, "shared", bobOwnsShared, aliceOwnsRetiredTeam))
	uassert.True(t, verifier(true, false, alice, "shared", bobOwnsShared, aliceOwnsRetiredTeam))
}

func TestRetirementDoesNotTouchPersonalAddressNamespaces(t *testing.T) {
	// The PA check sits above path 2, so a namespace that is literally an
	// address keeps working no matter what any team says about it. Pinned
	// because retiring reads like "this namespace is locked down" and must not
	// be allowed to grow into that.
	retiredEverything := func(namespace string, addr address) (bool, bool) {
		return false, true
	}
	uassert.True(t, verifier(true, false, alice, alice.String(), nil, retiredEverything))
}

func TestPauseAndPreEnableStillOutrankTeams(t *testing.T) {
	// Pause rejects a team member like everyone else.
	uassert.False(t, verifier(true, true, alice, "shared", nil, aliceOwnsTeam))

	// Pre-Enable passes everyone, including an address no team names.
	uassert.True(t, verifier(false, false, bob, "shared", nil, aliceOwnsTeam))
}
