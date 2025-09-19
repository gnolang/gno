package extendednft

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/moul/authz"
)

func TestExtendedNFT(t *testing.T) {
	admin := testutils.TestAddress("admin")
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")
	royaltyReceiver := testutils.TestAddress("royalty")

	initAuth := authz.NewMemberAuthority(admin)
	auth.Transfer(initAuth)

	testing.SetOriginCaller(admin)
	err := MintToken(alice, "token1")
	urequire.NoError(t, err)

	owner, err := NFT.OwnerOf("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, alice, owner)

	testing.SetOriginCaller(alice)
	metadata := Metadata{
		Name:        "Test Token",
		Description: "A test token with metadata",
		Image:       "https://example.com/image.png",
		Attributes: []Trait{
			{TraitType: "Color", Value: "Blue"},
			{TraitType: "Size", Value: "Medium"},
		},
	}

	err = SetTokenMetadata("token1", metadata)
	urequire.NoError(t, err)

	retrievedMetadata, err := GetTokenMetadata("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, metadata.Name, retrievedMetadata.Name)
	uassert.Equal(t, metadata.Description, retrievedMetadata.Description)
	uassert.Equal(t, metadata.Image, retrievedMetadata.Image)
	uassert.Equal(t, len(metadata.Attributes), len(retrievedMetadata.Attributes))

	err = SetTokenRoyalty("token1", royaltyReceiver, 10) // 10% royalty
	urequire.NoError(t, err)

	royaltyInfo, err := GetTokenRoyalty("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, royaltyReceiver, royaltyInfo.PaymentAddress)
	uassert.Equal(t, uint64(10), royaltyInfo.Percentage)

	receiver, amount, err := CalculateRoyalty("token1", 1000)
	urequire.NoError(t, err)
	uassert.Equal(t, royaltyReceiver, receiver)
	uassert.Equal(t, uint64(100), amount) // 10% of 1000 = 100

	err = Teller.Transfer(bob, "token1")
	urequire.NoError(t, err)

	owner, err = NFT.OwnerOf("token1")
	urequire.NoError(t, err)
	uassert.Equal(t, bob, owner)

	testing.SetOriginCaller(alice)
	err = SetTokenMetadata("token1", metadata)
	if err == nil {
		t.Error("Expected error when non-owner tries to set metadata")
	}

	err = SetTokenRoyalty("token1", royaltyReceiver, 5)
	if err == nil {
		t.Error("Expected error when non-owner tries to set royalty")
	}

	testing.SetOriginCaller(admin)
	err = BatchMintTokens(alice, []TokenID{"token2", "token3"})
	urequire.NoError(t, err)

	balance, err := NFT.BalanceOf(alice)
	urequire.NoError(t, err)
	uassert.Equal(t, uint64(2), balance)

	err = BurnToken("token2")
	urequire.NoError(t, err)

	_, err = NFT.OwnerOf("token2")
	if err == nil {
		t.Error("Expected error when querying burned token")
	}
}

func TestAdminFunctions(t *testing.T) {
	admin := testutils.TestAddress("admin")
	newAdmin := testutils.TestAddress("newadmin")
	alice := testutils.TestAddress("alice")

	initAuth := authz.NewMemberAuthority(admin)
	auth.Transfer(initAuth)

	testing.SetOriginCaller(admin)
	err := AddAdmin(newAdmin)
	urequire.NoError(t, err)

	testing.SetOriginCaller(newAdmin)
	err = MintToken(alice, "token4")
	urequire.NoError(t, err)

	testing.SetOriginCaller(admin)
	err = RemoveAdmin(newAdmin)
	urequire.NoError(t, err)

	testing.SetOriginCaller(newAdmin)
	err = MintToken(alice, "token5")
	if err == nil {
		t.Error("Expected error when removed admin tries to mint")
	}
}

func TestTransferAuthority(t *testing.T) {
	admin := testutils.TestAddress("admin")
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	initAuth := authz.NewMemberAuthority(admin)
	auth.Transfer(initAuth)

	newAuth := authz.NewMemberAuthority(alice)

	testing.SetOriginCaller(admin)
	err := TransferAuthority(newAuth)
	urequire.NoError(t, err)

	err = MintToken(bob, "token6")
	if err == nil {
		t.Error("Expected error when old admin tries to mint")
	}

	testing.SetOriginCaller(alice)
	err = MintToken(bob, "token6")
	urequire.NoError(t, err)

	owner, err := NFT.OwnerOf("token6")
	urequire.NoError(t, err)
	uassert.Equal(t, bob, owner)
}
