package flippando

import (
	"errors"
	"strconv"
	"std"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/avl"
)

var (
	flipCompositeNFT	= grc721.NewBasicNFT("Flippando Composite NFT", "CFLP")
	compositeNFTtokenIDCounter	uint64
	children avl.Tree // compositeNFT tokenId -> []int64 basicNFTs tokenIds
)

/** 
* generates the composite NFT for the player, with a canvas having cWidth and cHeight, and tokenIds
* in an array "[1,7,3,8]""
*/
func CreateCompositeFlipNFT(player std.Address, cWidth string, cHeight string, bTokenIDsArray string) (string, string) {
	result := ""
	nftError := ""
	bTokenIDs, _ := ParseString(bTokenIDsArray)

	// add checks for cWidth, cHeight (enforce canvas size based on number of existing basicNFTs)
	// 50 / 2 x 2 - 100 / 3 x 3 - 400 / 4 x 4 - 600 / 5 x 5 - 800 / 6 x 6 - 1000 / 7 x 7 - 1200 / 8 x 8 
	// custom error messages

	totalAvailableBasicNFTs := GetTotalAvailableBasicNFTs()
	
	intWidth, werr := strconv.Atoi(cWidth)
	if werr != nil {
		return "", "not a valid integer representation for canvas width"
	}

	intHeight, herr := strconv.Atoi(cWidth)
	if herr != nil {
		return "", "not a valid integer representation for canvas height"
	}

	if totalAvailableBasicNFTs <= 50 {
		if intHeight != 2 && intWidth != 2 {
			return "", "invalid canvas widht and height for under 50 available basic NFTs"
		}
	} else if totalAvailableBasicNFTs > 50 && totalAvailableBasicNFTs <= 100 {
		if intHeight != 3 && intWidth != 3 {
			return "", "invalid canvas widht and height for under 100 available basic NFTs"
		}
	} else if totalAvailableBasicNFTs > 100 && totalAvailableBasicNFTs <= 400 {
		if intHeight != 4 && intWidth != 4 {
			return "", "invalid canvas widht and height for under 400 available basic NFTs"
		}
	} else if totalAvailableBasicNFTs > 400 && totalAvailableBasicNFTs <= 600 {
		if intHeight != 5 && intWidth != 5 {
			return "", "invalid canvas widht and height for under 600 available basic NFTs"
		}
	} else if totalAvailableBasicNFTs > 600 && totalAvailableBasicNFTs <= 800 {
		if intHeight != 6 && intWidth != 6 {
			return "", "invalid canvas widht and height for under 800 available basic NFTs"
		}
	} else if totalAvailableBasicNFTs > 800 && totalAvailableBasicNFTs <= 1000 {
		if intHeight != 7 && intWidth != 7 {
			return "", "invalid canvas widht and height for under 800 available basic NFTs"
		}
	} else if totalAvailableBasicNFTs > 1000 {
		if intHeight != 8 && intWidth != 8 {
			return "", "invalid canvas widht and height for over 1000 available basic NFTs"
		}
	}

	
	count := flipCompositeNFT.TokenCount()
	compositeNFTtokenIDCounter = count + 1
	tidString := strconv.FormatUint(compositeNFTtokenIDCounter, 10)
	tid := grc721.TokenID(tidString)
	mintResultError := flipCompositeNFT.Mint(player, tid)
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
			panic("error in marshaling flipCompositeNFTMetaData");
		}
		
		// iterate through the array and transfer the nft with tokenID to player
		// unlock the token and transfer it to basic nftOwner FLIP balance
		for _, t := range bTokenIDs {
			basicTidString := strconv.FormatInt(t, 10)
			flipBasicOwner := GetOwnerOfBasicNFT(basicTidString)
			UnlockAndTransferNFT(flipBasicOwner, player, basicTidString, tidString)
			// TODO: account for airdropped tokens, the FLIP in these cases should be sent
			// to the airdrop account
			UnlockAndTransferFLIP(flipBasicOwner, basicTidString)
		}
		// set compositeNFT tokenURI
		flipCompositeNFT.SetTokenURI(tid, grc721.TokenURI(flipCompositeNFTMetaDataJSON))
		children.Set(tidString, bTokenIDs)
		// set approval for the realm owner so we can later transfer the NFT
		flipCompositeNFT.Approve(std.CurrentRealm().Addr(), tid)
		result = string(flipCompositeNFTMetaDataJSON)
	}
	res := &nftResult{
		result: result,
		nftError: nftError,
	}
	
	createNFTResult, err := res.MarshalJSON()
	if(err != nil){
		panic("error in minting")
	}
	return string(createNFTResult), nftError
}

// it is used in marketplace
func TransferCompositeNFT(from, to std.Address, tid string) error {
	tokenId := grc721.TokenID(tid)
	err := flipCompositeNFT.TransferFrom(from, to, tokenId)
	if err != nil {
		return errors.New("transfer composite NFT from error")
	}
	bTokenIDs := GetChildrenOf(tid)
	for _, t := range bTokenIDs {
		basicTokenId := grc721.TokenID(strconv.FormatInt(t, 10))
		TransferBasicNFTToBuyer(from, to, basicTokenId)
	}
	return nil
}

func GetCompositeTokenURI(tokenID string) string {
	tid := grc721.TokenID(tokenID)
	tokenURI, err := flipCompositeNFT.TokenURI(tid)
	if err != nil {
		return "error"
	}
	return string(tokenURI);
}

func GetChildrenOf(tokenID string) []int64 {
	childrenOf, exists := children.Get(tokenID)
	if exists {
		return childrenOf.([]int64);
	}
	return nil
}

func GetUserCompositeFlipNFTs(addr std.Address) string {
	tokenCount := flipCompositeNFT.TokenCount()
	tokenURIs := []string{}
	userNFTsError := ""

	// Assuming there's a function to get the token ID by index for the given address
	for i := uint64(1); i < uint64(tokenCount+1); i++ {		
		tidString := strconv.FormatUint(i, 10)
		owner, err := flipCompositeNFT.OwnerOf(grc721.TokenID(tidString))
		// debug
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