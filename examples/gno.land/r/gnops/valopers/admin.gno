package valopers

import (
	"gno.land/p/moul/authz"
)

var auth *authz.Authorizer

func Auth() *authz.Authorizer {
	return auth
}

func updateInstructions(_ int, rlm realm, newInstructions string) {
	err := auth.DoByCurrent(0, rlm, "update-instructions", func() error {
		instructions = newInstructions
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func NewInstructionsProposalCallback(newInstructions string) func(realm) error {
	cb := func(cur realm) error {
		updateInstructions(0, cur, newInstructions)
		return nil
	}

	return cb
}

// The min-fee callback was removed: the fee now lives in sysparams
// under node:valoper:register_fee, and
// proposal.ProposeNewMinFeeProposalRequest delegates to
// sys/params.NewSysParamUint64PropRequest. Removing avoids the
// forward-compat hazard of a no-op shim with no caller-auth gating.
