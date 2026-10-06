package vm_test

import (
	"chain"
	"fmt"
	"testing"

	"gno.land/p/nt/testutils/v0"
	tests "gno.land/r/tests/vm"
)

func TestAssertOriginCall(cur realm, t *testing.T) {
	// CallAssertOriginCall(): no panic
	caller := testutils.TestAddress("caller")
	testing.SetRealm(testing.NewUserRealm(caller))
	tests.CallAssertOriginCall(cross(cur))
	if !tests.CallIsOriginCall(cross(cur)) {
		t.Errorf("expected IsOriginCall=true but got false")
	}

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/tests/vm"))
	// CallAssertOriginCall() from a block: abort
	r := revive(func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal block).
		if tests.CallIsOriginCall(cross(cur)) {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		tests.CallAssertOriginCall(cross(cur)) // <---
	})
	if fmt.Sprintf("%v", r) != "invalid non-origin call" {
		t.Error("expected abort but did not")
	}
	// CallSubtestsAssertOriginCall(): abort
	r = revive(func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal block).
		if tests.CallSubtestsIsOriginCall(cross(cur)) {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		tests.CallSubtestsAssertOriginCall(cross(cur))
	})
	if fmt.Sprintf("%v", r) != "invalid non-origin call" {
		t.Error("expected abort but did not")
	}
}

func TestPreviousRealm(cur realm, t *testing.T) {
	var (
		firstRealm = chain.PackageAddress("gno.land/r/tests/vm_test")
		rTestsAddr = chain.PackageAddress("gno.land/r/tests/vm")
	)
	// When only one realm in the frames, PreviousRealm returns the same realm
	if addr := tests.GetPreviousRealm(cross(cur)).Address(); addr != firstRealm {
		println(tests.GetPreviousRealm(cross(cur)))
		t.Errorf("want GetPreviousRealm().Address==%s, got %s", firstRealm, addr)
	}
	// When 2 or more realms in the frames, PreviousRealm returns the second to last
	if addr := tests.GetRSubtestsPreviousRealm(cross(cur)).Address(); addr != rTestsAddr {
		t.Errorf("want GetRSubtestsPreviousRealm().Address==%s, got %s", rTestsAddr, addr)
	}
}
