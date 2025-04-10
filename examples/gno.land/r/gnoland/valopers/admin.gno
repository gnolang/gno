package valopers

import (
	"chain/banker"

	"gno.land/p/moul/authz"
)

var auth *authz.Authorizer

func Auth() *authz.Authorizer {
	return auth
}

func updateInstructions(newInstructions string) {
	err := auth.Do("update-instructions", func() error {
		instructions = newInstructions
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func updateMinFee(newMinFee int64) {
	err := auth.Do("update-min-fee", func() error {
		minFee = banker.NewCoin("ugnot", newMinFee)
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func NewInstructionsProposalCallback(newInstructions string) func() error {
	cb := func() error {
		updateInstructions(newInstructions)
		return nil
	}

	return cb
}

func NewMinFeeProposalCallback(newMinFee int64) func() error {
	cb := func() error {
		updateMinFee(newMinFee)
		return nil
	}

	return cb
}
