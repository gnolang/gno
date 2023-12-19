package flippando

import (
	"errors"
	"strconv"
	"strings"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc721f"
	"gno.land/p/demo/avl"
)

var (
	flipCompositeNFT	= grc721f.NewBasicNFT("Flippando Composite NFT", "CFLP")
	compositeNFTtokenIDCounter	uint64
	children avl.Tree // compositeNFT tokenId -> []int64 basicNFTs tokenIds
)

/** 
* generates the composite NFT for the player, with a canvas havong cWidth and cHeight, and tokenIds
* in an array "[1,7,3,8]""
*/
func CreateCompositeFlipNFT(player std.Address, cWidth string, cHeight string, bTokenIDsArray string) (string, string) {
	result := ""
	nftError := ""
	bTokenIDs, perr := ParseString(bTokenIDsArray)

	count := flipCompositeNFT.TokenCount()
	compositeNFTtokenIDCounter = count + 1
	tidString := strconv.FormatUint(compositeNFTtokenIDCounter, 10)
	tid := grc721f.TokenID(tidString)
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
			panic("error in marhasling flipCompositeNFTMetaData");
		}
		
		// iterate through the array and transfer the nft with tokenID to player
		// unlock the token and transfer it to basic nftOwner FLIP balance
		for _, t := range bTokenIDs {
			basicTidString := strconv.FormatInt(t, 10)
			flipBasicOwner := GetOwnerOfBasicNFT(basicTidString)
			UnlockAndTransferNFT(flipBasicOwner, player, basicTidString, tidString)
			UnlockAndTransferFLIP(flipBasicOwner, basicTidString)
		}
		// set compositeNFT tokenURI
		flipCompositeNFT.SetTokenURI(tid, grc721f.TokenURI(flipCompositeNFTMetaDataJSON))
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
	return result, nftError
}

// it is used in marketplace
func TransferCompositeNFT(from, to std.Address, tid string) error {
	tokenId := grc721f.TokenID(tid)
	err := flipCompositeNFT.TransferFrom(from, to, tokenId)
	if err != nil {
		return error("transfer from error")
	}
	bTokenIDs := GetChildrenOf(tid)
	for _, t := range bTokenIDs {
		tokenID := grc721f.TokenID(tid)
		flipBasicNFT.TransferFrom(from, to, tokenID)
	}
	return nil
}

func GetCompositeTokenURI(tokenID string) string {
	tid := grc721f.TokenID(tokenID)
	tokenURI, err := flipCompositeNFT.TokenURI(tid)
	if err != nil {
		return "error"
	}
	return string(tokenURI);
}

func GetChildrenOf(tokenID string) []int64 {
	childrenOf, exists := children.Get(tokenID)
	if exists {
		return []int64(childrenOf);
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
		owner, err := flipCompositeNFT.OwnerOf(grc721f.TokenID(tidString))
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