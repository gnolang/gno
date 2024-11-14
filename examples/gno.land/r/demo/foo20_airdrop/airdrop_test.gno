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
			std.DerivePkgAddr(ufmt.Sprintf("gno.land/r/test/%d", i)),
			10000,
		)
	}

	return leaves
}

func TestRegisterMerkle(t *testing.T) {
	defer reset()
	leaves := getLeaves(5)

	tree := merkle.NewTree(leaves)
	root := tree.Root()

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
	for i, leaf := range leaves {
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

func TestClaimJSONAirdrop(t *testing.T) {
	defer reset()
	tests := []struct {
		name string
		root string
		in   string
	}{
		{
			name: "10 leaves-leaf5",
			root: "0cce01ece896de259789bbbef8ce0481c6dbde1fa67dbe08089d66d26ec8e9a5",
			in: `{
  "address": "g1tvhw8hqz272fglt8p5efwj226r3uqf064u3wxk",
  "amount": "10000",
  "proof": [
    {
      "data": "942abf9387714a4a9c3348a7803d22f8964e1eb8fba0b74fdd20b4e67aa07ffe",
      "position": 1
    },
    {
      "data": "d4aaaeb914c8367099ad24975c70a2a3d5af91ab3decbe7033496fbf76fdc176",
      "position": 0
    },
    {
      "data": "9f52c2dda050e932ca109d2a5a025acd73ed258c5d2b1b51dfbda3896bc60a0a",
      "position": 1
    },
    {
      "data": "8cb6b56e7db744b79d7f0da4532404c65c5332f2b44b58fa006fe5107848a110",
      "position": 0
    }
  ]
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterMerkleRoot(tt.root)

			ClaimJSON(tt.in)
		})
	}
}
