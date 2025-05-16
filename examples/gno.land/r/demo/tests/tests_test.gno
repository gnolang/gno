package tests_test

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/r/demo/tests"
)

func TestAssertOriginCall(t *testing.T) {
	// CallAssertOriginCall(): no panic
	caller := testutils.TestAddress("caller")
	testing.SetRealm(std.NewUserRealm(caller))
	cross(tests.CallAssertOriginCall)()
	if !cross(tests.CallIsOriginCall)() {
		t.Errorf("expected IsOriginCall=true but got false")
	}

	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
	// CallAssertOriginCall() from a block: abort
	uassert.AbortsWithMessage(t, "invalid non-origin call", func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal block).
		if cross(tests.CallIsOriginCall)() {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		cross(tests.CallAssertOriginCall)() // <---
	})
	// CallSubtestsAssertOriginCall(): abort
	uassert.AbortsWithMessage(t, "invalid non-origin call", func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal block).
		if cross(tests.CallSubtestsIsOriginCall)() {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		cross(tests.CallSubtestsAssertOriginCall)()
	})
}

func TestPreviousRealm(t *testing.T) {
	var (
		firstRealm = std.DerivePkgAddr("gno.land/r/demo/tests_test")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/demo/tests")
	)
	// When only one realm in the frames, PreviousRealm returns the same realm
	if addr := cross(tests.GetPreviousRealm)().Address(); addr != firstRealm {
		println(cross(tests.GetPreviousRealm)())
		t.Errorf("want GetPreviousRealm().Address==%s, got %s", firstRealm, addr)
	}
	// When 2 or more realms in the frames, PreviousRealm returns the second to last
	if addr := cross(tests.GetRSubtestsPreviousRealm)().Address(); addr != rTestsAddr {
		t.Errorf("want GetRSubtestsPreviousRealm().Address==%s, got %s", rTestsAddr, addr)
	}
}
