package tests

import (
	"testing"

	"gno.land/p/demo/testutils"
)

func TestAssertOriginCall(t *testing.T) {
	AssertOriginCall()
	testutils.RequirePanic(t, func() {
		// if called inside a function literal, this is no longer an origin call
		// because there's one additional frame (the function literal).
		AssertOriginCall()
	}, "invalid non-origin call")
}
