package flippando

import (
	"errors"
	"strconv"
	"strings"
	"std"
	//"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
)

type Marketplace struct {
	listings map[string]Listing // CompositeTokenID -> Listing
	// listings avl.Tree  // CompositeTokenID -> Listing
}

type Listing struct {
	CompositeTokenID grc721.TokenID // TokenID of the composite NFT
	BasicTokenIDs    []int64        // Slice of TokenIDs for the basic NFTs contained within the composite NFT
	Seller           std.Address    // Address of the seller
	Price            uint64         // Sale price
}

func NewMarketplace() *Marketplace {
	return &Marketplace{
		listings: make(map[string]Listing),
		// listings: avl.Tree{},
	}
}

func (m *Marketplace) ListNFT(compositeTokenID grc721.TokenID, seller std.Address, price uint64) error {
	// get tokenURI of the composite NFT
    compositeTokenIDstr := string(compositeTokenID)
	basicTokenIDArray, err := GetChildrenOf(compositeTokenIDstr)
    if err != nil {
        return err
    }
    basicTokenIDs, _ := ParseString(basicTokenIDArray)
	// Create a new listing
	newListing := Listing{
		CompositeTokenID: compositeTokenID,
		BasicTokenIDs:    basicTokenIDs,
		Seller:           seller,
		Price:            price,
	}

	// Add the new listing to the marketplace
	m.listings[string(compositeTokenID)] = newListing
	// m.listings.Set(string(compositeTokenID), newListing)
	return nil
}

func (m *Marketplace) RemoveListing(compositeTokenID grc721.TokenID, seller std.Address) error {
	// Retrieve the listing
	rawListing, exists := m.listings[string(compositeTokenID)]
	// rawListing, exists := m.listings.Get(string(compositeTokenID))
	if !exists {
		return errors.New("listing does not exist")
	}

	listing := rawListing

	// Verify that the caller is the seller
	if listing.Seller != seller {
		return errors.New("unauthorized - caller is not the seller")
	}

	// Remove the listing from the marketplace
	delete(m.listings, string(compositeTokenID))
	// m.listings.Remove(string(compositeTokenID))
	return nil
}

func (m *Marketplace) GetListings() string {
	var allListings []string
	for _, value := range m.listings {
		// Type assert the value to a Listing
		listing := value
		//if !ok {
			// Handle the case where the value is not a Listing
		//	continue // Skip this iteration
		//}
		compositeTokenURI := GetCompositeTokenURI(string(listing.CompositeTokenID))
		compositeTokenURI = strings.Trim(compositeTokenURI, "\"")

		listingResult := &listingResult{
			compositeTokenID: string(listing.CompositeTokenID),
			tokenURI:         compositeTokenURI,
			seller:           string(listing.Seller),
			price:            strconv.FormatUint(listing.Price, 10),
		}
		listingResultJSON, _ := listingResult.MarshalJSON()
		allListings = append(allListings, string(listingResultJSON))
	}

	marketplaceListings := &marketplaceListings{
		marketplaceListings: allListings,
	}
	marketplaceJSON, _ := marketplaceListings.MarshalJSON()

	return string(marketplaceJSON)
}

func (m *Marketplace) BuyNFT(buyer std.Address, compositeTokenID grc721.TokenID) error {
	listingRaw, exists := m.listings[string(compositeTokenID)]
	// listingRaw, exists := m.listings.Get(string(compositeTokenID))
	if !exists {
		return errors.New("listing does not exist")
	}
	listing := listingRaw

	// Use the specific function to transfer FLIP tokens
	TransferFLIPToSeller(buyer, listing.Seller, listing.Price)

	// Use the specific function to transfer the composite NFT
	err := TransferCompositeNFT(listing.Seller, buyer, string(compositeTokenID))
	if err != nil {
		return err
	}

	// Remove the listing after successful transfer
	// todo: maybe use the function RemoveListing(), without the conversion to string of compositeTokenID
	delete(m.listings, string(compositeTokenID))
	// m.listings.Remove(string(compositeTokenID))
	return nil
}
