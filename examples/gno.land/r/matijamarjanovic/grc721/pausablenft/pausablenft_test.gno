package pausablenft

import (
	"testing"

	"gno.land/p/demo/grc/grc721remake"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/moul/authz"
)

func TestPausableNFT(t *testing.T) {
	var (
		admin = testutils.TestAddress("admin")
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
	)

	initAuth := authz.NewMemberAuthority(admin)
	auth.Transfer(initAuth)

	uassert.Equal(t, false, IsPaused())

	testing.SetOriginCaller(admin)
	err := SetPaused(true)
	uassert.NoError(t, err)
	uassert.Equal(t, true, IsPaused())

	err = SetPaused(false)
	uassert.NoError(t, err)
	uassert.Equal(t, false, IsPaused())

	tokenID := "token1"
	testing.SetOriginCaller(admin)
	err = ledger.Mint(admin, grc721remake.TokenID(tokenID))
	uassert.NoError(t, err)

	owner, err := NFT.OwnerOf(grc721remake.TokenID(tokenID))
	uassert.NoError(t, err)
	uassert.Equal(t, admin, owner)

	err = ledger.TransferFrom(admin, admin, alice, grc721remake.TokenID(tokenID))
	uassert.NoError(t, err)

	owner, err = NFT.OwnerOf(grc721remake.TokenID(tokenID))
	uassert.NoError(t, err)
	uassert.Equal(t, alice, owner)

	err = SetPaused(true)
	uassert.NoError(t, err)

	tokenID2 := "token2"
	err = ledger.Mint(admin, grc721remake.TokenID(tokenID2))
	if err == nil {
		t.Error("Expected error when minting while paused")
	}

	testing.SetOriginCaller(alice)
	err = ledger.TransferFrom(alice, alice, admin, grc721remake.TokenID(tokenID))
	if err == nil {
		t.Error("Expected error when transferring while paused")
	}
}

func TestAdminFunctions(t *testing.T) {
	var (
		admin = testutils.TestAddress("admin")
		alice = testutils.TestAddress("alice")
		bob   = testutils.TestAddress("bob")
	)

	initAuth := authz.NewMemberAuthority(admin)
	auth.Transfer(initAuth)

	testing.SetOriginCaller(admin)
	err := AddAdmin(alice)
	uassert.NoError(t, err)

	testing.SetOriginCaller(alice)
	err = SetPaused(true)
	uassert.NoError(t, err)
	uassert.Equal(t, true, IsPaused())

	testing.SetOriginCaller(admin)
	err = RemoveAdmin(alice)
	uassert.NoError(t, err)

	testing.SetOriginCaller(alice)
	err = SetPaused(false)
	if err == nil {
		t.Error("Expected error when non-admin tries to set paused state")
	}

	// Reset state for other tests
	testing.SetOriginCaller(admin)
	err = SetPaused(false)
	uassert.NoError(t, err)
}

func TestTransferAuthority(t *testing.T) {
	var (
		admin = testutils.TestAddress("admin")
		alice = testutils.TestAddress("alice")
	)

	initAuth := authz.NewMemberAuthority(admin)
	auth.Transfer(initAuth)

	newAuth := authz.NewMemberAuthority(alice)

	testing.SetOriginCaller(admin)
	err := TransferAuthority(newAuth)
	uassert.NoError(t, err)

	err = SetPaused(true)
	if err == nil {
		t.Error("Expected error when old admin tries to set paused state")
	}

	testing.SetOriginCaller(alice)
	err = SetPaused(true)
	uassert.NoError(t, err)
	uassert.Equal(t, true, IsPaused())

	err = SetPaused(false)
	uassert.NoError(t, err)
}
