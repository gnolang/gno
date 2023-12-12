package flippando

import (
    "std"
    "gno.land/p/demo/avl"
    "gno.land/p/demo/grc/grc721f"
    "gno.land/p/demo/grc/grc20" // Import the GRC20 package
)

type Marketplace struct {
    listings           avl.Tree            // CompositeTokenID -> Listing
    compositeNFTContract *grc721f.NewBasicNFT  // Reference to the composite NFT contract
    basicNFTContract    *grc721f.NewBasicNFT   // Reference to the basic NFT contract
    flipToken           *grc20.AdminToken  // Reference to the FLIP token contract
}

type Listing struct {
    CompositeTokenID grc721f.TokenID   // TokenID of the composite NFT
    BasicTokenIDs    []grc721f.TokenID // Slice of TokenIDs for the basic NFTs contained within the composite NFT
    Seller           std.Address      // Address of the seller
    Price            uint64           // Sale price
}

func NewMarketplace(compositeNFTContract, basicNFTContract *grc721f.BasicNFT, flipToken *grc20.AdminToken) *Marketplace {
    return &Marketplace{
        listings: avl.Tree{},
        compositeNFTContract: compositeNFTContract,
        basicNFTContract: basicNFTContract,
        flipToken: flipToken,
    }
}

func (m *Marketplace) ListNFT(compositeTokenID grc721f.TokenID, basicTokenIDs []grc721f.TokenID, seller std.Address, price uint64) error {
    // Additional checks and logic...
    listing := Listing{
        CompositeTokenID: compositeTokenID,
        BasicTokenIDs:    basicTokenIDs,
        Seller:           seller,
        Price:            price,
    }
    m.listings.Set(string(compositeTokenID), listing)
    return nil
}

func (m *Marketplace) RemoveListing(compositeTokenID grc721f.TokenID, seller std.Address) error {
    // Retrieve the listing from the marketplace
    rawListing, exists := m.listings.Get(string(compositeTokenID))
    if !exists {
        return std.ErrNotFound // Listing does not exist
    }

    listing := rawListing.(Listing)
    
    // Verify that the caller is the seller of the NFT
    if listing.Seller != seller {
        return std.ErrUnauthorized // Caller is not the seller
    }

    // Remove the listing from the marketplace
    m.listings.Remove(string(compositeTokenID))
    return nil
}


func (m *Marketplace) BuyNFT(buyer std.Address, compositeTokenID grc721f.TokenID) error {
    listing, exists := m.listings.Get(string(compositeTokenID))
    if !exists {
        return std.ErrNotFound
    }

    // Transfer FLIP tokens from buyer to seller
    err := m.flipToken.TransferFrom(buyer, listing.(Listing).Seller, listing.(Listing).Price)
    if err != nil {
        return err
    }

    // Transfer composite NFT from seller to buyer
    err = m.compositeNFTContract.TransferFrom(listing.(Listing).Seller, buyer, compositeTokenID)
    if err != nil {
        return err
    }

    // Transfer all basic NFTs contained within the composite NFT
    for _, basicTokenID := range listing.(Listing).BasicTokenIDs {
        err = m.basicNFTContract.TransferFrom(listing.(Listing).Seller, buyer, basicTokenID)
        if err != nil {
            return err
        }
    }

    // Remove the listing after successful transfer
    m.listings.Remove(string(compositeTokenID))
    return nil
}
