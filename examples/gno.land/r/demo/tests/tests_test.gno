package tests

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestAssertOriginCall(t *testing.T) {
	// Call method package, shouldn't panic
	AssertOriginCall()
	if !IsOriginCall() {
		t.Errorf("expected IsOriginCall=true but got false")
	}

	// Call other package method, should panic
	defer func() {
		expectedReason := "invalid non-origin call"
		r := recover()
		if r == nil || r.(string) != expectedReason {
			t.Errorf("expected panic with '%v', got '%v'", expectedReason, r)
		}
	}()
	if RSubtestsIsOriginCall() {
		t.Errorf("expected RSubtestsIsOriginCall=false but got true")
	}
	RSubtestsAssertOriginCall()
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
