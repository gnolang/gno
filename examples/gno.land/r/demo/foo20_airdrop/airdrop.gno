package foo20_airdrop

import (
	"gno.land/p/demo/airdrop"
	"gno.land/p/demo/airdrop/grc20_merkle_airdrop"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/merkle"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/foo20"
)

var (
	token grc20.Token

	foo20airdrop *grc20_merkle_airdrop.MerkleAirdrop
)

func RegisterMerkleRoot(root string) {
	token = foo20.Token()

	if foo20airdrop != nil {
		panic("foo20 airdrop merkle root is already registered")
	}
	foo20airdrop = grc20_merkle_airdrop.NewMerkleAirdrop(root, token)
}

func Claim(data airdrop.AirdropData, proofs []merkle.Node) {
	err := foo20airdrop.Claim(data, proofs)
	if err != nil {
		panic(err.Error())
	}
}

func TotalClaimed() uint64 {
	return foo20airdrop.TotalClaimed()
}

func Render(path string) string {
	if foo20airdrop == nil {
		return "Airdrop is not registered yet"
	}
	return ufmt.Sprintf("total claimed: %d", foo20airdrop.TotalClaimed())
}
