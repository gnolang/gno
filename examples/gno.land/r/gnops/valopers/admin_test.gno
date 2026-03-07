package valopers

import (
	"chain"
	"testing"

	"gno.land/p/moul/authz"
	"gno.land/p/nt/uassert/v0"
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

	uassert.PanicsWithMessage(t, "action can only be executed by the contract", func() {
		updateInstructions(newInstructions)
	})

	testing.SetOriginCaller(chain.PackageAddress("gno.land/r/gov/dao"))

	uassert.NotPanics(t, func() {
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

	uassert.PanicsWithMessage(t, "action can only be executed by the contract", func() {
		updateMinFee(newMinFee)
	})

	testing.SetOriginCaller(chain.PackageAddress("gno.land/r/gov/dao"))

	uassert.NotPanics(t, func() {
		updateMinFee(newMinFee)
	})

	uassert.Equal(t, newMinFee, minFee.Amount)
}
