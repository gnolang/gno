package pausablenft

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/uassert"
	"gno.land/p/moul/authz"
	"gno.land/p/demo/grc/grc721remake"
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

	std.TestSetOriginCaller(admin)
	err := SetPaused(true)
	uassert.NoError(t, err)
	uassert.Equal(t, true, IsPaused())

	err = SetPaused(false)
	uassert.NoError(t, err)
	uassert.Equal(t, false, IsPaused())

	tokenID := "token1"
	std.TestSetOriginCaller(admin)
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
	
	std.TestSetOriginCaller(alice)
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
	
	std.TestSetOriginCaller(admin)
	err := AddAdmin(alice)
	uassert.NoError(t, err)
	
	std.TestSetOriginCaller(alice)
	err = SetPaused(true)
	uassert.NoError(t, err)
	uassert.Equal(t, true, IsPaused())
	
	std.TestSetOriginCaller(admin)
	err = RemoveAdmin(alice)
	uassert.NoError(t, err)
	
	std.TestSetOriginCaller(alice)
	err = SetPaused(false)
	if err == nil {
		t.Error("Expected error when non-admin tries to set paused state")
	}
	
	std.TestSetOriginCaller(admin)
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
	
	std.TestSetOriginCaller(admin)
	err := TransferAuthority(newAuth)
	uassert.NoError(t, err)
	
	err = SetPaused(true)
	if err == nil {
		t.Error("Expected error when old admin tries to set paused state")
	}
	
	std.TestSetOriginCaller(alice)
	err = SetPaused(true)
	uassert.NoError(t, err)
	uassert.Equal(t, true, IsPaused())
	
	err = SetPaused(false)
	uassert.NoError(t, err)
}
