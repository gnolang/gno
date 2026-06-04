package grc721

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	dummyNFTName   = "DummyNFT"
	dummyNFTSymbol = "DNFT"
)

// The IIFE same-realm cross promotes the test's EOA-origin cur into a
// /p/grc721 CodeRealm cur so NewBasicNFT's IsCurrent + non-empty PkgPath
// checks pass.
func newDummyNFT(_ int, rlm realm) *BasicNFT {
	var nft *BasicNFT
	func(cur realm) {
		nft = NewBasicNFT(0, cur, dummyNFTName, dummyNFTSymbol)
	}(cross(rlm))
	return nft
}

func TestNewBasicNFTValidation(cur realm, t *testing.T) {
	const (
		nameErr   = "invalid token name"
		symbolErr = "invalid token symbol"
	)

	uassert.AbortsContains(t, cur, nameErr, func(cur realm) {
		NewBasicNFT(0, cur, "", "OK")
	}, "empty name")
	uassert.AbortsContains(t, cur, symbolErr, func(cur realm) {
		NewBasicNFT(0, cur, "Name", "")
	}, "empty symbol")
	uassert.AbortsContains(t, cur, nameErr, func(cur realm) {
		NewBasicNFT(0, cur, strings.Repeat("a", MaxNameLen+1), "OK")
	}, "name too long")
	uassert.AbortsContains(t, cur, symbolErr, func(cur realm) {
		NewBasicNFT(0, cur, "Name", strings.Repeat("A", MaxSymbolLen+1))
	}, "symbol too long")
	uassert.AbortsContains(t, cur, nameErr, func(cur realm) {
		NewBasicNFT(0, cur, "bad\x01name", "OK")
	}, "name with control char")
	uassert.AbortsContains(t, cur, symbolErr, func(cur realm) {
		NewBasicNFT(0, cur, "Name", "BA.D")
	}, "symbol with dot")
	uassert.AbortsContains(t, cur, symbolErr, func(cur realm) {
		NewBasicNFT(0, cur, "Name", "BA D")
	}, "symbol with space")

	uassert.NotAborts(t, cur, func(cur realm) {
		NewBasicNFT(0, cur, strings.Repeat("a", MaxNameLen), strings.Repeat("A", MaxSymbolLen))
	}, "boundary name+symbol should succeed")
}

func TestNewBasicNFT(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")
}

func TestName(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	name := dummy.Name()
	uassert.Equal(t, dummyNFTName, name)
}

func TestSymbol(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	symbol := dummy.Symbol()
	uassert.Equal(t, dummyNFTSymbol, symbol)
}

func TestTokenCount(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	count := dummy.TokenCount()
	uassert.Equal(t, int64(0), count)

	dummy.mint("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", TokenID("1"))
	dummy.mint("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", TokenID("2"))

	count = dummy.TokenCount()
	uassert.Equal(t, int64(2), count)
}

func TestBalanceOf(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	balanceAddr1, err := dummy.BalanceOf(addr1)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(0), balanceAddr1)

	dummy.mint(addr1, TokenID("1"))
	dummy.mint(addr1, TokenID("2"))
	dummy.mint(addr2, TokenID("3"))

	balanceAddr1, err = dummy.BalanceOf(addr1)
	uassert.NoError(t, err, "should not result in error")

	balanceAddr2, err := dummy.BalanceOf(addr2)
	uassert.NoError(t, err, "should not result in error")

	uassert.Equal(t, int64(2), balanceAddr1)
	uassert.Equal(t, int64(1), balanceAddr2)
}

func TestOwnerOf(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

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

func TestIsApprovedForAll(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	isApprovedForAll := dummy.IsApprovedForAll(addr1, addr2)
	uassert.False(t, isApprovedForAll)
}

func TestSetApprovalForAll(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := testutils.TestAddress("caller")
	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	// Test setting approval to true
	isApprovedForAll := dummy.IsApprovedForAll(caller, addr)
	uassert.False(t, isApprovedForAll)

	err := dummy.SetApprovalForAll(caller, addr, true)
	uassert.NoError(t, err, "should not result in error")

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	uassert.True(t, isApprovedForAll)

	// Test setting approval to false
	err = dummy.SetApprovalForAll(caller, addr, false)
	uassert.NoError(t, err, "should not result in error")

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	uassert.False(t, isApprovedForAll)
}

func TestGetApproved(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	_, err := dummy.GetApproved(TokenID("invalid"))
	uassert.Error(t, err, "should result in error")
}

func TestApprove(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")

	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	err := dummy.mint(alice, TokenID("1"))
	uassert.NoError(t, err, "failed to mint")

	_, err = dummy.GetApproved(TokenID("1"))
	uassert.Error(t, err, "should result in error")

	err = dummy.Approve(alice, addr, TokenID("1"))
	uassert.NoError(t, err, "should not result in error")

	approvedAddr, err := dummy.GetApproved(TokenID("1"))
	uassert.NoError(t, err, "should result in error")
	uassert.Equal(t, addr.String(), approvedAddr.String())
}

func TestTransferFrom(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")

	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(alice, TokenID("1"))
	dummy.mint(alice, TokenID("2"))

	err := dummy.TransferFrom(alice, alice, addr, TokenID("1"))
	uassert.NoError(t, err, "should result in error")

	// Check balance of alice after transfer
	balanceOfAlice, err := dummy.BalanceOf(alice)
	uassert.NoError(t, err, "should result in error")
	uassert.Equal(t, int64(1), balanceOfAlice)

	// Check balance of addr after transfer
	balanceOfAddr, err := dummy.BalanceOf(addr)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(1), balanceOfAddr)

	// Check Owner of transferred Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	uassert.NoError(t, err, "should result in error")
	uassert.Equal(t, addr.String(), owner.String())
}

func TestSafeTransferFrom(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")

	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(alice, TokenID("1"))
	dummy.mint(alice, TokenID("2"))

	err := dummy.SafeTransferFrom(alice, alice, addr, TokenID("1"))
	uassert.NoError(t, err, "should not result in error")

	// Check balance of alice after transfer
	balanceOfAlice, err := dummy.BalanceOf(alice)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(1), balanceOfAlice)

	// Check balance of addr after transfer
	balanceOfAddr, err := dummy.BalanceOf(addr)
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, int64(1), balanceOfAddr)

	// Check Owner of transferred Token id
	owner, err := dummy.OwnerOf(TokenID("1"))
	uassert.NoError(t, err, "should not result in error")
	uassert.Equal(t, addr.String(), owner.String())
}

func TestMint(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

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

func TestBurn(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(addr, TokenID("1"))
	dummy.mint(addr, TokenID("2"))

	err := dummy.Burn(TokenID("1"))
	uassert.NoError(t, err, "should not result in error")

	// Check Owner of Token id
	_, err = dummy.OwnerOf(TokenID("1"))
	uassert.Error(t, err, "should result in error")
}

func TestSetTokenURI(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	tokenURI := "http://example.com/token"

	dummy.mint(addr1, TokenID("1"))

	_, derr := dummy.SetTokenURI(addr1, TokenID("1"), TokenURI(tokenURI))
	uassert.NoError(t, derr, "should not result in error")

	// Test case: Invalid token ID
	_, err := dummy.SetTokenURI(addr1, TokenID("3"), TokenURI(tokenURI))
	uassert.ErrorIs(t, err, ErrInvalidTokenId)

	// addr2 trying to set URI for token 1 (not owner)
	_, cerr := dummy.SetTokenURI(addr2, TokenID("1"), TokenURI(tokenURI))
	uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

	dummyTokenURI, err := dummy.TokenURI(TokenID("1"))
	uassert.NoError(t, err, "TokenURI error")
	uassert.Equal(t, string(tokenURI), string(dummyTokenURI))
}

func TestIsApprovedOrOwner(cur realm, t *testing.T) {
	dummy := newDummyNFT(0, cur)
	uassert.True(t, dummy != nil, "should not be nil")

	var (
		owner    = testutils.TestAddress("owner")
		operator = testutils.TestAddress("operator")
		approved = testutils.TestAddress("approved")
		other    = testutils.TestAddress("other")
	)

	tid := TokenID("1")

	err := dummy.mint(owner, tid)
	uassert.NoError(t, err)

	// check owner
	isApprovedOrOwner := dummy.isApprovedOrOwner(owner, tid)
	uassert.True(t, isApprovedOrOwner, "owner should be approved")

	// check operator
	err = dummy.SetApprovalForAll(owner, operator, true)
	uassert.NoError(t, err)
	isApprovedOrOwner = dummy.isApprovedOrOwner(operator, tid)
	uassert.True(t, isApprovedOrOwner, "operator should be approved")

	// check approved
	err = dummy.Approve(owner, approved, tid)
	uassert.NoError(t, err)
	isApprovedOrOwner = dummy.isApprovedOrOwner(approved, tid)
	uassert.True(t, isApprovedOrOwner, "approved address should be approved")

	// check other
	isApprovedOrOwner = dummy.isApprovedOrOwner(other, tid)
	uassert.False(t, isApprovedOrOwner, "other address should not be approved")

	// check non-existent token
	isApprovedOrOwner = dummy.isApprovedOrOwner(owner, TokenID("999"))
	uassert.False(t, isApprovedOrOwner, "non-existent token should not be approved")
}
