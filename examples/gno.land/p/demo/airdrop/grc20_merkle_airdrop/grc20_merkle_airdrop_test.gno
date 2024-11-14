package grc20_merkle_airdrop

import (
	"std"
	"testing"

	"gno.land/p/demo/airdrop"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/merkle"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func getLeaves(size int) []merkle.Hashable {
	leaves := make([]merkle.Hashable, size)

	for i := 0; i < size; i++ {
		leaves[i] = airdrop.NewData(
			std.DerivePkgAddr(ufmt.Sprintf("gno.land/test/%d", i)),
			10000,
		)
	}

	return leaves
}

func TestRegisterMerkle(t *testing.T) {
	leaves := getLeaves(3)
	tree := merkle.NewTree(leaves)
	root := tree.Root()
	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")

	token, admin := grc20.NewToken("TOKEN", "TOK", 6)
	admin.Mint(contractAddr, 50000) // Airdrop contract

	tok20airdrop := NewMerkleAirdrop(root, token.RealmTeller())
	_ = tok20airdrop
}

func TestClaimAirdrop(t *testing.T) {
	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")
	std.TestSetOrigCaller(contractAddr)

	leaves := getLeaves(5)
	tree := merkle.NewTree(leaves)
	root := tree.Root()

	// instantiate foo20 airdrop contract
	token, admin := grc20.NewToken("TOKEN", "TOK", 6)
	admin.Mint(contractAddr, 50000) // Airdrop contract

	tok20airdrop := NewMerkleAirdrop(root, token.RealmTeller())

	sumClaimed := uint64(0)
	for _, leaf := range leaves {
		data := leaf.(airdrop.Data)

		sumClaimed += data.Amount()

		proofs, err := tree.Proof(leaf)
		urequire.NoError(t, err)

		// claim airdrop
		err = tok20airdrop.Claim(data, proofs)
		urequire.NoError(t, err)

		balance := token.BalanceOf(data.Address())
		urequire.Equal(t, balance, uint64(10000))
	}

	ttClaimed := tok20airdrop.TotalClaimed()
	urequire.Equal(t, ttClaimed, sumClaimed)
}

func TestDoubleClaim(t *testing.T) {
	leaves := getLeaves(5)

	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")
	std.TestSetOrigCaller(contractAddr)

	tree := merkle.NewTree(leaves)
	token, admin := grc20.NewToken("TOKEN", "TOK", 6)
	admin.Mint(contractAddr, 50000)

	tok20airdrop := NewMerkleAirdrop(tree.Root(), token.RealmTeller())

	leaf := leaves[0]
	proofs, err := tree.Proof(leaf)
	urequire.NoError(t, err)

	err = tok20airdrop.Claim(leaf.(airdrop.Data), proofs)
	urequire.NoError(t, err)

	err = tok20airdrop.Claim(leaf.(airdrop.Data), proofs)
	urequire.Error(t, err, ErrAlreadyClaimed.Error())
}
