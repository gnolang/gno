package tests_test

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/tests"
)

func TestAssertOriginCall(t *testing.T) {
	// CallAssertOriginCall(): no panic
	caller := testutils.TestAddress("caller")
	std.TestSetRealm(std.NewUserRealm(caller))
	tests.CallAssertOriginCall()
	if !tests.CallIsOriginCall() {
		t.Errorf("expected IsOriginCall=true but got false")
	}

	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
	// CallAssertOriginCall() from a block: panic
	expectedReason := "invalid non-origin call"
	func() {
		defer func() {
			r := recover()
			if r == nil || r.(string) != expectedReason {
				t.Errorf("expected panic with '%v', got '%v'", expectedReason, r)
			}
		}()
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal block).
		if tests.CallIsOriginCall() {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		tests.CallAssertOriginCall()
	}()

	// CallSubtestsAssertOriginCall(): panic
	defer func() {
		r := recover()
		if r == nil || r.(string) != expectedReason {
			t.Errorf("expected panic with '%v', got '%v'", expectedReason, r)
		}
	}()
	if tests.CallSubtestsIsOriginCall() {
		t.Errorf("expected IsOriginCall=false but got true")
	}
	tests.CallSubtestsAssertOriginCall()
}

func TestPrevRealm(t *testing.T) {
	var (
		firstRealm = std.DerivePkgAddr("gno.land/r/demo/tests_test")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/demo/tests")
	)
	// When only one realm in the frames, PrevRealm returns the same realm
	if addr := tests.GetPrevRealm().Addr(); addr != firstRealm {
		println(tests.GetPrevRealm())
		t.Errorf("want GetPrevRealm().Addr==%s, got %s", firstRealm, addr)
	}
	// When 2 or more realms in the frames, PrevRealm returns the second to last
	if addr := tests.GetRSubtestsPrevRealm().Addr(); addr != rTestsAddr {
		t.Errorf("want GetRSubtestsPrevRealm().Addr==%s, got %s", rTestsAddr, addr)
	}
}
