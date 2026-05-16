package valopers

import (
	"chain"
	"testing"

	"gno.land/p/moul/authz"
	"gno.land/p/nt/uassert/v0"
)

func TestUpdateInstructions(cur realm, t *testing.T) {
	auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			"gno.land/r/gov/dao",
			func(title string, action authz.PrivilegedAction) error {
				return action()
			},
		),
	)

	newInstructions := "new instructions"

	uassert.PanicsWithMessage(t, cur, "action can only be executed by the contract", func() {
		updateInstructions(0, cur, newInstructions)
	})

	testing.SetOriginCaller(chain.PackageAddress("gno.land/r/gov/dao"))

	uassert.NotPanics(t, cur, func() {
		updateInstructions(0, cur, newInstructions)
	})

	uassert.Equal(t, newInstructions, instructions)
}

// TestUpdateMinFee was removed: register_fee now lives in sysparams
// (node:valoper:register_fee), governed via the generic
// NewSysParamUint64PropRequest factory. See proposal.gno for the
// replacement flow.
