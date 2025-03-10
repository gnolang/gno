package valopers

import (
	"std"

	"gno.land/p/moul/authz"
)

var auth *authz.Authorizer

func Auth() *authz.Authorizer {
	return auth
}

func UpdateInstructions(newInstructions string) {
	err := auth.Do("update-instructions", func() error {
		instructions = newInstructions
		return nil
	})

	if err != nil {
		panic(err)
	}
}

func UpdateMinFee(newMinFee int64) {
	err := auth.Do("update-min-fee", func() error {
		minFee = newMinFee
		minCoin = std.NewCoin("ugnot", minFee)
		return nil
	})

	if err != nil {
		panic(err)
	}
}
