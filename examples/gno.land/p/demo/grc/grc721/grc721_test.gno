package grc721

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

var (
	exampleNFTName   = "ExampleNFT"
	exampleNFTSymbol = "EXNFT"
	alice            = testutils.TestAddress("alice")
	bob              = testutils.TestAddress("bob")
)

func TestNewGRC721Token(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)
	if exampleNFT == nil {
		t.Errorf("should not be nil")
	}
}

func TestName(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)
	if exampleNFT == nil {
		t.Errorf("should not be nil")
	}
	name := exampleNFT.Name()
	if name != exampleNFTName {
		t.Errorf("expected: (%s), got: (%s)", exampleNFTName, name)
	}
}

func TestSymbol(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)
	symbol := exampleNFT.Symbol()
	if symbol != exampleNFTSymbol {
		t.Errorf("expected: (%s), got: (%s)", exampleNFTSymbol, symbol)
	}
}

func TestBalanceOf(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)

	balanceAddr1 := exampleNFT.BalanceOf(alice)

	if balanceAddr1 != 0 {
		t.Errorf("expected: (%d), got: (%d)", 0, balanceAddr1)
	}

	exampleNFT.Mint(alice, "1")
	exampleNFT.Mint(alice, "2")
	exampleNFT.Mint(bob, "3")

	balanceAddr1 = exampleNFT.BalanceOf(alice)
	balanceAddr2 := exampleNFT.BalanceOf(bob)

	if balanceAddr1 != 2 {
		t.Errorf("expected: (%d), got: (%d)", 2, balanceAddr1)
	}
	if balanceAddr2 != 1 {
		t.Errorf("expected: (%d), got: (%d)", 1, balanceAddr2)
	}
}

func TestOwnerOf(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)

	exampleNFT.Mint(alice, "1")
	exampleNFT.Mint(bob, "2")

	// Checking for token id "1"
	owner := exampleNFT.OwnerOf("1")

	if owner != alice {
		t.Errorf("expected: (%s), got: (%s)", alice.String(), owner.String())
	}

	// Checking for token id "2"
	owner = exampleNFT.OwnerOf("2")

	if owner != bob {
		t.Errorf("expected: (%s), got: (%s)", bob.String(), owner.String())
	}
}

func TestIsApprovedForAll(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)

	isApprovedForAll := exampleNFT.IsApprovedForAll(alice, bob)
	if isApprovedForAll != false {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}
}

func TestSetApprovalForAll(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)

	caller := std.PrevRealm().Addr()

	isApprovedForAll := exampleNFT.IsApprovedForAll(caller, alice)
	if isApprovedForAll != false {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}

	exampleNFT.SetApprovalForAll(alice, true)

	isApprovedForAll = exampleNFT.IsApprovedForAll(caller, alice)
	if isApprovedForAll != true {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}
}

func TestApprove(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)
	if exampleNFT == nil {
		t.Errorf("should not be nil")
	}

	caller := std.PrevRealm().Addr()

	exampleNFT.Mint(caller, "1")

	exampleNFT.Approve(alice, "1")
	approvedAddr := exampleNFT.GetApproved("1")

	if approvedAddr != alice {
		t.Errorf("expected: (%s), got: (%s)", alice.String(), approvedAddr.String())
	}
}

func TestTransferFrom(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)
	caller := std.PrevRealm().Addr()

	exampleNFT.Mint(caller, "1")
	exampleNFT.Mint(caller, "2")

	exampleNFT.TransferFrom(caller, alice, "1")

	// Check the balance of caller after transfer
	balanceOfCaller := exampleNFT.BalanceOf(caller)
	if balanceOfCaller != 1 {
		t.Errorf("expected: (%d), got: (%d)", 1, balanceOfCaller)
	}

	// Check balance of addr after transfer
	balanceOfAlice := exampleNFT.BalanceOf(alice)
	if balanceOfAlice != 1 {
		t.Errorf("expected: (%d), got: (%d)", 1, balanceOfAlice)
	}

	// Check Owner of transferred Token id
	owner := exampleNFT.OwnerOf("1")
	if owner != alice {
		t.Errorf("expected: (%s), got: (%s)", alice.String(), owner.String())
	}
}

func TestMint(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)

	exampleNFT.Mint(alice, "1")
	exampleNFT.Mint(alice, "2")
	exampleNFT.Mint(bob, "3")

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Mint should have panicked")
		}
	}()

	// Try Minting duplicate token id
	exampleNFT.Mint(bob, "1")

	// Check Owner of Token id
	owner := exampleNFT.OwnerOf("1")
	if owner != alice {
		t.Errorf("expected: (%s), got: (%s)", alice.String(), owner.String())
	}

	balance := exampleNFT.BalanceOf(alice)
	if balance != uint64(2) {
		t.Errorf("balances expected: (%d), got: (%d)", 2, balance)
	}

}

func TestBurn(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)

	exampleNFT.Mint(alice, "1")
	exampleNFT.Burn("1")

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("OwnerOf should have panicked after burn")
		}
	}()

	// Make code panic
	_ = exampleNFT.OwnerOf("1")
}

func TestSetTokenURI(t *testing.T) {
	exampleNFT := NewGRC721Token(exampleNFTName, exampleNFTSymbol)

	tokenURI := "https://example.com/token"
	exampleNFT.Mint(alice, "1")

	std.TestSetOrigCaller(alice)
	uri := exampleNFT.SetTokenURI("1", tokenURI)
	if uri != tokenURI {
		t.Fatalf("expected %s, got %s", tokenURI, uri)
	}
}
