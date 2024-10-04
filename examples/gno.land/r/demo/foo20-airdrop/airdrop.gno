package foo20airdrop

import (
	"std"

	"gno.land/p/demo/airdrop/grc20_merkle_airdrop"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/merkle"
	"gno.land/r/demo/foo20"
)

var (
	token grc20.Token

	// admin std.Address = "g1sw5xklxjjuv0yvuxy5f5s3l3mnj0nqq626a9wr" // albttx.gno

	foo20airdrop *grc20_merkle_airdrop.MerkleAirdrop
)

func RegisterMerkleRoot(root string) {
	token = foo20.Token()
	foo20airdrop = nil

	if foo20airdrop != nil {
		panic("foo20 airdrop merkle root is already registered")
	}
	foo20airdrop = grc20_merkle_airdrop.NewMerkleAirdrop(root, token)
}

func Claim(data grc20_merkle_airdrop.AirdropData, proofs []merkle.Node) {
	err := foo20airdrop.Claim(data, proofs)
	if err != nil {
		panic(err.Error())
	}

	// err = token.Transfer(data.Address, data.Amount)
	// if err != nil {
	// 	panic(err.Error())
	// }
}

func TotalClaimed() uint64 {
	return foo20airdrop.TotalClaimed()
}
