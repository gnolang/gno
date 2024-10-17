package foo20_airdrop

import (
	"std"
	"testing"

	"gno.land/p/demo/airdrop"
	"gno.land/p/demo/merkle"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
	"gno.land/p/demo/users"
	"gno.land/r/demo/foo20"
)

func reset() {
	foo20airdrop = nil
}

func getLeaves(size int) []merkle.Hashable {
	leaves := make([]merkle.Hashable, size)

	for i := 0; i < size; i++ {
		leaves[i] = airdrop.NewData(
			std.DerivePkgAddr(ufmt.Sprintf("gno.land/test/%d", i)),
			10000,
		)
	}
	leaves = append(leaves, airdrop.NewData(
		"g1sw5xklxjjuv0yvuxy5f5s3l3mnj0nqq626a9wr",
		1000,
	))

	return leaves
}

func TestRegisterMerkle(t *testing.T) {
	defer reset()
	leaves := getLeaves(5)

	tree := merkle.NewTree(leaves)
	root := tree.Root()

	println(root)

	RegisterMerkleRoot(root)
}

func TestClaimAirdrop(t *testing.T) {
	defer reset()
	leaves := getLeaves(5)

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
		data := leaf.(airdrop.Data)
		user := data.Address()
		sumClaimed += data.Amount()

		proofs, err := tree.Proof(leaf)
		urequire.NoError(t, err)

		// claim airdrop
		Claim(data, proofs)

		addr := users.AddressOrName(data.Address())
		balance := foo20.BalanceOf(addr)
		urequire.Equal(t, balance, uint64(10000))
	}

	urequire.Equal(t, TotalClaimed(), sumClaimed)
}
