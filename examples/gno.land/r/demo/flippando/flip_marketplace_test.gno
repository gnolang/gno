package flippando

import (
	"std"
	"testing"

	"gno.land/r/demo/users"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/avl"
)

var (
	dummyNFTName   = "DummyNFT"
	dummyNFTSymbol = "DNFT"
)

/*
func TestListNFT(t *testing.T) {
    marketplace, compositeNFT, _ := setupMockMarketplace()

    seller := users.AddressOrName("g1seller")
    compositeTokenID := grc721.TokenID("1")
    basicTokenIDs := []grc721.TokenID{grc721.TokenID("1"), grc721.TokenID("2")}

    // Mock minting of composite NFT
    compositeNFT.mint(seller.Resolve(), compositeTokenID)

    err := marketplace.ListNFT(compositeTokenID, basicTokenIDs, seller.Resolve(), 100)
    if err != nil {
        t.Errorf("listing NFT failed: %v", err)
    }

    _, exists := marketplace.listings.Get(string(compositeTokenID))
    if !exists {
        t.Errorf("NFT not listed in the marketplace")
    }
}

func TestRemoveListing(t *testing.T) {
    marketplace, compositeNFT, _, _ := setupMockMarketplace()

    seller := users.AddressOrName("g1seller")
    compositeTokenID := TokenID("1")
    basicTokenIDs := []TokenID{TokenID("1"), TokenID("2")}

    mintMockNFTs(compositeNFT, seller.Resolve(), compositeTokenID)
    marketplace.ListNFT(compositeTokenID, basicTokenIDs, seller.Resolve(), 100)

    err := marketplace.RemoveListing(compositeTokenID, seller.Resolve())
    if err != nil {
        t.Errorf("removing listing failed: %v", err)
    }

    _, exists := marketplace.listings.Get(string(compositeTokenID))
    if exists {
        t.Errorf("NFT listing was not removed")
    }
}

func TestUnauthorizedListingRemoval(t *testing.T) {
    marketplace, compositeNFT, _, _ := setupMockMarketplace()

    seller := users.AddressOrName("g1seller")
    unauthorized := users.AddressOrName("g1unauth")
    compositeTokenID := TokenID("1")
    basicTokenIDs := []TokenID{TokenID("1"), TokenID("2")}

    mintMockNFTs(compositeNFT, seller.Resolve(), compositeTokenID)
    marketplace.ListNFT(compositeTokenID, basicTokenIDs, seller.Resolve(), 100)

    err := marketplace.RemoveListing(compositeTokenID, unauthorized.Resolve())
    if err == nil {
        t.Errorf("unauthorized removal did not fail as expected")
    }
}

func TestBuyNFT(t *testing.T) {
    marketplace, compositeNFT, basicNFT, flipToken := setupMockMarketplace()

    seller := users.AddressOrName("g1seller")
    buyer := users.AddressOrName("g1buyer")
    compositeTokenID := TokenID("1")
    basicTokenIDs := []TokenID{TokenID("1"), TokenID("2")}

    mintMockNFTs(compositeNFT, seller.Resolve(), compositeTokenID)
    mintMockNFTs(basicNFT, seller.Resolve(), basicTokenIDs...)
    flipToken.Mint(buyer.Resolve(), 1000)
    marketplace.ListNFT(compositeTokenID, basicTokenIDs, seller.Resolve(), 100)

    err := marketplace.BuyNFT(buyer.Resolve(), compositeTokenID)
    if err != nil {
        t.Errorf("buying NFT failed: %v", err)
    }

    newOwner, _ := compositeNFT.OwnerOf(compositeTokenID)
    if newOwner != buyer.Resolve() {
        t.Errorf("ownership not transferred to buyer")
    }
}

func TestBuyNFTWithInsufficientFLIP(t *testing.T) {
    marketplace, compositeNFT, basicNFT, flipToken := setupMockMarketplace()

    seller := users.AddressOrName("g1seller")
    buyer := users.AddressOrName("g1buyer")
    compositeTokenID := TokenID("1")
    basicTokenIDs := []TokenID{TokenID("1"), TokenID("2")}

    mintMockNFTs(compositeNFT, seller.Resolve(), compositeTokenID)
    mintMockNFTs(basicNFT, seller.Resolve(), basicTokenIDs...)
    flipToken.Mint(buyer.Resolve(), 50) // Insufficient balance
    marketplace.ListNFT(compositeTokenID, basicTokenIDs, seller.Resolve(), 100)

    err := marketplace.BuyNFT(buyer.Resolve(), compositeTokenID)
    if err == nil {
        t.Errorf("expected failure due to insufficient FLIP tokens, but transaction succeeded")
    }
}


func setupMockMarketplace() (*Marketplace, grc721.IGRC721, *grc20.AdminToken) {
    // Mock setup for flipCompositeNFT and FLIP token
    compositeNFT := grc721.NewBasicNFT("Flippando Composite NFT", "CFLP")
    basicNFT := grc721.NewBasicNFT("Flippando Basic NFT", "BFLP")
    flipToken := grc20.NewAdminToken("Flippando", "FLIP", 6)

    // Create a marketplace instance with mock contracts
    marketplace := NewMarketplace(compositeNFT, basicNFT, flipToken)
    return marketplace, compositeNFT, basicNFT, flipToken
}



func mintMockNFTs(nftContract grc721.IGRC721, owner std.Address, tokenIDs ...grc721.TokenID) {
    for _, tokenID := range tokenIDs {
        nftContract.mint(owner, tokenID)
    }
}*/

