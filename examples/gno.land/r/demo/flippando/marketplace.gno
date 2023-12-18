package flippando

import (
    "std"
	"strconv"
	"strings"
    "gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721f"
)

type Marketplace struct {
    listings avl.Tree  // CompositeTokenID -> Listing
}

type Listing struct {
    CompositeTokenID grc721f.TokenID   // TokenID of the composite NFT
    BasicTokenIDs    []int64 // Slice of TokenIDs for the basic NFTs contained within the composite NFT
    Seller           std.Address      // Address of the seller
    Price            uint64           // Sale price
}


func NewMarketplace() *Marketplace {
    return &Marketplace{
        listings: avl.Tree{},
    }
}

func (m *Marketplace) ListNFT(compositeTokenID grc721f.TokenID, seller std.Address, price uint64) error {
    // get tokenURI of the composite NFT
	basicTokenIDs := GetChildrenOf(string(compositeTokenID))
	
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
        return error("listing does not exist")
    }

    listing := rawListing.(Listing)

    // Verify that the caller is the seller
    if listing.Seller != seller {
        return error("unauthorized - caller is not the seller")
    }

    // Remove the listing from the marketplace
    m.listings.Remove(string(compositeTokenID))
    return nil
}

func (m *Marketplace) GetListings() string {
    var allListings []string
    m.listings.Iterate("", "", func(key string, value interface{}) bool {
        // Type assert the value to a Listing
        listing, ok := value.(Listing)
        if !ok {
            // Handle the case where the value is not a Listing
            return true // Skip this iteration
        }
		compositeTokenURI := GetCompositeTokenURI(string(listing.CompositeTokenID))

        listingResult := &listingResult{
            compositeTokenID: string(listing.CompositeTokenID),
            tokenURI:		  compositeTokenURI,
            seller:           string(listing.Seller),
            price:            strconv.FormatUint(listing.Price, 10),
        }
        listingResultJSON, _ := listingResult.MarshalJSON()
        allListings = append(allListings, string(listingResultJSON))
        return true // Continue iterating
    })

    marketplaceListings := &marketplaceListings{
        marketplaceListings: allListings,
    }
    marketplaceJSON, _ := marketplaceListings.MarshalJSON()

    return string(marketplaceJSON)
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
    err := TransferCompositeNFT(listing.Seller, buyer, string(compositeTokenID))
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


func parseBTokenIDs(jsonStr string) ([]int64, error) {
    prefix := `"bTokenIDs":[`
    suffix := `]`

    startIndex := strings.Index(jsonStr, prefix)
    if startIndex == -1 {
        return nil, error("bTokenIDs not found")
    }

    startIndex += len(prefix)
    endIndex := strings.Index(jsonStr[startIndex:], suffix)
    if endIndex == -1 {
        return nil, error("malformed JSON")
    }

	endIndex += startIndex

	bTokenIDsStr := jsonStr[startIndex:endIndex]
    return ParseString(bTokenIDsStr)
}
