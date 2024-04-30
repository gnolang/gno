package grc721

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/users"
)

var (
	dummyNFTName   = "DummyNFT"
	dummyNFTSymbol = "DNFT"
)

func TestNewBasicNFT(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}
}

func TestName(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}
	name := dummy.Name()
	if name != dummyNFTName {
		t.Errorf("expected: (%s), got: (%s)", dummyNFTName, name)
	}
}

func TestSymbol(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}
	symbol := dummy.Symbol()
	if symbol != dummyNFTSymbol {
		t.Errorf("expected: (%s), got: (%s)", dummyNFTSymbol, symbol)
	}
}

func TestTokenCount(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	count := dummy.TokenCount()
	if count != 0 {
		t.Errorf("expected: (%d), got: (%d)", 0, count)
	}

	dummy.mint("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", TokenID("1"))
	dummy.mint("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", TokenID("2"))

	count = dummy.TokenCount()
	if count != 2 {
		t.Errorf("expected: (%d), got: (%d)", 2, count)
	}
}

func TestBalanceOf(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	balanceAddr1, err := dummy.BalanceOf(addr1)
	if err != nil {
		t.Errorf("should not result in error")
	}
	if balanceAddr1 != 0 {
		t.Errorf("expected: (%d), got: (%d)", 0, balanceAddr1)
	}

	dummy.mint(addr1, TokenID("1"))
	dummy.mint(addr1, TokenID("2"))
	dummy.mint(addr2, TokenID("3"))

	balanceAddr1, err = dummy.BalanceOf(addr1)
	if err != nil {
		t.Errorf("should not result in error")
	}
	balanceAddr2, err := dummy.BalanceOf(addr2)
	if err != nil {
		t.Errorf("should not result in error")
	}

	if balanceAddr1 != 2 {
		t.Errorf("expected: (%d), got: (%d)", 2, balanceAddr1)
	}
	if balanceAddr2 != 1 {
		t.Errorf("expected: (%d), got: (%d)", 1, balanceAddr2)
	}
}

func TestOwnerOf(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	owner, err := dummy.OwnerOf(TokenID("invalid"))
	if err == nil {
		t.Errorf("should result in error")
	}

	dummy.mint(addr1, TokenID("1"))
	dummy.mint(addr2, TokenID("2"))

	// Checking for token id "1"
	owner, err = dummy.OwnerOf(TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}
	if owner != addr1 {
		t.Errorf("expected: (%s), got: (%s)", addr1.String(), owner.String())
	}

	// Checking for token id "2"
	owner, err = dummy.OwnerOf(TokenID("2"))
	if err != nil {
		t.Errorf("should not result in error")
	}
	if owner != addr2 {
		t.Errorf("expected: (%s), got: (%s)", addr2.String(), owner.String())
	}
}

func TestIsApprovedForAll(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	isApprovedForAll := dummy.IsApprovedForAll(addr1, addr2)
	if isApprovedForAll != false {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}
}

func TestSetApprovalForAll(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	caller := std.PrevRealm().Addr()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	isApprovedForAll := dummy.IsApprovedForAll(caller, addr)
	if isApprovedForAll != false {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}

	err := dummy.SetApprovalForAll(addr, true)
	if err != nil {
		t.Errorf("should not result in error")
	}

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	if isApprovedForAll != true {
		t.Errorf("expected: (%v), got: (%v)", false, isApprovedForAll)
	}
}

func TestGetApproved(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	approvedAddr, err := dummy.GetApproved(TokenID("invalid"))
	if err == nil {
		t.Errorf("should result in error")
	}
}

func TestApprove(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	caller := std.PrevRealm().Addr()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(caller, TokenID("1"))

	_, err := dummy.GetApproved(TokenID("1"))
	if err == nil {
		t.Errorf("should result in error")
	}

	err = dummy.Approve(addr, TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}

	approvedAddr, err := dummy.GetApproved(TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}
	if approvedAddr != addr {
		t.Errorf("expected: (%s), got: (%s)", addr.String(), approvedAddr.String())
	}
}

func TestTransferFrom(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	caller := std.PrevRealm().Addr()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(caller, TokenID("1"))
	dummy.mint(caller, TokenID("2"))

	err := dummy.TransferFrom(caller, addr, TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}

	// Check balance of caller after transfer
	balanceOfCaller, err := dummy.BalanceOf(caller)
	if err != nil {
		t.Errorf("should not result in error")
	}
	if balanceOfCaller != 1 {
		t.Errorf("expected: (%d), got: (%d)", 1, balanceOfCaller)
	}

	// Check balance of addr after transfer
	balanceOfAddr, err := dummy.BalanceOf(addr)
	if err != nil {
		t.Errorf("should not result in error")
	}
	if balanceOfAddr != 1 {
		t.Errorf("expected: (%d), got: (%d)", 1, balanceOfAddr)
	}

	// Check Owner of transferred Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}
	if owner != addr {
		t.Errorf("expected: (%s), got: (%s)", addr.String(), owner.String())
	}
}

func TestSafeTransferFrom(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	caller := std.PrevRealm().Addr()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(caller, TokenID("1"))
	dummy.mint(caller, TokenID("2"))

	err := dummy.SafeTransferFrom(caller, addr, TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}

	// Check balance of caller after transfer
	balanceOfCaller, err := dummy.BalanceOf(caller)
	if err != nil {
		t.Errorf("should not result in error")
	}
	if balanceOfCaller != 1 {
		t.Errorf("expected: (%d), got: (%d)", 1, balanceOfCaller)
	}

	// Check balance of addr after transfer
	balanceOfAddr, err := dummy.BalanceOf(addr)
	if err != nil {
		t.Errorf("should not result in error")
	}
	if balanceOfAddr != 1 {
		t.Errorf("expected: (%d), got: (%d)", 1, balanceOfAddr)
	}

	// Check Owner of transferred Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}
	if owner != addr {
		t.Errorf("expected: (%s), got: (%s)", addr.String(), owner.String())
	}
}

func TestMint(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	err := dummy.Mint(addr1, TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}
	err = dummy.Mint(addr1, TokenID("2"))
	if err != nil {
		t.Errorf("should not result in error")
	}
	err = dummy.Mint(addr2, TokenID("3"))
	if err != nil {
		t.Errorf("should not result in error")
	}

	// Try minting duplicate token id
	err = dummy.Mint(addr2, TokenID("1"))
	if err == nil {
		t.Errorf("should result in error")
	}

	// Check Owner of Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}
	if owner != addr1 {
		t.Errorf("expected: (%s), got: (%s)", addr1.String(), owner.String())
	}
}

func TestBurn(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(addr, TokenID("1"))
	dummy.mint(addr, TokenID("2"))

	err := dummy.Burn(TokenID("1"))
	if err != nil {
		t.Errorf("should not result in error")
	}

	// Check Owner of Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	if err == nil {
		t.Errorf("should result in error")
	}
}

func TestSetTokenURI(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	if dummy == nil {
		t.Errorf("should not be nil")
	}

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	tokenURI := "http://example.com/token"

	std.TestSetOrigCaller(std.Address(addr1)) // addr1

	dummy.mint(addr1, TokenID("1"))
	_, derr := dummy.SetTokenURI(TokenID("1"), TokenURI(tokenURI))

	if derr != nil {
		t.Errorf("Should not result in error ", derr.Error())
	}

	// Test case: Invalid token ID
	_, err := dummy.SetTokenURI(TokenID("3"), TokenURI(tokenURI))
	if err != ErrInvalidTokenId {
		t.Errorf("Expected error %v, got %v", ErrInvalidTokenId, err)
	}

	std.TestSetOrigCaller(std.Address(addr2)) // addr2

	_, cerr := dummy.SetTokenURI(TokenID("1"), TokenURI(tokenURI)) // addr2 trying to set URI for token 1
	if cerr != ErrCallerIsNotOwner {
		t.Errorf("Expected error %v, got %v", ErrCallerIsNotOwner, err)
	}

	// Test case: Retrieving TokenURI
	std.TestSetOrigCaller(std.Address(addr1)) // addr1

	dummyTokenURI, err := dummy.TokenURI(TokenID("1"))
	if err != nil {
		t.Errorf("TokenURI error: %v, ", err.Error())
	}
	if dummyTokenURI != tokenURI {
		t.Errorf("Expected URI %v, got %v", tokenURI, dummyTokenURI)
	}
}
