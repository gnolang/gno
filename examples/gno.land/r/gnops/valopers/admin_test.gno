package valopers

import (
	"testing"

	"gno.land/p/moul/authz"
	"gno.land/p/nt/uassert/v0"
)

// Earlier this test also asserted a panic ("action can only be
// executed by the contract") when called without SetOriginCaller.
// That panic was emitted by a `runtime.CurrentRealm() == contractAddr`
// check inside authz.ContractAuthority.Authorize, which has been
// removed — it was .Title()-bypassable (runtime.CurrentRealm inside a
// non-crossing closure walks past non-crossing frames to the
// most-recent crossing ancestor, so a malicious caller could shape
// the call chain to make the check resolve to contractAddr).
//
// Caller authentication is now upstream of the deleted gate:
// authz.Authorizer.DoByCurrent requires `rlm.IsCurrent()`, and the
// contract handler closure (registered at init in this realm) is the
// remaining Class-4 trust root by lexical capture. There is no
// negative path here to assert, so the prior expectation was retired
// along with its source.
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

	uassert.NotPanics(t, cur, func() {
		updateInstructions(0, cur, newInstructions)
	})

	uassert.Equal(t, newInstructions, instructions)
}

// TestUpdateMinFee was removed: register_fee now lives in sysparams
// (node:valoper:register_fee), governed via the generic
// NewSysParamUint64PropRequest factory. See proposal.gno for the
// replacement flow.
