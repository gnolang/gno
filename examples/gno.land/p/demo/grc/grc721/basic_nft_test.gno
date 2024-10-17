package grc721

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

var (
	dummyNFTName   = "DummyNFT"
	dummyNFTSymbol = "DNFT"
)

func TestNewBasicNFT(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")
}

func TestName(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	name := dummy.Name()
	uassert.Equal(t, dummyNFTName, name)
}

func TestSymbol(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	symbol := dummy.Symbol()
	uassert.Equal(t, dummyNFTSymbol, symbol)
}

func TestTokenCount(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	count := dummy.TokenCount()
	uassert.Equal(t, uint64(0), count)

	dummy.mint("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", TokenID("1"))
	dummy.mint("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", TokenID("2"))

	count = dummy.TokenCount()
	uassert.Equal(t, uint64(2), count)
}

func TestBalanceOf(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	balanceAddr1, err := dummy.BalanceOf(addr1)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, uint64(0), balanceAddr1)

	dummy.mint(addr1, TokenID("1"))
	dummy.mint(addr1, TokenID("2"))
	dummy.mint(addr2, TokenID("3"))

	balanceAddr1, err = dummy.BalanceOf(addr1)
	uassert.NoError(t, err, "should not result in error")

	balanceAddr2, err := dummy.BalanceOf(addr2)
	uassert.NoError(t, err, "should not result in error")

	uassert.Equal(t, uint64(2), balanceAddr1)
	uassert.Equal(t, uint64(1), balanceAddr2)
}

func TestOwnerOf(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	owner, err := dummy.OwnerOf(TokenID("invalid"))
	uassert.Error(t, err, "should not result in error")

	dummy.mint(addr1, TokenID("1"))
	dummy.mint(addr2, TokenID("2"))

	// Checking for token id "1"
	owner, err = dummy.OwnerOf(TokenID("1"))
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, addr1.String(), owner.String())

	// Checking for token id "2"
	owner, err = dummy.OwnerOf(TokenID("2"))
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, addr2.String(), owner.String())
}

func TestIsApprovedForAll(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	isApprovedForAll := dummy.IsApprovedForAll(addr1, addr2)
	uassert.False(t, isApprovedForAll)
}

func TestSetApprovalForAll(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := std.PrevRealm().Addr()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	isApprovedForAll := dummy.IsApprovedForAll(caller, addr)
	uassert.False(t, isApprovedForAll)

	err := dummy.SetApprovalForAll(addr, true)
	uassert.NoError(t, err, "should not result in error")

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	uassert.True(t, isApprovedForAll)
}

func TestGetApproved(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	approvedAddr, err := dummy.GetApproved(TokenID("invalid"))
	uassert.Error(t, err, "should result in error")
}

func TestApprove(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := std.PrevRealm().Addr()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(caller, TokenID("1"))

	_, err := dummy.GetApproved(TokenID("1"))
	uassert.Error(t, err, "should result in error")

	err = dummy.Approve(addr, TokenID("1"))
	uassert.NoError(t, err, "should not result in error")

	approvedAddr, err := dummy.GetApproved(TokenID("1"))
	uassert.NoError(t, err, "should result in error")
	uassert.Equal(t, addr.String(), approvedAddr.String())
}

func TestTransferFrom(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := std.PrevRealm().Addr()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(caller, TokenID("1"))
	dummy.mint(caller, TokenID("2"))

	err := dummy.TransferFrom(caller, addr, TokenID("1"))
	uassert.NoError(t, err, "should result in error")

	// Check balance of caller after transfer
	balanceOfCaller, err := dummy.BalanceOf(caller)
	uassert.NoError(t, err, "should result in error")
	uassert.Equal(t, uint64(1), balanceOfCaller)

	// Check balance of addr after transfer
	balanceOfAddr, err := dummy.BalanceOf(addr)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, uint64(1), balanceOfAddr)

	// Check Owner of transferred Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	uassert.NoError(t, err, "should result in error")
	uassert.Equal(t, addr.String(), owner.String())
}

func TestSafeTransferFrom(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := std.PrevRealm().Addr()
	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(caller, TokenID("1"))
	dummy.mint(caller, TokenID("2"))

	err := dummy.SafeTransferFrom(caller, addr, TokenID("1"))
	uassert.NoError(t, err, "should not result in error")

	// Check balance of caller after transfer
	balanceOfCaller, err := dummy.BalanceOf(caller)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, uint64(1), balanceOfCaller)

	// Check balance of addr after transfer
	balanceOfAddr, err := dummy.BalanceOf(addr)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, uint64(1), balanceOfAddr)

	// Check Owner of transferred Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, addr.String(), owner.String())
}

func TestMint(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	err := dummy.Mint(addr1, TokenID("1"))
	uassert.NoError(t, err, "should not result in error")
	err = dummy.Mint(addr1, TokenID("2"))
	uassert.NoError(t, err, "should not result in error")
	err = dummy.Mint(addr2, TokenID("3"))
	uassert.NoError(t, err, "should not result in error")

	// Try minting duplicate token id
	err = dummy.Mint(addr2, TokenID("1"))
	uassert.Error(t, err, "should not result in error")

	// Check Owner of Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, addr1.String(), owner.String())
}

func TestBurn(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(addr, TokenID("1"))
	dummy.mint(addr, TokenID("2"))

	err := dummy.Burn(TokenID("1"))
	uassert.NoError(t, err, "should not result in error")

	// Check Owner of Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	uassert.Error(t, err, "should result in error")
}

func TestSetTokenURI(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := std.Address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	tokenURI := "http://example.com/token"

	std.TestSetOrigCaller(std.Address(addr1)) // addr1

	dummy.mint(addr1, TokenID("1"))
	_, derr := dummy.SetTokenURI(TokenID("1"), TokenURI(tokenURI))
	uassert.NoError(t, derr, "should not result in error")

	// Test case: Invalid token ID
	_, err := dummy.SetTokenURI(TokenID("3"), TokenURI(tokenURI))
	uassert.ErrorIs(t, err, ErrInvalidTokenId)

	std.TestSetOrigCaller(std.Address(addr2)) // addr2

	_, cerr := dummy.SetTokenURI(TokenID("1"), TokenURI(tokenURI)) // addr2 trying to set URI for token 1
	uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

	// Test case: Retrieving TokenURI
	std.TestSetOrigCaller(std.Address(addr1)) // addr1

	dummyTokenURI, err := dummy.TokenURI(TokenID("1"))
	uassert.NoError(t, err, "TokenURI error")
	uassert.Equal(t, string(tokenURI), string(dummyTokenURI))
}
