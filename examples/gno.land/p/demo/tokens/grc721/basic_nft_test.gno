package grc721

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
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
	uassert.Equal(t, int64(0), count)

	dummy.mint("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", TokenID("1"))
	dummy.mint("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm", TokenID("2"))

	count = dummy.TokenCount()
	uassert.Equal(t, int64(2), count)
}

func TestBalanceOf(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
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

func TestOwnerOf(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
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

func TestIsApprovedForAll(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

	isApprovedForAll := dummy.IsApprovedForAll(addr1, addr2)
	uassert.False(t, isApprovedForAll)
}

func TestSetApprovalForAll(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := runtime.CurrentRealm().Address()
	testing.SetOriginCaller(caller)
	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	// Test setting approval to true
	isApprovedForAll := dummy.IsApprovedForAll(caller, addr)
	uassert.False(t, isApprovedForAll)

	// set beyond origin caller
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
	err := dummy.SetApprovalForAll(addr, true)
	uassert.NoError(t, err, "should not result in error")

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	uassert.True(t, isApprovedForAll)

	// Test setting approval to false
	err = dummy.SetApprovalForAll(addr, false)
	uassert.NoError(t, err, "should not result in error")

	isApprovedForAll = dummy.IsApprovedForAll(caller, addr)
	uassert.False(t, isApprovedForAll)
}

func TestGetApproved(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	_, err := dummy.GetApproved(TokenID("invalid"))
	uassert.Error(t, err, "should result in error")
}

func TestApprove(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)

	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := runtime.CurrentRealm().Address()
	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	err := dummy.mint(caller, TokenID("1"))
	uassert.NoError(t, err, "failed to mint")

	_, err = dummy.GetApproved(TokenID("1"))
	uassert.Error(t, err, "should result in error")

	testing.SetRealm(testing.NewUserRealm(alice))
	func() {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
		err = dummy.Approve(addr, TokenID("1"))
		uassert.NoError(t, err, "should not result in error")

		approvedAddr, err := dummy.GetApproved(TokenID("1"))
		uassert.NoError(t, err, "should result in error")
		uassert.Equal(t, addr.String(), approvedAddr.String())
	}()
}

func TestTransferFrom(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)

	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := runtime.CurrentRealm().Address()
	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(caller, TokenID("1"))
	dummy.mint(caller, TokenID("2"))

	testing.SetRealm(testing.NewUserRealm(alice))
	func() {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))

		err := dummy.TransferFrom(caller, addr, TokenID("1"))
		uassert.NoError(t, err, "should result in error")

		// Check balance of caller after transfer
		balanceOfCaller, err := dummy.BalanceOf(caller)
		uassert.NoError(t, err, "should result in error")
		uassert.Equal(t, int64(1), balanceOfCaller)

		// Check balance of addr after transfer
		balanceOfAddr, err := dummy.BalanceOf(addr)
		uassert.NoError(t, err, "should not result in error")
		uassert.Equal(t, int64(1), balanceOfAddr)

		// Check Owner of transferred Token id
		owner, err := dummy.OwnerOf(TokenID("1"))
		uassert.NoError(t, err, "should result in error")
		uassert.Equal(t, addr.String(), owner.String())
	}()
}

func TestSafeTransferFrom(t *testing.T) {
	alice := testutils.TestAddress("alice")
	testing.SetOriginCaller(alice)

	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	caller := runtime.CurrentRealm().Address()
	addr := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")

	dummy.mint(caller, TokenID("1"))
	dummy.mint(caller, TokenID("2"))

	testing.SetRealm(testing.NewUserRealm(alice))
	func() {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))

		err := dummy.SafeTransferFrom(caller, addr, TokenID("1"))
		uassert.NoError(t, err, "should not result in error")

		// Check balance of caller after transfer
		balanceOfCaller, err := dummy.BalanceOf(caller)
		uassert.NoError(t, err, "should not result in error")
		uassert.Equal(t, int64(1), balanceOfCaller)

		// Check balance of addr after transfer
		balanceOfAddr, err := dummy.BalanceOf(addr)
		uassert.NoError(t, err, "should not result in error")
		uassert.Equal(t, int64(1), balanceOfAddr)

		// Check Owner of transferred Token id
		owner, err := dummy.OwnerOf(TokenID("1"))
		uassert.NoError(t, err, "should not result in error")
		uassert.Equal(t, addr.String(), owner.String())
	}()
}

func TestMint(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
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

func TestBurn(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
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

func TestSetTokenURI(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
	uassert.True(t, dummy != nil, "should not be nil")

	addr1 := address("g1var589z07ppjsjd24ukm4uguzwdt0tw7g47cgm")
	addr2 := address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	tokenURI := "http://example.com/token"

	testing.SetOriginCaller(addr1) // addr1

	dummy.mint(addr1, TokenID("1"))

	testing.SetRealm(testing.NewUserRealm(addr1))
	func() {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))

		_, derr := dummy.SetTokenURI(TokenID("1"), TokenURI(tokenURI))
		uassert.NoError(t, derr, "should not result in error")

		// Test case: Invalid token ID
		_, err := dummy.SetTokenURI(TokenID("3"), TokenURI(tokenURI))
		uassert.ErrorIs(t, err, ErrInvalidTokenId)

		testing.SetOriginCaller(addr2) // addr2
		func() {
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
			_, cerr := dummy.SetTokenURI(TokenID("1"), TokenURI(tokenURI)) // addr2 trying to set URI for token 1
			uassert.ErrorIs(t, cerr, ErrCallerIsNotOwner)

			// Test case: Retrieving TokenURI
			testing.SetOriginCaller(addr1) // addr1

			dummyTokenURI, err := dummy.TokenURI(TokenID("1"))
			uassert.NoError(t, err, "TokenURI error")
			uassert.Equal(t, string(tokenURI), string(dummyTokenURI))
		}()
	}()
}

func TestIsApprovedOrOwner(t *testing.T) {
	dummy := NewBasicNFT(dummyNFTName, dummyNFTSymbol)
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
	testing.SetOriginCaller(owner)
	func() {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))

		err = dummy.SetApprovalForAll(operator, true)
		uassert.NoError(t, err)
		isApprovedOrOwner = dummy.isApprovedOrOwner(operator, tid)
		uassert.True(t, isApprovedOrOwner, "operator should be approved")

		// check approved
		testing.SetOriginCaller(owner)
		func() {
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
			err = dummy.Approve(approved, tid)
			uassert.NoError(t, err)
			isApprovedOrOwner = dummy.isApprovedOrOwner(approved, tid)
			uassert.True(t, isApprovedOrOwner, "approved address should be approved")

			// check other
			isApprovedOrOwner = dummy.isApprovedOrOwner(other, tid)
			uassert.False(t, isApprovedOrOwner, "other address should not be approved")

			// check non-existent token
			isApprovedOrOwner = dummy.isApprovedOrOwner(owner, TokenID("999"))
			uassert.False(t, isApprovedOrOwner, "non-existent token should not be approved")
		}()
	}()
}
