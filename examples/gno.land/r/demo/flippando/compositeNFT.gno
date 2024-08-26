package flippando

import (
	"errors"
	"strconv"
	"std"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"
)

var (
	FlipCompositeNFT	= grc721.NewBasicNFT("Flippando Composite NFT", "CFLP")
	compositeNFTtokenIDCounter	uint64
	children map[string]string
)

/** 
* generates the composite NFT for the player, with a canvas having cWidth and cHeight, and tokenIds
* in an array "[1,7,3,8]""
*/
func CreateCompositeFlipNFT(player std.Address, cWidth string, cHeight string, bTokenIDsArray string) (string, string) {
	result := ""
	nftError := ""
	
	bTokenIDs, _ := ParseString(bTokenIDsArray)
	// Initialize the children map if it's nil
	if children == nil {
		children = make(map[string]string)
	}

	// add checks for cWidth, cHeight (enforce canvas size based on number of existing basicNFTs)
	totalAvailableBasicNFTs := GetTotalAvailableBasicNFTs()
	
	intWidth, werr := strconv.Atoi(cWidth)
	if werr != nil {
		return "", "{error: not a valid integer representation for canvas width}"
	}

	intHeight, herr := strconv.Atoi(cWidth)
	if herr != nil {
		return "", "{error: not a valid integer representation for canvas height}"
	}

	if totalAvailableBasicNFTs <= 50 {
		if intHeight != 2 && intWidth != 2 {
			return "", "{error: invalid canvas width and height for under 50 available basic NFTs}"
		}
	} else if totalAvailableBasicNFTs > 50 && totalAvailableBasicNFTs <= 100 {
		if intHeight != 3 && intWidth != 3 {
			return "", "{error: invalid canvas width and height for under 100 available basic NFTs}"
		}
	} else if totalAvailableBasicNFTs > 100 && totalAvailableBasicNFTs <= 400 {
		if intHeight != 4 && intWidth != 4 {
			return "", "{error: invalid canvas width and height for under 400 available basic NFTs}"
		}
	} else if totalAvailableBasicNFTs > 400 && totalAvailableBasicNFTs <= 600 {
		if intHeight != 5 && intWidth != 5 {
			return "", "{error: invalid canvas width and height for under 600 available basic NFTs}"
		}
	} else if totalAvailableBasicNFTs > 600 && totalAvailableBasicNFTs <= 800 {
		if intHeight != 6 && intWidth != 6 {
			return "", "{error: invalid canvas width and height for under 800 available basic NFTs}"
		}
	} else if totalAvailableBasicNFTs > 800 && totalAvailableBasicNFTs <= 1000 {
		if intHeight != 7 && intWidth != 7 {
			return "", "{error: invalid canvas width and height for under 800 available basic NFTs}"
		}
	} else if totalAvailableBasicNFTs > 1000 {
		if intHeight != 8 && intWidth != 8 {
			return "", "{error: invalid canvas width and height for over 1000 available basic NFTs}"
		}
	}

	// check if any of the bTokenIds is listed in the marketplace
	for _, tokenId := range bTokenIDs {
		if err := IsBasicTokenIdListed(strconv.FormatInt(tokenId, 10)); err != nil {
			return "", "{error: one or more tokenIds are already listed}"
		}
	}

	// check if the tokens are all airdropped or all organic
	allAirdrop := true
	allOrganic := true

	var bTokenMetadata []BasicNFTMetaData

	for _, tokenId := range bTokenIDs {
		metadata, err := GetTokenMetadata(strconv.FormatInt(tokenId, 10))
		if err != nil {
			return "", "{error: error retrieving basic token metadata}"
		}
		bTokenMetadata = append(bTokenMetadata, metadata)

		if metadata.airdropName == "" {
			allAirdrop = false
		} else {
			allOrganic = false
		}
	}

	if !(allAirdrop || allOrganic) {
		return "", "{error: cannot mix airdropped and organic NFTs in a composite NFT}"
	}

	// If all tokens are airdropped, enforce the ordering logic
	if allAirdrop {
		expectedXPos := 0
		expectedYPos := 0
		gridSize := 0

		// Determine grid size based on gameLevel
		if bTokenMetadata[0].gameLevel == "16" {
			gridSize = 4
		} else if bTokenMetadata[0].gameLevel == "64" {
			gridSize = 8
		} else {
			return "", "{error: unsupported game level}"
		}

		for _, metadata := range bTokenMetadata {
			if metadata.airdropXPos != strconv.Itoa(expectedXPos) || metadata.airdropYPos != strconv.Itoa(expectedYPos) {
				return "", "{error: invalid XPos and YPos order for the airdropped NFTs}"
			}

			if expectedXPos == gridSize-1 {
				expectedXPos = 0
				expectedYPos++
			} else {
				expectedXPos++
			}
		}
	}

	count := FlipCompositeNFT.TokenCount()
	compositeNFTtokenIDCounter = count + 1
	tidString := strconv.FormatUint(compositeNFTtokenIDCounter, 10)
	tid := grc721.TokenID(tidString)
	mintResultError := FlipCompositeNFT.Mint(player, tid)
	if mintResultError == nil {
		flipCompositeNFTMetaData := &compositeNFTMetaData{
			tokenID: tidString,
			name: "Flippando - the Art Gallery",
			version: "1.0.0",
			canvasWidth: string(cWidth),
			canvasHeight: string(cHeight),
			bTokenIDs: bTokenIDs,
		}
		flipCompositeNFTMetaDataJSON, err := flipCompositeNFTMetaData.MarshalJSON()
		if err != nil {
			return "", "{error: error in marshaling flipCompositeNFTMetaData}";
		}
		
		// iterate through the array and transfer the nft with tokenID to player
		for _, t := range bTokenIDs {
			basicTidString := strconv.FormatInt(t, 10)
			flipBasicOwner := GetOwnerOfBasicNFT(basicTidString)
			uerr := UnlockAndTransferNFT(flipBasicOwner, player, basicTidString, tidString)
			if uerr != nil {
				return "", ufmt.Sprintf("Failed to UnlockAndTransferNFT: %s, for basicTokenId %s, owner by %s ", uerr.Error(), basicTidString, flipBasicOwner.String())
			}
			ReleaseLockedFLIP(flipBasicOwner, basicTidString)
		}
		FlipCompositeNFT.SetTokenURI(tid, grc721.TokenURI(flipCompositeNFTMetaDataJSON))
		children[tidString] = bTokenIDsArray
		FlipCompositeNFT.Approve(std.CurrentRealm().Addr(), tid)
		result = string(flipCompositeNFTMetaDataJSON)
	}
	res := &nftResult{
		result: result,
		nftError: nftError,
	}
	
	createNFTResult, err := res.MarshalJSON()
	if(err != nil){
		return "" , "{error: error in minting}"
	}
	return string(createNFTResult), nftError
}


// it is used in marketplace
func TransferCompositeNFT(from, to std.Address, tid grc721.TokenID) error {
	err := FlipCompositeNFT.TransferFrom(from, to, tid)
	if err != nil {
		return err
	}
	return nil
}

func GetCompositeTokenURI(tokenID string) string {
	tid := grc721.TokenID(tokenID)
	tokenURI, err := FlipCompositeNFT.TokenURI(tid)
	if err != nil {
		return "error"
	}
	return string(tokenURI);
}

func GetOwnerOfCompositeTokenId(tokenId string) string {
	owner, err := FlipCompositeNFT.OwnerOf(grc721.TokenID(tokenId))
	if err != nil {
		return "error"
	}
	return string(owner)
}

func GetChildrenOf(tokenID string) (string, error) {
	childrenOf, exists := children[tokenID]
	if !exists {
		return "", errors.New("can't get children of this tokenId")
	}
	return childrenOf, nil
}

func GetUserCompositeFlipNFTs(addr std.Address) string {
	tokenCount := FlipCompositeNFT.TokenCount()
	tokenURIs := []string{}
	userNFTsError := ""

	for i := uint64(1); i < uint64(tokenCount+1); i++ {		
		tidString := strconv.FormatUint(i, 10)
		owner, err := FlipCompositeNFT.OwnerOf(grc721.TokenID(tidString))
		
		if err != nil {
			userNFTsError += err.Error()
		}
		if err != nil {
			if err == ErrInvalidTokenId {
				// do we need to do smth else here?
				return "invalid token"
			} 
		}
		
		// If the owner matches the address we're looking for, add the TokenID to the slice.
		if owner == addr {
			tokenURI := GetCompositeTokenURI(tidString)
			tokenURIs = append(tokenURIs, tokenURI)
		}
	}

	fetchedUserNFTs := &userNFTsResult{
		UserNFTs: tokenURIs,
	}

	userNFTs, err := fetchedUserNFTs.MarshalJSON() 
	if(err != nil){
		panic("error in fetching user NFTs")
	}

	return string(userNFTs)
}

// util

func GetCompositeNFTRealmAddress() std.Address {
	return std.CurrentRealm().Addr()
}