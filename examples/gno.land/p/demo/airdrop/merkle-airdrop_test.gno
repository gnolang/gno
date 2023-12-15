package airdrop

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/merkle"
	"gno.land/r/demo/foo20"
	"gno.land/r/demo/users"
)

var leaves []merkle.Hashable = []AirdropData{
	{
		Address: "g1auhc2cymv7gn9qmls0ttdr3wqrljgz0dhq90e",
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

func TestRegisterMerkle(t *testing.T) {
	tree := merkle.NewTree(leaves)
	root := tree.Root()
	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")

	token := grc20.NewAdminToken("TOKEN", "TOK", 6)
	token.Mint(contractAddr, 50000) // Airdrop contract

	tok20airdrop := NewMerkleAirdrop(root, token.GRC20())
}

func TestClaimAirdrop(t *testing.T) {
	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")
	std.TestSetOrigCaller(contractAddr)

	// instantiate foo20 airdrop contract
	tree := merkle.NewTree(leaves)
	root := tree.Root()

	token := grc20.NewAdminToken("TOKEN", "TOK", 6)
	token.Mint(contractAddr, 50000) // Airdrop contract

	tok20airdrop := NewMerkleAirdrop(root, token.GRC20())

	sumClaimed := uint64(0)
	for _, leaf := range leaves {
		data := leaf.(AirdropData)
		user := data.Address
		sumClaimed += data.Amount

		proofs, err := tree.Proof(leaf)
		if err != nil {
			t.Fatalf("failed to generate proof, %v", err)
			return
		}

		// claim airdrop
		tok20airdrop.Claim(data, proofs)
	}

	ttClaimed := tok20airdrop.TotalClaimed()
	if ttClaimed != sumClaimed {
		t.Fatalf("expected: %d, got: %d", sumClaimed, ttClaimed)
	}
}

func TestDoubleClaim(t *testing.T) {
	contractAddr := std.DerivePkgAddr("gno.land/r/demo/tok20-airdrop")
	std.TestSetOrigCaller(contractAddr)

	tree := merkle.NewTree(leaves)
	token := grc20.NewAdminToken("TOKEN", "TOK", 6)
	token.Mint(contractAddr, 50000)

	tok20airdrop := NewMerkleAirdrop(tree.Root(), token.GRC20())

	leaf := leaves[0]
	proofs, err := tree.Proof(leaf)
	if err != nil {
		t.Fatalf("failed to generate proof, %v", err)
		return
	}

	err = tok20airdrop.Claim(leaf.(AirdropData), proofs)
	if err != nil {
		t.Fatalf("failed to claim airdrop: %v", err)
	}

	err = tok20airdrop.Claim(leaf.(AirdropData), proofs)
	if err != ErrAlreadyClaimed {
		t.Fatalf("want: %v, got: %v", ErrAlreadyClaimed, err)
	}
}
