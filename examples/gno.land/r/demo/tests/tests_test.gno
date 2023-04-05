package tests

import (
	"testing"

	"gno.land/p/demo/testutils"
)

func TestAssertOriginCall(t *testing.T) {
	// No-panic case
	AssertOriginCall()

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
		AssertOriginCall()
	}()
}
