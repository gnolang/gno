package grc20_merkle_airdrop

import (
	"std"
	"testing"

	"gno.land/p/demo/airdrop/grc20_merkle_airdrop"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/merkle"
	"gno.land/p/demo/urequire"
	"gno.land/r/demo/foo20"
	"gno.land/r/demo/users"
)

func getLeaves() []merkle.Hashable {
	var leaves = []AirdropData{
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
	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")

	token := grc20.NewBanker("TOKEN", "TOK", 6)
	token.Mint(contractAddr, 50000) // Airdrop contract

	tok20airdrop := NewMerkleAirdrop(root, token.Token())
	_ = tok20airdrop
}

func TestClaimAirdrop(t *testing.T) {
	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")
	std.TestSetOrigCaller(contractAddr)

	leaves := getLeaves()
	tree := merkle.NewTree(leaves)
	root := tree.Root()

	// instantiate foo20 airdrop contract
	token := grc20.NewBanker("TOKEN", "TOK", 6)
	token.Mint(contractAddr, 50000) // Airdrop contract

	tok20airdrop := NewMerkleAirdrop(root, token.Token())

	sumClaimed := uint64(0)
	for _, leaf := range leaves {
		data := leaf.(AirdropData)
		println(data)

		sumClaimed += data.Amount

		proofs, err := tree.Proof(leaf)
		urequire.NoError(t, err)

		// claim airdrop
		err = tok20airdrop.Claim(data, proofs)
		urequire.NoError(t, err)
	}

	ttClaimed := tok20airdrop.TotalClaimed()
	urequire.Equal(t, ttClaimed, sumClaimed)
}

func TestDoubleClaim(t *testing.T) {
	leaves := getLeaves()

	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")
	std.TestSetOrigCaller(contractAddr)

	tree := merkle.NewTree(leaves)
	token := grc20.NewBanker("TOKEN", "TOK", 6)
	token.Mint(contractAddr, 50000)

	tok20airdrop := NewMerkleAirdrop(tree.Root(), token.Token())

	leaf := leaves[0]
	proofs, err := tree.Proof(leaf)
	urequire.NoError(t, err)

	err = tok20airdrop.Claim(leaf.(AirdropData), proofs)
	urequire.NoError(t, err)

	err = tok20airdrop.Claim(leaf.(AirdropData), proofs)
	urequire.Error(t, err, ErrAlreadyClaimed.Error())
}
