package valopers

import (
	"chain"
	"testing"

	"gno.land/p/moul/authz"
	"gno.land/p/nt/uassert/v0"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
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
