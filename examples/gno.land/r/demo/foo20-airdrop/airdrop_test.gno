package foo20airdrop

import (
	"std"
	"testing"

	"gno.land/p/demo/airdrop/grc20_merkle_airdrop"
	"gno.land/p/demo/merkle"
	"gno.land/p/demo/urequire"
	"gno.land/p/demo/users"
	"gno.land/r/demo/foo20"
)

func getLeaves() []merkle.Hashable {
	var leaves = []grc20_merkle_airdrop.AirdropData{
		{
			Address: "g1sw5xklxjjuv0yvuxy5f5s3l3mnj0nqq626a9wr", // albttx.gno
			Amount:  10000,
		},
		{
			Address: "g1zyvskpxg5lv4qpygtuvp93zprrrjpk2exa9rfx",
			Amount:  10000,
		},
		{
			Address: "g14szvkruznx49sxe4m9dmg3m8606sm6yp4a0wv8",
			Amount:  10000,
		},
	}

	resp := make([]merkle.Hashable, len(leaves))
	for i, leave := range leaves {
		resp[i] = leave
	}
	return resp
}

func TestRegisterMerkle(t *testing.T) {
	leaves := getLeaves()

	tree := merkle.NewTree(leaves)
	root := tree.Root()

	RegisterMerkleRoot(root)
}

func TestClaimAirdrop(t *testing.T) {
	leaves := getLeaves()

	contractAddr := std.DerivePkgAddr("gno.land/r/demo/foo20-airdrop")
	std.TestSetOrigCaller(contractAddr)

	for _, leave := range leaves {
		// Get 10k token
		foo20.Faucet()
	}

	// instantiate foo20 airdrop contract
	tree := merkle.NewTree(leaves)
	RegisterMerkleRoot(tree.Root())

	sumClaimed := uint64(0)
	for _, leaf := range leaves {
		data := leaf.(grc20_merkle_airdrop.AirdropData)
		user := data.Address
		sumClaimed += data.Amount

		proofs, err := tree.Proof(leaf)
		urequire.NoError(t, err)

		// claim airdrop
		Claim(data, proofs)

		balance := foo20.BalanceOf(users.AddressOrName(data.Address))
		urequire.Equal(t, balance, uint64(10000))
	}

	ttClaimed := TotalClaimed()
	urequire.Equal(t, ttClaimed, sumClaimed)
}
