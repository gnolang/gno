package valopers

import (
	"gno.land/p/moul/authz"
)

var auth *authz.Authorizer

func Auth() *authz.Authorizer {
	return auth
}

func updateInstructions(newInstructions string) {
	err := auth.DoByCurrent("update-instructions", func() error {
		instructions = newInstructions
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func NewInstructionsProposalCallback(newInstructions string) func(realm) error {
	cb := func(cur realm) error {
		updateInstructions(newInstructions)
		return nil
	}

	return cb
}

// NewMinFeeProposalCallback was deleted as part of VALOPLAN2: the fee
// now lives in sysparams (node:valoper:register_fee), and
// proposal.ProposeNewMinFeeProposalRequest delegates to
// sys/params.NewSysParamUint64PropRequest. The gnoland-1 tx archive
// contains zero historical callers of this function (only one
// historical reference exists, via proposal.ProposeNewMinFeeProposalRequest,
// which is rewired to the new path). Removing avoids forward-compat
// hazard of a no-op shim with no caller-auth gating.
