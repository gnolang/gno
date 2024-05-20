package tests

import (
	"std"
	"testing"
)

func TestAssertOriginCall(t *testing.T) {
	// No-panic case
	AssertOriginCall()
	if !IsOriginCall() {
		t.Errorf("expected IsOriginCall=true but got false")
	}

	// Panic case
	expectedReason := "invalid non-origin call"
	defer func() {
		r := recover()
		if r == nil || r.(string) != expectedReason {
			t.Errorf("expected panic with '%v', got '%v'", expectedReason, r)
		}
	}()
	func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal).
		if IsOriginCall() {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		AssertOriginCall()
	}()
}

func TestPrevRealm(t *testing.T) {
	var (
		user1Addr  = std.DerivePkgAddr("user1.gno")
		rTestsAddr = std.DerivePkgAddr("gno.land/r/demo/tests")
	)
	// When a single realm in the frames, PrevRealm returns the user
	if addr := GetPrevRealm().Addr(); addr != user1Addr {
		t.Errorf("want GetPrevRealm().Addr==%s, got %s", user1Addr, addr)
	}
	// When 2 or more realms in the frames, PrevRealm returns the second to last
	if addr := GetRSubtestsPrevRealm().Addr(); addr != rTestsAddr {
		t.Errorf("want GetRSubtestsPrevRealm().Addr==%s, got %s", rTestsAddr, addr)
	}
}
