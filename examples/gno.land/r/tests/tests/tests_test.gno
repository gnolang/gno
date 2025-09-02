package tests_test

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/r/tests/tests"
)

func TestAssertOriginCall(t *testing.T) {
	// CallAssertOriginCall(): no panic
	caller := testutils.TestAddress("caller")
	testing.SetRealm(std.NewUserRealm(caller))
	tests.CallAssertOriginCall(cross)
	if !tests.CallIsOriginCall(cross) {
		t.Errorf("expected IsOriginCall=true but got false")
	}

	testing.SetRealm(std.NewCodeRealm("gno.land/r/tests/tests"))
	// CallAssertOriginCall() from a block: abort
	r := revive(func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal block).
		if tests.CallIsOriginCall(cross) {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		tests.CallAssertOriginCall(cross) // <---
	})
	if fmt.Sprintf("%v", r) != "invalid non-origin call" {
		t.Error("expected abort but did not")
	}
	// CallSubtestsAssertOriginCall(): abort
	r = revive(func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal block).
		if tests.CallSubtestsIsOriginCall(cross) {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		tests.CallSubtestsAssertOriginCall(cross)
	})
	if fmt.Sprintf("%v", r) != "invalid non-origin call" {
		t.Error("expected abort but did not")
	}

}

func TestPreviousRealm(t *testing.T) {
	var (
		firstRealm = std.DerivePkgAddr("gno.land/r/tests_test")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/tests/tests")
	)
	// When only one realm in the frames, PreviousRealm returns the same realm
	if addr := tests.GetPreviousRealm(cross).Address(); addr != firstRealm {
		println(tests.GetPreviousRealm(cross))
		t.Errorf("want GetPreviousRealm().Address==%s, got %s", firstRealm, addr)
	}
	// When 2 or more realms in the frames, PreviousRealm returns the second to last
	if addr := tests.GetRSubtestsPreviousRealm(cross).Address(); addr != rTestsAddr {
		t.Errorf("want GetRSubtestsPreviousRealm().Address==%s, got %s", rTestsAddr, addr)
	}
}
