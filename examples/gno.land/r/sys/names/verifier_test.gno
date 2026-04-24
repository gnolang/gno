package names

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestDefaultVerifier(t *testing.T) {
	// Disabled: any case is true (regardless of lookup).
	uassert.True(t, verifier(false, alice, alice.String(), nil))
	uassert.True(t, verifier(false, "", alice.String(), nil))
	uassert.True(t, verifier(false, alice, "somerandomusername", nil))

	// Enabled: PA namespace check.
	uassert.True(t, verifier(true, alice, alice.String(), nil))

	// Enabled: non-PA namespaces denied when no registered-name lookup.
	uassert.False(t, verifier(true, alice, "notregistered", nil))
	uassert.False(t, verifier(true, alice, "alice", nil))

	// Enabled: empty name/address.
	uassert.False(t, verifier(true, address(""), "", nil))
	uassert.False(t, verifier(true, alice, "", nil))
	uassert.False(t, verifier(true, address(""), "something", nil))
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
	uassert.True(t, verifier(true, alice, "gnobody", aliceOwnsGnobody))

	// bob cannot deploy under "gnobody" (registered to alice, not bob).
	uassert.False(t, verifier(true, bob, "gnobody", aliceOwnsGnobody))

	// Unknown name still rejects (lookup returns ok=false).
	uassert.False(t, verifier(true, alice, "notregistered", aliceOwnsGnobody))

	// PA path still works alongside registered-name lookup. Tried first.
	uassert.True(t, verifier(true, alice, alice.String(), aliceOwnsGnobody))

	// Disabled bypasses both paths entirely (returns true regardless).
	uassert.True(t, verifier(false, bob, "gnobody", aliceOwnsGnobody))

	// Empty namespace + non-nil lookup: still rejects (early-out before lookup).
	uassert.False(t, verifier(true, alice, "", aliceOwnsGnobody))

	// Lookup returns ok=false for current-but-mismatched address.
	bobOwnsGnobody := func(name string) (address, bool) {
		if name == "gnobody" {
			return bob, true
		}
		return "", false
	}
	uassert.False(t, verifier(true, alice, "gnobody", bobOwnsGnobody))
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
