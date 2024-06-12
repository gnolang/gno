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
	}
}

func (m *Marketplace) ListNFT(compositeTokenID grc721.TokenID, seller std.Address, price string) error {
	// get tokenURI of the composite NFT
    compositeTokenIDstr := string(compositeTokenID)
	basicTokenIDArray, err := GetChildrenOf(compositeTokenIDstr)
    if err != nil {
        return err
    }
    basicTokenIDs, _ := ParseString(basicTokenIDArray)
    // convert price into uflips, multiply by 1000
    intPrice, ferr := convertPriceStringToUint64(price)
	if ferr != nil {
		return ferr
	}
	// Create a new listing
	newListing := Listing{
		CompositeTokenID: compositeTokenID,
		BasicTokenIDs:    basicTokenIDs,
		Seller:           seller,
		Price:            intPrice,
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

func (m *Marketplace) BuyNFT(buyer std.Address, compositeTokenID grc721.TokenID) (buyResult string) {
	var buyError error
    
    flipPaid := ""
    flipBurned := ""
    listingRaw, exists := m.listings[string(compositeTokenID)]
	// listingRaw, exists := m.listings.Get(string(compositeTokenID))
	if !exists {
		buyError = ErrMarketplaceListingDoesNotExist
	}
	listing := listingRaw
    

    // Use the specific function to transfer the composite NFT
	tcerr := TransferCompositeNFT(listing.Seller, buyer, string(compositeTokenID))
	if tcerr != nil {
		buyError = tcerr
	}

    // transfer basic NFTs to buyer
    for _, t := range listing.BasicTokenIDs {
        basicTokenId := grc721.TokenID(t)
        berr := TransferBasicNFTToBuyer(listing.Seller, buyer, basicTokenId)
        if berr != nil {
            buyError = berr
        }
    }

	// Use the specific function to transfer FLIP tokens
    if tcerr == nil {
        fPricePaid, fPriceBurned, terr := TransferFLIPToSeller(buyer, listing.Seller, std.NewCoin("FLIP", int64(listing.Price)))
        if terr != nil {
            buyError = terr
        }
        flipPaid = fPricePaid
        flipBurned = fPriceBurned
    }

	// Remove the listing after successful transfer
    if buyError == nil {
	    delete(m.listings, string(compositeTokenID))
    }

    var buyErrorString string
    if buyError != nil {
        buyErrorString = buyError.Error()
    } else {
        buyErrorString = ""
    }

    buyNFTResultObject := &buyNFTResult{
        flipPaid: flipPaid,
        flipBurned: flipBurned,
        buyerror: buyErrorString,
    }
    buyResultJSON, merr := buyNFTResultObject.MarshalJSON() 

    if merr != nil {
        buyError = errors.New("Error in marshaling buyNFTResult")
    }
	return string(buyResultJSON)
}


// util

func convertPriceStringToUint64(price string) (uint64, error) {
	// Split the string at the decimal point
	parts := strings.Split(price, ".")
	
	// Parse the integer part
	integerPart, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	
	// Parse the fractional part and ensure it has at least three digits
	var fractionalPart int
	if len(parts) > 1 {
		fractionalPartStr := parts[1]
		if len(fractionalPartStr) > 3 {
			fractionalPartStr = fractionalPartStr[:3]
		}
		fractionalPart, err = strconv.Atoi(fractionalPartStr)
		if err != nil {
			return 0, err
		}
		for len(fractionalPartStr) < 3 {
			fractionalPart *= 10
			fractionalPartStr += "0"
		}
	}
	
	// Combine the integer and fractional parts
	combined := uint64(integerPart * 1000 + fractionalPart)
	return combined, nil
}
