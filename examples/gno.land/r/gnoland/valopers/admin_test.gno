package valopers

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestUpdateInstructions(t *testing.T) {
	instructions := "new instructions"

	uassert.PanicsWithMessage(t, "action can only be executed by the contract", func() {
		UpdateInstructions(instructions)
	})

	std.TestSetOriginCaller(std.DerivePkgAddr("gno.land/r/gov/dao/bridge"))

	uassert.NotPanics(t, func() {
		UpdateInstructions(instructions)
	})
}
