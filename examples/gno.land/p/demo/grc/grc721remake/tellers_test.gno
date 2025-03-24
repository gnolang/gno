package grc721remake

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestCallerTellerImpl(t *testing.T) {
	nft, _ := NewNFT("Dummy NFT", "DNFT")
	teller := nft.CallerTeller()
	urequire.False(t, nft == nil)
	var _ NFTTeller = teller
}

func TestTeller(t *testing.T) {
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
		uassert.Equal(t, owner, expectedOwner, "invalid owner")
	}

	checkBalances(0, 0, 0)

	urequire.NoError(t, ledger.Mint(alice, "token1"))
	urequire.NoError(t, ledger.Mint(alice, "token2"))
	checkBalances(2, 0, 0)
	checkOwnership("token1", alice)
	checkOwnership("token2", alice)

	urequire.NoError(t, ledger.Approve(alice, bob, "token1"))
	approved, err := nft.GetApproved("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, approved, bob, "invalid approval")

	urequire.NoError(t, ledger.TransferFrom(bob, alice, carl, "token1"))
	checkBalances(1, 0, 1)
	checkOwnership("token1", carl)
	checkOwnership("token2", alice)

	urequire.NoError(t, ledger.Transfer(alice, bob, "token2"))
	checkBalances(0, 1, 1)
	checkOwnership("token2", bob)

	urequire.NoError(t, ledger.SetApprovalForAll(bob, alice, true))
	uassert.True(t, nft.IsApprovedForAll(bob, alice), "approval for all not set")

	urequire.NoError(t, ledger.TransferFrom(alice, bob, carl, "token2"))
	checkBalances(0, 0, 2)
	checkOwnership("token2", carl)
}

func TestCallerTeller(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	carl := testutils.TestAddress("carl")

	nft, ledger := NewNFT("Dummy NFT", "DNFT")
	teller := nft.CallerTeller()

	checkBalances := func(aliceEB, bobEB, carlEB uint64) {
		t.Helper()
		exp := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceEB, bobEB, carlEB)
		aliceGB, _ := nft.BalanceOf(alice)
		bobGB, _ := nft.BalanceOf(bob)
		carlGB, _ := nft.BalanceOf(carl)
		got := ufmt.Sprintf("alice=%d bob=%d carl=%d", aliceGB, bobGB, carlGB)
		uassert.Equal(t, got, exp, "invalid balances")
	}

	urequire.NoError(t, ledger.Mint(alice, "token1"))
	checkBalances(1, 0, 0)

	testing.SetOriginCaller(alice)

	urequire.NoError(t, teller.Approve(bob, "token1"))
	approved, err := nft.GetApproved("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, approved, bob, "invalid approval")

	testing.SetOriginCaller(bob)

	urequire.NoError(t, teller.TransferFrom(alice, carl, "token1"))
	checkBalances(0, 0, 1)
	testing.SetOriginCaller(carl)

	urequire.NoError(t, teller.Transfer(bob, "token1"))
	checkBalances(0, 1, 0)
}

func TestBatchOperations(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	nft, ledger := NewNFT("Dummy NFT", "DNFT")
	teller := nft.CallerTeller()

	urequire.NoError(t, ledger.Mint(alice, "token1"))
	urequire.NoError(t, ledger.Mint(alice, "token2"))
	urequire.NoError(t, ledger.Mint(alice, "token3"))

	aliceBalance, _ := nft.BalanceOf(alice)
	uassert.Equal(t, aliceBalance, uint64(3), "invalid balance after minting")

	testing.SetOriginCaller(alice)

	urequire.NoError(t, teller.BatchTransfer(bob, []TokenID{"token1", "token2"}))

	aliceBalance, _ = nft.BalanceOf(alice)
	bobBalance, _ := nft.BalanceOf(bob)
	uassert.Equal(t, aliceBalance, uint64(1), "invalid alice balance after batch transfer")
	uassert.Equal(t, bobBalance, uint64(2), "invalid bob balance after batch transfer")

	urequire.NoError(t, teller.Approve(bob, "token3"))

	testing.SetOriginCaller(alice)

	urequire.NoError(t, teller.Transfer(bob, "token3"))

	aliceBalance, _ = nft.BalanceOf(alice)
	bobBalance, _ = nft.BalanceOf(bob)
	uassert.Equal(t, aliceBalance, uint64(0), "invalid alice balance after transferFrom")
	uassert.Equal(t, bobBalance, uint64(3), "invalid bob balance after transferFrom")
}

func TestReadonlyTeller(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	nft, ledger := NewNFT("Dummy NFT", "DNFT")
	readonlyTeller := nft.ReadonlyTeller()

	urequire.NoError(t, ledger.Mint(alice, "token1"))

	err := readonlyTeller.Transfer(bob, "token1")
	uassert.ErrorContains(t, err, ErrReadOnly.Error(), "readonly teller should not allow transfers")

	err = readonlyTeller.Approve(bob, "token1")
	uassert.ErrorContains(t, err, ErrReadOnly.Error(), "readonly teller should not allow approvals")

	name := readonlyTeller.GetName()
	uassert.Equal(t, name, "Dummy NFT", "readonly teller should allow reading name")

	balance, err := readonlyTeller.BalanceOf(alice)
	urequire.NoError(t, err)
	uassert.Equal(t, balance, uint64(1), "readonly teller should allow reading balances")
}
