package valopers

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/moul/authz"
	"gno.land/r/gov/dao/bridge"
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

func NewInstructionsExecutor(newInstructions string) dao.Executor {
	cb := func() error {
		updateInstructions(newInstructions)
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}

func NewMinFeeExecutor(newMinFee int64) dao.Executor {
	cb := func() error {
		updateMinFee(newMinFee)
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}
