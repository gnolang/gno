package flippando

import (
	"errors"
	"strconv"
	"strings"
	"std"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc721"
)

type Marketplace struct {
	listings map[string]Listing // CompositeTokenID -> Listing
	basicListings map[string]BasicListing // BasicTokenID -> Listing
}

type Listing struct {
	CompositeTokenID grc721.TokenID // TokenID of the composite NFT
	BasicTokenIDs    []int64        // Slice of TokenIDs for the basic NFTs contained within the composite NFT
	Seller           std.Address    // Address of the seller
	Price            uint64         // Sale price
}

type BasicListing struct {
	BasicTokenID grc721.TokenID // TokenID of the basic NFT
	Seller           std.Address    // Address of the seller
	Price            uint64         // Sale price
}

func NewMarketplace() *Marketplace {
	return &Marketplace{
		listings: make(map[string]Listing),
		basicListings: make(map[string]BasicListing),
	}
}

func (m *Marketplace) ListCompositeNFT(compositeTokenID grc721.TokenID, seller std.Address, price string) error {
    // approve it for buying
    FlipCompositeNFT.Approve(std.CurrentRealm().Addr(), compositeTokenID)
	// get tokenURI of the composite NFT
    compositeTokenIDstr := string(compositeTokenID)
	basicTokenIDArray, err := GetChildrenOf(compositeTokenIDstr)
    if err != nil {
        return err
    }
    basicTokenIDs, _ := ParseString(basicTokenIDArray)
    // approve all basic NFTs for transferring
    for _, t := range basicTokenIDs {
        basicTokenId := grc721.TokenID(strconv.FormatInt(t, 10))
        FlipBasicNFT.Approve(std.CurrentRealm().Addr(), basicTokenId)
    }
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
	m.listings[compositeTokenIDstr] = newListing
	return nil
}

func (m *Marketplace) ListBasicNFT(basicTokenID grc721.TokenID, seller std.Address, price string) error {
	// approve it for buying
    FlipBasicNFT.Approve(std.CurrentRealm().Addr(), basicTokenID)
    // convert price into uflips, multiply by 1000
    intPrice, ferr := convertPriceStringToUint64(price)
	if ferr != nil {
		return ferr
	}
	// Create a new listing
	newBasicListing := BasicListing{
		BasicTokenID: basicTokenID,
		Seller:           seller,
		Price:            intPrice,
	}

	// Add the new listing to the marketplace
	m.basicListings[string(basicTokenID)] = newBasicListing
	return nil
}

func (m *Marketplace) RemoveCompositeNFTListing(compositeTokenID grc721.TokenID, seller std.Address) error {
	// Retrieve the listing
	rawListing, exists := m.listings[string(compositeTokenID)]
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
	return nil
}

func (m *Marketplace) RemoveBasicNFTListing(basicTokenID grc721.TokenID, seller std.Address) error {
	// Retrieve the listing
	rawListing, exists := m.basicListings[string(basicTokenID)]
	if !exists {
		return errors.New("basic listing does not exist")
	}

	listing := rawListing

	// Verify that the caller is the seller
	if listing.Seller != seller {
		return errors.New("unauthorized - caller is not the seller")
	}

	// Remove the listing from the marketplace
	delete(m.basicListings, string(basicTokenID))
	// move approval to Basic NFT in case the basic NFT is to be used
	// in the assembly of a composite NFT later on
	FlipBasicNFT.Approve(GetBasicNFTRealmAddress(), basicTokenID)
	return nil
}

func (m *Marketplace) GetCompositeNFTListings() string {
	var allListings []string
	for _, value := range m.listings {
		// Type assert the value to a Listing
		listing := value
		compositeTokenURI := GetCompositeTokenURI(string(listing.CompositeTokenID))
		compositeTokenURI = strings.Trim(compositeTokenURI, "\"")

		listingResult := &listingCompositeResult{
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

func (m *Marketplace) GetBasicNFTListings() string {
	var allListings []string
	for _, value := range m.basicListings {
		// Type assert the value to a Listing
		listing := value
		basicTokenURI := GetBasicTokenURI(string(listing.BasicTokenID))
		
		listingBasicResult := &listingBasicResult{
			basicTokenID: string(listing.BasicTokenID),
			tokenURI:         basicTokenURI,
			seller:           string(listing.Seller),
			price:            strconv.FormatUint(listing.Price, 10),
		}
		listingResultJSON, _ := listingBasicResult.MarshalJSON()
		allListings = append(allListings, string(listingResultJSON))
	}

	marketplaceListings := &marketplaceListings{
		marketplaceListings: allListings,
	}
	marketplaceJSON, _ := marketplaceListings.MarshalJSON()

	return string(marketplaceJSON)
}

func (m *Marketplace) IsBasicTokenIdListed(tokenId string) error {
	if _, exists := m.basicListings[tokenId]; exists {
		return errors.New("tokenId is already listed")
	}
	return nil
}

func (m *Marketplace) BuyCompositeNFT(buyer std.Address, compositeTokenID grc721.TokenID) (buyResult string) {
    var buyError error
    
    flipPaid := ""
    flipBurned := ""
    listingRaw, exists := m.listings[string(compositeTokenID)]
    if !exists {
        buyError = ErrMarketplaceListingDoesNotExist
    }
    listing := listingRaw
    

    compositeNFTOwner := GetOwnerOfCompositeTokenId(string(compositeTokenID))
    tcerr := TransferCompositeNFT(std.Address(listing.Seller), buyer, compositeTokenID)
    if tcerr != nil {
        buyError = errors.New(ufmt.Sprintf("TransferCompositeNFT error: seller=%s, buyer=%s, tokenID=%s, owner=%s, error: %s", listing.Seller, buyer, string(compositeTokenID), compositeNFTOwner, tcerr.Error()))
    }

    transferredBasicNFTs := []grc721.TokenID{}

    if buyError == nil {
        for _, t := range listing.BasicTokenIDs {
            basicTokenId := grc721.TokenID(strconv.FormatInt(t, 10))
            ownerBasicNFT := GetOwnerOfBasicNFT(string(basicTokenId))
			//FlipBasicNFT.Approve(std.CurrentRealm().Addr(), basicTokenId)
            berr := FlipBasicNFT.TransferFrom(std.Address(listing.Seller), buyer, basicTokenId)
            if berr != nil {
                buyError = errors.New(ufmt.Sprintf("FlipBasicNFT.TransferFrom error: seller=%s, buyer=%s, basicTokenId=%s, owner=%s, error: %s", listing.Seller, buyer, string(basicTokenId), string(ownerBasicNFT), berr.Error()))
                break
            }
            transferredBasicNFTs = append(transferredBasicNFTs, basicTokenId)
        }

        if buyError != nil {
            reverr := TransferCompositeNFT(buyer, std.Address(listing.Seller), compositeTokenID)
            if reverr != nil {
                buyError = errors.New(ufmt.Sprintf("revert TransferCompositeNFT error: seller=%s, buyer=%s, tokenID=%s, owner=%s, error: %s", listing.Seller, buyer, string(compositeTokenID), compositeNFTOwner, tcerr.Error()))
            }
            // Revert any successfully transferred basic NFTs
            for _, basicTokenId := range transferredBasicNFTs {
				//FlipBasicNFT.Approve(std.CurrentRealm().Addr(), basicTokenId)
                _ = FlipBasicNFT.TransferFrom(buyer, std.Address(listing.Seller), basicTokenId)
            }
        }
    }

    if buyError == nil {
        fPricePaid, fPriceBurned, terr := PayFLIPToSeller(buyer, listing.Seller, std.NewCoin("FLIP", int64(listing.Price)))
        if terr != nil {
            buyError = terr
            //revert the sale on FLIP transfer error, most likely not enough tokens
            reverr := TransferCompositeNFT(buyer, std.Address(listing.Seller), compositeTokenID)
            if reverr != nil {
                buyError = errors.New(ufmt.Sprintf("revert TransferCompositeNFT on FLIP error: seller=%s, buyer=%s, tokenID=%s, owner=%s, error: %s", listing.Seller, buyer, string(compositeTokenID), compositeNFTOwner, tcerr.Error()))
            }
            // Revert any successfully transferred basic NFTs
            for _, basicTokenId := range transferredBasicNFTs {
				//FlipBasicNFT.Approve(std.CurrentRealm().Addr(), basicTokenId)
                _ = FlipBasicNFT.TransferFrom(buyer, std.Address(listing.Seller), basicTokenId)
            }
        }
        flipPaid = fPricePaid
        flipBurned = fPriceBurned
    }

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

func (m *Marketplace) BuyBasicNFT(buyer std.Address, basicTokenID grc721.TokenID) (buyResult string) {
	var buyError error

	flipPaid := ""
	flipBurned := ""
	listingRaw, exists := m.basicListings[string(basicTokenID)]
	if !exists {
		buyError = ErrMarketplaceListingDoesNotExist
	}
	listing := listingRaw

	basicNFTOwner := GetOwnerOfBasicNFT(string(basicTokenID))

	// Pay the price, including random burning
	if buyError == nil {
		fPricePaid, fPriceBurned, terr := PayFLIPToSeller(buyer, listing.Seller, std.NewCoin("FLIP", int64(listing.Price)))
		if terr != nil {
			buyError = terr
		}
		flipPaid = fPricePaid
		flipBurned = fPriceBurned
	}

	// Transfer the basic NFT
	if buyError == nil {
		berr := FlipBasicNFT.TransferFrom(std.Address(listing.Seller), buyer, basicTokenID)
		if berr != nil {
			buyError = errors.New(ufmt.Sprintf("TransferFrom inside BuyBsicNFT error: seller=%s, buyer=%s, basicTokenId=%s, owner=%s, error: %s", listing.Seller, buyer, string(basicTokenID), string(basicNFTOwner), berr.Error()))
		}
	}

	// Transfer locked FLIP to buyer
	if buyError == nil {
		lerr := TransferLockedFLIP(std.Address(listing.Seller), buyer, string(basicTokenID))
		if lerr != nil {
			buyError = errors.New(ufmt.Sprintf("TransferLockedFLIP error: seller=%s, buyer=%s, basicTokenId=%s, owner=%s, error: %s", listing.Seller, buyer, string(basicTokenID), string(basicNFTOwner), lerr.Error()))
		}
	}

	// If no errors, remove the listing
	if buyError == nil {
		delete(m.basicListings, string(basicTokenID))
		// move approval to Basic NFT in case the basic NFT is to be used
		// in the assembly of a composite NFT later on
		FlipBasicNFT.Approve(GetBasicNFTRealmAddress(), basicTokenID)
	}

	

	return prepareBuyResult(flipPaid, flipBurned, buyError)
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

func prepareBuyResult(flipPaid, flipBurned string, buyError error) string {
	var buyErrorString string
	if buyError != nil {
		buyErrorString = buyError.Error()
	} else {
		buyErrorString = ""
	}

	buyNFTResultObject := &buyNFTResult{
		flipPaid:    flipPaid,
		flipBurned:  flipBurned,
		buyerror:    buyErrorString,
	}
	buyResultJSON, merr := buyNFTResultObject.MarshalJSON()

	if merr != nil {
		return errors.New("Error in marshaling buyNFTResult").Error()
	}
	return string(buyResultJSON)
}

