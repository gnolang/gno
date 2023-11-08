package flippando

import (
	"errors"
	"strconv"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc721f"
)

var (
	flipCompositeNFT	= grc721f.NewBasicNFT("Flippando Composite NFT", "CFLP")
	compositeNFTtokenIDCounter	uint64
)

/** 
* generates the composite NFT for the player, with a canvas havong cWidth and cHeight, and tokenIds
* in an array "[1,7,3,8]""
*/
func CreateCompositeNFT(player std.Address, cWidth string, cHeight string, bTokenIDsArray string) (string, string) {
	result := ""
	nftError := ""
	bTokenIDs, perr := ParseString(bTokenIDsArray)

	count := flipCompositeNFT.TokenCount()
	compositeNFTtokenIDCounter = count + 1
	tidString := strconv.FormatUint(basicNFTtokenIDCounter, 10)
	tid := grc721f.TokenID(tidString)
	mintResultError := flipCompositeNFT.Mint(player, tid)
	if mintResultError == nil {
		flipCompositeNFTMetaData := &compositeNFTMetaData{
			tokenID: tidString,
			name: "Flippando - the Art Gallery",
			version: "1.0.0",
			canvasWidth: cWidth,
			canvasHeight: cHeight,
			bTokenIDs: bTokenIDsArray,
		}
		flipCompositeNFTMetaDataJSON, err := flipCompositeNFTMetaData.MarshalJSON()
		if err != nil {
			panic("error in marhasling flipCompositeNFTMetaData");
		}
		for _, t := range bTokenIDs {
			tidString := strconv.FormatInt(t, 10)
			flipBasicOwner := GetOwnerOfBasicNFT(tidString)
			UnlockAndTransferNFT(flipBasicOwner, player, tidString)
			UnlockAndTransferFLIP(flipBasicOwner, tidString)
		}
		// iterate through the array and transfer the nft with tokenID to player
		// unlock the token and transfer it to basic nftOwner FLIP balance
		MintLocked(player, tidString, uint64(1))
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
	return result, nftError
}
/*

func GetUserCompositeFlipNFTs(addr std.Address) string {
	tokenCount := flipCompositeNFT.BalanceOf(addr)
	tokenIDs := make([]int64, 0, tokenCount) 

	// Assuming there's a function to get the token ID by index for the given address
	for i := int64(0); i < tokenCount; i++ {
		owner, err := flipBasicNFT.OwnerOf(i)
		if err != nil {
			if err == ErrInvalidTokenId {
				// do we need to do smth else here?
				continue
			} else {
				// resort to nil
				return nil, err
			}
		}
		
		// If the owner matches the address we're looking for, add the TokenID to the slice.
		if owner == addr {
			tokenIDs = append(tokenIDs, i)
		}
	}

	return tokenIDs
}*/