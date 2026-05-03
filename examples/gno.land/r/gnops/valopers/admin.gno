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

// NewMinFeeProposalCallback is preserved for historical-replay
// compatibility. The fee now lives in sysparams (node:valoper:register_fee);
// on each invocation the callback no-ops the historical update because
// proposal.ProposeNewMinFeeProposalRequest returns a generic
// NewSysParamUint64PropRequest. Kept as a zero-effect shim in case any
// historical tx calls it directly (none observed in the gnoland-1 archive).
func NewMinFeeProposalCallback(newMinFee int64) func(realm) error {
	cb := func(cur realm) error {
		// no-op: governance now updates register_fee via sys/params.
		return nil
	}
	return cb
}
