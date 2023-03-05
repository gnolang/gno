package origincall

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestOriginCall(t *testing.T) {
	requirePanic(t, func() {
		std.AssertOriginCall()
	}, "invalid non-origin call")
	// NOTE(tb): it's deceptively not possible to have a test for AssertOriginCall
	// and an other test for IsOriginCall because the context is shared between
	// tests (e.g. if std.TestSetOrigCaller is called in a test, this affects also
	// the following test).
	if b := std.IsOriginCall(); b {
		t.Errorf("expect isOriginCall false, got true")
	}

	std.TestSetOrigCaller(testutils.TestAddress("user"))

	std.AssertOriginCall()
	if b := std.IsOriginCall(); !b {
		t.Errorf("expect isOriginCall true, got false")
	}
}

func requirePanic(t *testing.T, f func(), expRecover string) {
	defer func() {
		r := recover()
		if r == nil || r.(string) != expRecover {
			t.Fatalf("expected panic with '%v', got '%v'", expRecover, r)
		}
	}()
	f()
}
