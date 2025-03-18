package valopers

import (
	"std"

	"gno.land/p/moul/authz"
	"gno.land/r/gov/dao"
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
		minFee = std.NewCoin("ugnot", newMinFee)
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func NewInstructionsProposalCallback(newInstructions string) func(dao.Metadata) error {
	cb := func(dao.Metadata) error {
		updateInstructions(newInstructions)
		return nil
	}

	return cb
}

func NewMinFeeProposalCallback(newMinFee int64) func(dao.Metadata) error {
	cb := func(dao.Metadata) error {
		updateMinFee(newMinFee)
		return nil
	}

	return cb
}
