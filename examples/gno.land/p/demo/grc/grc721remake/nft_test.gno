package grc721remake

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestTestImpl(t *testing.T) {
	nft, _ := NewNFT("Dummy NFT", "DNFT")
	urequire.False(t, nft == nil, "dummy should not be nil")
}

func TestNFT(t *testing.T) {
	var (
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
		carl  = testutils.TestAddress("carl")
	)

	nft, ledger := NewNFT("Dummy NFT", "DNFT")

	checkBalances := func(aliceEB, bobEB, carlEB uint64) {
		t.Helper()
		exp := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceEB, bobEB, carlEB)
		aliceGB, _ := nft.BalanceOf(alice)
		bobGB, _ := nft.BalanceOf(bob)
		carlGB, _ := nft.BalanceOf(carl)
		got := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceGB, bobGB, carlGB)
		uassert.Equal(t, got, exp, "invalid balances")
	}

	checkOwnership := func(tid TokenID, expectedOwner std.Address) {
		t.Helper()
		owner, err := nft.OwnerOf(tid)
		urequire.NoError(t, err)
		uassert.Equal(t, owner, expectedOwner, "invalid owner for token "+string(tid))
	}

	checkApproval := func(tid TokenID, expectedApproved std.Address) {
		t.Helper()
		approved, err := nft.GetApproved(tid)
		if err != nil && err != ErrTokenIdNotHasApproved {
			t.Fatalf("unexpected error: %v", err)
		}
		if err == nil {
			uassert.Equal(t, approved, expectedApproved, "invalid approval for token "+string(tid))
		}
	}

	checkBalances(0, 0, 0)
	uassert.Equal(t, nft.TokenCount(), uint64(0), "initial token count should be 0")
	uassert.Equal(t, nft.KnownAccounts(), 0, "initial known accounts should be 0")

	urequire.NoError(t, ledger.Mint(alice, "token1"))
	checkBalances(1, 0, 0)
	checkOwnership("token1", alice)
	uassert.Equal(t, nft.TokenCount(), uint64(1), "token count should be 1 after minting")
	uassert.Equal(t, nft.KnownAccounts(), 1, "known accounts should be 1 after minting")

	urequire.NoError(t, ledger.BatchMint(alice, []TokenID{"token2", "token3"}))
	checkBalances(3, 0, 0)
	checkOwnership("token2", alice)
	checkOwnership("token3", alice)
	uassert.Equal(t, nft.TokenCount(), uint64(3), "token count should be 3 after batch minting")

	urequire.NoError(t, ledger.Approve(alice, bob, "token1"))
	checkApproval("token1", bob)

	urequire.NoError(t, ledger.TransferFrom(bob, alice, carl, "token1"))
	checkBalances(2, 0, 1)
	checkOwnership("token1", carl)

	_, err := nft.GetApproved("token1")
	uassert.ErrorContains(t, err, ErrTokenIdNotHasApproved.Error(), "approval should be cleared after transfer")

	urequire.NoError(t, ledger.Transfer(alice, bob, "token2"))
	checkBalances(1, 1, 1)
	checkOwnership("token2", bob)

	urequire.NoError(t, ledger.SetApprovalForAll(alice, bob, true))
	uassert.True(t, nft.IsApprovedForAll(alice, bob), "bob should be approved for all of alice's tokens")

	urequire.NoError(t, ledger.TransferFrom(bob, alice, carl, "token3"))
	checkBalances(0, 1, 2)
	checkOwnership("token3", carl)

	urequire.NoError(t, ledger.BatchMint(bob, []TokenID{"token4", "token5"}))
	checkBalances(0, 3, 2)
	urequire.NoError(t, ledger.BatchTransfer(bob, carl, []TokenID{"token4", "token5"}))
	checkBalances(0, 1, 4)
	checkOwnership("token4", carl)
	checkOwnership("token5", carl)

	urequire.NoError(t, ledger.SetApprovalForAll(carl, bob, true))
	urequire.NoError(t, ledger.Burn("token1"))
	checkBalances(0, 1, 3)
	uassert.Equal(t, nft.TokenCount(), uint64(4), "token count should be 4 after burning")

	urequire.NoError(t, ledger.BatchBurn([]TokenID{"token3", "token4"}))
	checkBalances(0, 1, 1)
	uassert.Equal(t, nft.TokenCount(), uint64(2), "token count should be 2 after batch burning")

	err = ledger.Mint(alice, "token2")
	uassert.ErrorContains(t, err, ErrTokenIdAlreadyExists.Error(), "should not be able to mint existing token")

	err = ledger.Transfer(alice, bob, "nonexistent")
	uassert.ErrorContains(t, err, ErrInvalidTokenId.Error(), "should not be able to transfer non-existent token")

	err = ledger.Approve(alice, bob, "nonexistent")
	uassert.ErrorContains(t, err, ErrInvalidTokenId.Error(), "should not be able to approve non-existent token")

	err = ledger.Transfer(alice, bob, "token5")
	uassert.ErrorContains(t, err, ErrTransferFromIncorrectOwner.Error(), "should not be able to transfer token you don't own")
}

func TestMetadata(t *testing.T) {
	nft, _ := NewNFT("Dummy NFT", "DNFT")

	uassert.Equal(t, nft.GetName(), "Dummy NFT", "name should match")
	uassert.Equal(t, nft.GetSymbol(), "DNFT", "symbol should match")

	getter := nft.Getter()
	nftFromGetter := getter()
	uassert.Equal(t, nftFromGetter.GetName(), "Dummy NFT", "name from getter should match")
}
