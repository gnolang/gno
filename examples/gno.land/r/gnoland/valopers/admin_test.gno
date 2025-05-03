package valopers

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/moul/authz"
)

func TestUpdateInstructions(t *testing.T) {
	auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			"gno.land/r/gov/dao",
			func(title string, action authz.PrivilegedAction) error {
				return action()
			},
		),
	)

	newInstructions := "new instructions"

	uassert.AbortsWithMessage(t, "action can only be executed by the contract", func() {
		updateInstructions(newInstructions)
	})

	testing.SetOriginCaller(std.DerivePkgAddr("gno.land/r/gov/dao"))

	uassert.NotAborts(t, func() {
		updateInstructions(newInstructions)
	})

	uassert.Equal(t, newInstructions, instructions)
}

func TestUpdateMinFee(t *testing.T) {
	auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			"gno.land/r/gov/dao",
			func(title string, action authz.PrivilegedAction) error {
				return action()
			},
		),
	)

	newMinFee := int64(100)

	uassert.AbortsWithMessage(t, "action can only be executed by the contract", func() {
		updateMinFee(newMinFee)
	})

	testing.SetOriginCaller(std.DerivePkgAddr("gno.land/r/gov/dao"))

	uassert.NotAborts(t, func() {
		updateMinFee(newMinFee)
	})

	uassert.Equal(t, newMinFee, minFee.Amount)
}
