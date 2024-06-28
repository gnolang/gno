package tests

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestAssertOriginCall(t *testing.T) {
	// CallAssertOriginCall(): no panic
	CallAssertOriginCall()
	if !CallIsOriginCall() {
		t.Errorf("expected IsOriginCall=true but got false")
	}

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
		if CallIsOriginCall() {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		CallAssertOriginCall()
	}()

	// CallSubtestsAssertOriginCall(): panic
	defer func() {
		r := recover()
		if r == nil || r.(string) != expectedReason {
			t.Errorf("expected panic with '%v', got '%v'", expectedReason, r)
		}
	}()
	if CallSubtestsIsOriginCall() {
		t.Errorf("expected IsOriginCall=false but got true")
	}
	CallSubtestsAssertOriginCall()
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

func TestSetRealm_UserRealm(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetRealm(std.NewUserRealm(alice))

	got := GetPrevRealm().Addr()

	if alice != got {
		t.Fatalf("expected %s, got %s", alice, got)
	}
}

func TestSetRealm_CodeRealm(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/users"))
	expected := "g17m4ga9t9dxn8uf06p3cahdavzfexe33ecg8v2s" // derived from the path above

	got := GetPrevRealm().Addr()

	if expected != got.String() {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
