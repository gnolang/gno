package vm_test

import (
	"chain"
	"fmt"
	"testing"

	"gno.land/p/nt/testutils/v0"
	tests "gno.land/r/tests/vm"
)

func TestAssertOriginCall(cur realm, t *testing.T) {
	// Pretend the caller is a user: the message itself, nothing in between.
	caller := testutils.TestAddress("caller")
	testing.SetRealm(testing.NewUserRealm(caller))
	tests.CallAssertOriginCall(cross(cur))
	if !tests.CallIsOriginCall(cross(cur)) {
		t.Errorf("expected IsOriginCall=true but got false")
	}

	// From a function literal in the caller's own code: still an origin
	// call. Only another realm on the stack changes the answer.
	r := revive(func() {
		if !tests.CallIsOriginCall(cross(cur)) {
			t.Errorf("expected IsOriginCall=true but got false")
		}
		tests.CallAssertOriginCall(cross(cur))
	})
	if r != nil {
		t.Errorf("expected no abort but got %v", r)
	}
	// Pretend the caller is realm subtests. A realm cannot originate a
	// message, so on chain this shape is user -> subtests -> r/tests/vm,
	// with subtests in between: abort.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/tests/vm/subtests"))
	r = revive(func() {
		if tests.CallIsOriginCall(cross(cur)) {
			t.Errorf("expected IsOriginCall=false but got true")
		}
		tests.CallAssertOriginCall(cross(cur))
	})
	if fmt.Sprintf("%v", r) != "invalid non-origin call" {
		t.Error("expected abort but did not")
	}
	testing.SetRealm(testing.NewUserRealm(caller))
	// Through a second realm (r/tests/vm -> r/tests/vm/subtests): abort.
	r = revive(func() {
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
