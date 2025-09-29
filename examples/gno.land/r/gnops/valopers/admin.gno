package valopers

import (
	"chain"

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

func updateMinFee(newMinFee int64) {
	err := auth.DoByCurrent("update-min-fee", func() error {
		minFee = chain.NewCoin("ugnot", newMinFee)
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

func NewMinFeeProposalCallback(newMinFee int64) func(realm) error {
	cb := func(cur realm) error {
		updateMinFee(newMinFee)
		return nil
	}

	return cb
}
