package flippando

import (
    "std"
	"encoding/json"
    "gno.land/p/demo/avl"
)

type Marketplace struct {
    listings avl.Tree  // CompositeTokenID -> Listing
}

type Listing struct {
    CompositeTokenID grc721.TokenID   // TokenID of the composite NFT
    BasicTokenIDs    []grc721.TokenID // Slice of TokenIDs for the basic NFTs contained within the composite NFT
    Seller           std.Address      // Address of the seller
    Price            uint64           // Sale price
}


func NewMarketplace() *Marketplace {
    return &Marketplace{
        listings: avl.Tree{},
    }
}

func (m *Marketplace) ListNFT(compositeTokenID grc721f.TokenID, basicTokenIDs []grc721f.TokenID, seller std.Address, price uint64) error {
    // get tokenURI of the composite NFT
	cTokenURIString, err := GetCompositeTokenURI(string(compositeTokenID));
	if err != nil {
		return error("can't get composite token uri")
	}
	var cTokenURI compositeNFTMetaData;
	err := json.Unmarshal([]byte(cTokenURIString), &cTokenURI)
    if err != nil {
        error("can't unmarshal token uri")
    }

	var basicTokenIDs []grc721.TokenID
    for _, id := range cTokenURI.bTokenIDs {
        basicTokenIDs = append(basicTokenIDs, grc721.TokenID(strconv.FormatInt(id, 10)))
    }
	
	// Create a new listing
    newListing := Listing{
        CompositeTokenID: compositeTokenID,
        BasicTokenIDs:    basicTokenIDs,
        Seller:           seller,
        Price:            price,
    }

    // Add the new listing to the marketplace
    m.listings.Set(string(compositeTokenID), newListing)
    return nil
}

func (m *Marketplace) RemoveListing(compositeTokenID grc721f.TokenID, seller std.Address) error {
    // Retrieve the listing
    rawListing, exists := m.listings.Get(string(compositeTokenID))
    if !exists {
        return std.Error("listing does not exist")
    }

    listing := rawListing.(Listing)

    // Verify that the caller is the seller
    if listing.Seller != seller {
        return std.Error("unauthorized - caller is not the seller")
    }

    // Remove the listing from the marketplace
    m.listings.Remove(string(compositeTokenID))
    return nil
}

func (m *Marketplace) GetListings() []Listing {
    var listings []Listing
    for _, key := range m.listings.Keys() {
        rawListing, _ := m.listings.Get(key)
        listing := rawListing.(Listing)
        listings = append(listings, listing)
    }
    return listings
}


func (m *Marketplace) BuyNFT(buyer std.Address, compositeTokenID grc721f.TokenID) error {
    listingRaw, exists := m.listings.Get(string(compositeTokenID))
    if !exists {
        return error("listing does not exist")
    }
    listing := listingRaw.(Listing)

    // Use the specific function to transfer FLIP tokens
    TransferFLIPToSeller(buyer, listing.Seller, listing.Price)

    // Use the specific function to transfer the composite NFT
    err := TransferCompositeNFTToBuyer(listing.Seller, buyer, string(compositeTokenID))
    if err != nil {
        return err
    }

    // Assume a function to retrieve basic NFT token IDs from a composite NFT
    // basicTokenIDs := GetBasicNFTsFromComposite(compositeTokenID)

    // Transfer each basic NFT to the buyer
    // for _, basicTokenID := range basicTokenIDs {
    //     err = TransferBasicNFTToBuyer(listing.Seller, buyer, string(basicTokenID))
    //     if err != nil {
    //         return err
    //     }
    // }

    // Remove the listing after successful transfer
    m.listings.Remove(string(compositeTokenID))
    return nil
}

