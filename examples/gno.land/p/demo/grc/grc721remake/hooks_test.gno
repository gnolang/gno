package grc721remake

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

// PausableHook is an example hook that can pause transfers
type PausableHook struct {
	paused bool
}

func (h *PausableHook) GetHookType() HookType {
	return HookTransfer
}

func (h *PausableHook) GetHookTime() HookTime {
	return HookBefore
}

func (h *PausableHook) OnTransfer(nft *NFT, from, to std.Address, tid TokenID) error {
	if h.paused {
		return ufmt.Errorf("transfers are paused")
	}
	return nil
}

// RoyaltyHook is an example hook that can enforce royalties on transfers
type RoyaltyHook struct {
	royaltyReceiver std.Address
	royaltyPercent  int
}

func (h *RoyaltyHook) GetHookType() HookType {
	return HookTransfer
}

func (h *RoyaltyHook) GetHookTime() HookTime {
	return HookBefore
}

func (h *RoyaltyHook) OnTransfer(nft *NFT, from, to std.Address, tid TokenID) error {
	println("RoyaltyHook OnTransfer called")
	return nil
}

// LoggingHook is an example hook that logs all operations
type LoggingHook struct{}

func (h *LoggingHook) GetHookType() HookType {
	return HookMint
}

func (h *LoggingHook) GetHookTime() HookTime {
	return HookAfter
}

func (h *LoggingHook) OnMint(nft *NFT, to std.Address, tid TokenID) error {
	println("LoggingHook OnMint called")
	return nil
}

// BurnLimitHook is an example hook that limits the number of tokens that can be burned
type BurnLimitHook struct {
	burnCount int
	maxBurns  int
}

func (h *BurnLimitHook) GetHookType() HookType {
	return HookBurn
}

func (h *BurnLimitHook) GetHookTime() HookTime {
	return HookBefore
}

func (h *BurnLimitHook) OnBurn(nft *NFT, tid TokenID) error {
	if h.burnCount >= h.maxBurns {
		return ufmt.Errorf("burn limit reached: maximum %d burns allowed", h.maxBurns)
	}
	h.burnCount++
	return nil
}

func TestPausableHook(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	nft, ledger := NewNFT("Dummy NFT", "DNFT")

	pausable := &PausableHook{paused: false}
	ledger.RegisterHook("pausable", pausable)

	urequire.NoError(t, ledger.Mint(alice, "token1"))
	urequire.NoError(t, ledger.Transfer(alice, bob, "token1"))

	owner, err := nft.OwnerOf("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, owner, bob, "token should be transferred to bob")

	retrievedHook, err := ledger.GetHook("pausable")
	uassert.NoError(t, err)
	pausableHook := retrievedHook.(*PausableHook)
	pausableHook.paused = true

	urequire.NoError(t, ledger.Mint(bob, "token2"))

	err = ledger.Transfer(bob, alice, "token2")
	uassert.Error(t, err, "transfer should fail when paused")
	uassert.ErrorContains(t, err, "transfers are paused", "error should indicate transfers are paused")

	owner, err = nft.OwnerOf("token2")
	urequire.NoError(t, err)
	uassert.Equal(t, owner, bob, "token should still be owned by bob")

	pausable.paused = false

	urequire.NoError(t, ledger.Transfer(bob, alice, "token2"))

	owner, err = nft.OwnerOf("token2")
	urequire.NoError(t, err)
	uassert.Equal(t, owner, alice, "token should be transferred to alice")
}

func TestRoyaltyHook(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	creator := testutils.TestAddress("creator")

	nft, ledger := NewNFT("Royalty NFT", "RNFT")

	royalty := &RoyaltyHook{
		royaltyReceiver: creator,
		royaltyPercent:  10,
	}
	ledger.RegisterHook("royalty", royalty)

	urequire.NoError(t, ledger.Mint(alice, "token1"))
	urequire.NoError(t, ledger.Transfer(alice, bob, "token1"))

	owner, err := nft.OwnerOf("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, owner, bob, "token should be transferred to bob")
}

func TestLoggingHook(t *testing.T) {
	alice := testutils.TestAddress("alice")

	nft, ledger := NewNFT("Logging NFT", "LNFT")

	logger := &LoggingHook{}
	ledger.RegisterHook("logging", logger)

	urequire.NoError(t, ledger.Mint(alice, "token1"))
}

func TestBurnLimitHook(t *testing.T) {
	alice := testutils.TestAddress("alice")

	nft, ledger := NewNFT("Burn Limit NFT", "BLNFT")

	burnLimit := &BurnLimitHook{
		burnCount: 0,
		maxBurns:  2,
	}
	ledger.RegisterHook("burnLimit", burnLimit)

	urequire.NoError(t, ledger.Mint(alice, "token1"))
	urequire.NoError(t, ledger.Mint(alice, "token2"))
	urequire.NoError(t, ledger.Mint(alice, "token3"))

	urequire.NoError(t, ledger.Burn("token1"))
	urequire.NoError(t, ledger.Burn("token2"))

	err := ledger.Burn("token3")
	uassert.Error(t, err, "burn should fail when limit reached")
	uassert.ErrorContains(t, err, "burn limit reached", "error should indicate burn limit reached")

	_, err = nft.OwnerOf("token3")
	uassert.NoError(t, err, "token3 should still exist")
}

func TestMultipleHooks(t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	creator := testutils.TestAddress("creator")

	nft, ledger := NewNFT("Multi Hook NFT", "MHNFT")

	pausable := &PausableHook{paused: false}
	royalty := &RoyaltyHook{
		royaltyReceiver: creator,
		royaltyPercent:  10,
	}
	logger := &LoggingHook{}

	ledger.RegisterHook("pausable", pausable)
	ledger.RegisterHook("royalty", royalty)
	ledger.RegisterHook("logging", logger)

	urequire.NoError(t, ledger.Mint(alice, "token1"))

	urequire.NoError(t, ledger.Transfer(alice, bob, "token1"))

	pausable.paused = true

	err := ledger.Transfer(bob, alice, "token1")
	uassert.Error(t, err, "transfer should fail when paused")

	owner, err := nft.OwnerOf("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, owner, bob, "token should still be owned by bob")
}
