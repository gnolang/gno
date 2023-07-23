package foo20airdrop

import (
	"std"
	"testing"

	"gno.land/p/demo/airdrop"
	"gno.land/p/demo/merkle"
	"gno.land/r/demo/foo20"
	"gno.land/r/demo/users"
)

var leaves []merkle.Hashable = []airdrop.AirdropData{
	{
		Address: "g1auhc2cymv7gn9qmls0ttdr3wqrljgz0dhq90e",
		Amount:  1_000_000,
	},
	{
		Address: "g1zyvskpxg5lv4qpygtuvp93zprrrjpk2exa9rfx",
		Amount:  1_000_000,
	},
	{
		Address: "g14szvkruznx49sxe4m9dmg3m8606sm6yp4a0wv8",
		Amount:  1_000_000,
	},
}

// func TestRegisterMerkle(t *testing.T) {
// 	tree := merkle.NewTree(leaves)
// 	root := tree.Root()

// 	RegisterMerkleRoot(root)
// 	reset()
// }

func TestClaimAirdrop(t *testing.T) {
	// println("prev", std.PrevRealm())
	// println("current", std.CurrentRealm())
	// println("foo20-airdrop.PrevRealm", PrevRealm())

	contractAddr := std.DerivePkgAddr("gno.land/r/demo/foo20-airdrop")
	println("contractADDR", contractAddr)
	std.TestSetOrigCaller(contractAddr)

	// instantiate foo20 airdrop contract
	tree := merkle.NewTree(leaves)
	RegisterMerkleRoot(tree.Root())

	sumClaimed := uint64(0)
	for _, leaf := range leaves {
		data := leaf.(airdrop.AirdropData)
		user := data.Address
		sumClaimed += data.Amount

		proofs, err := tree.Proof(leaf)
		if err != nil {
			t.Fatalf("failed to generate proof, %v", err)
			return
		}

		// claim airdrop
		Claim(leaf.(airdrop.AirdropData), proofs)
	}

	ttClaimed := TotalClaimed()
	if ttClaimed != sumClaimed {
		t.Fatalf("expected: %d", sumClaimed)
	}
}
