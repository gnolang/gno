package flippando

import (
	"strconv"
	"std"
	"errors"
	"strings"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	FlipBasicNFT	= grc721.NewBasicNFT("Flippando Basic NFT", "BFLP")
	basicNFTtokenIDCounter	uint64
	LockedNFTs avl.Tree // tokenId -> flipAmount
	childOf avl.Tree // tokenId -> parent (composite) tokenId
	mintingLock bool
)

func CreateBasicFlipNFT(player std.Address, gameId string) (string, string) {
	// - generates a GRC721 token with the solved board SVG as a tokenURI
	// - returns token metadata as string, to be used in client
	result := ""
	nftError := ""
	
	currentGame, err := GetGameByIDForUser(player, gameId)
	if err != nil {
		result = "{failedSafetyCheck: gameRetrieval}"
		nftError = "{error: no such game}"
		return result, nftError
	}
	hasZeroValues := func(s []int64) bool { for _, v := range s { if v == 0 { return true } }; return false }(currentGame.SolvedGameBoard)
	if(hasZeroValues){
		result = "{failedSafetyCheck: gameCompletion}"
		nftError = "{error: game is not finished}"
	}
	// Check and acquire the lock
    if mintingLock {
        return "", "minting in progress"
    }
    mintingLock = true
    defer func() { mintingLock = false }() // Release lock at the end
	
	count := FlipBasicNFT.TokenCount()
	basicNFTtokenIDCounter = count + 1
	tidString := strconv.FormatUint(basicNFTtokenIDCounter, 10)
	tid := grc721.TokenID(tidString)
	mintResultError := FlipBasicNFT.Mint(player, tid)
	if mintResultError == nil {
		gameSVGData, errSVG := GenerateGrid(currentGame.SolvedGameBoard, currentGame.TileType)
		if errSVG != nil {
			result = "{failedSafetyCheck: svgDataCorrupted}"
			nftError = "SVG data not generated"
		}
		flipBasicNFTMetaData := &BasicNFTMetaData{
			tokenID: tidString,
			name: "Flippando - the Game",
			version: "1.0.0",
			gameId: currentGame.ID,
			gameType: currentGame.TileType,
			gameLevel: strconv.Itoa(len(currentGame.SolvedGameBoard)),
			svgData: gameSVGData,
			airdropName: "",
			airdropParentID: "",
			airdropXPos: "",
			airdropYPos: "",
		}
		flipBasicNFTMetaDataJSON, err := flipBasicNFTMetaData.MarshalJSON()
		if err != nil {
			panic("error in marhasling flipBasicNFTMetaData");
		}
		FlipBasicNFT.SetTokenURI(tid, grc721.TokenURI(flipBasicNFTMetaDataJSON))
		// set approval for the realm owner so we can later transfer the NFT
		FlipBasicNFT.Approve(std.CurrentRealm().Addr(), tid)
		
		// add NFT to the LockedNFTs avl.Tree
		var flipAmount int64
		if len(currentGame.SolvedGameBoard) == 16 {
			flipAmount = 1000 // 1000 uflip = 1 FLIP
		} else if len(currentGame.SolvedGameBoard) == 64 {
			flipAmount = 4000 // 4000 uflip = 4 FLIP
		}
		// mint fungible FLIP token
		MintAndLockFLIP(player, tidString, std.NewCoin("FLIP", flipAmount))
		LockedNFTs.Set(tidString, flipAmount)

		result = string(flipBasicNFTMetaDataJSON)
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

// returns a string version of the TokenURI
func GetBasicTokenURI(tokenID string) string {
	tid := grc721.TokenID(tokenID)
	tokenURI, err := FlipBasicNFT.TokenURI(tid)
	if err != nil {
		return ufmt.Sprintf("error in GetBasicTokenURI in basicNFT, %s", err.Error())
	}
	return string(tokenURI);
}

// basically identical with the other one, but returns BasicNFTMetaData type
func GetTokenMetadata(tokenId string) (BasicNFTMetaData, error) {
	uri, gerr := FlipBasicNFT.TokenURI(grc721.TokenID(tokenId))
	
	var metadata BasicNFTMetaData
	if gerr != nil {
		return metadata, gerr
	}
	metadata, err := ParseBasicNFTMetaData(uri)
	if err != nil {
		return metadata, err
	}
	return metadata, nil
}

func GetParentdOf(tokenId string) string {
	parentTokenId, exists := childOf.Get(tokenId)
	if exists {
		return parentTokenId.(string);
	} 
	return "" 
}

func UnlockAndTransferNFT(from, to std.Address, tid string, parentTid string) error {
	tokenId := grc721.TokenID(tid)
	
	terr := FlipBasicNFT.TransferFrom(from, to, tokenId)
	if terr != nil {
		return terr
	}
	// remove from locked NFTs
	LockedNFTs.Remove(tid)
	childOf.Set(tid, parentTid)
	return nil
}

// it is used in marketplace
func TransferBasicNFTToBuyer(from, to std.Address, tid grc721.TokenID) error {
	var approveErr error
	var transferErr error

    // Check current approval status
    approvedAddress, err := FlipBasicNFT.GetApproved(tid)
    if err != nil {
        approveErr = errors.New(ufmt.Sprintf("Failed to get approval status: %s", err.Error()))
    } 

    if approvedAddress != std.CurrentRealm().Addr() {
        // Approve the marketplace to transfer the token on behalf of the owner
        err = FlipBasicNFT.Approve(std.CurrentRealm().Addr(), tid)
        if err != nil {
            approveErr = errors.New(ufmt.Sprintf("Failed to approve NFT transfer: %s, approvedAddress: %s", err.Error(), approvedAddress))
		}
    }

    // Perform the transfer
    err = FlipBasicNFT.TransferFrom(from, to, tid)
    if err != nil {
        transferErr = errors.New(ufmt.Sprintf("Failed to transfer NFT: %s", err.Error()))
    } 

    // Combine errors if any
    if approveErr != nil && transferErr != nil {
        return errors.New(ufmt.Sprintf("%s; %s", approveErr.Error(), transferErr.Error()))
    } else if approveErr != nil {
        return approveErr
    } else if transferErr != nil {
        return transferErr
    }

    return nil
}

// it is used after minting a composite NFT
func BurnBasicNFTsAfterMinting(tid grc721.TokenID) error {
	err := FlipBasicNFT.Burn(tid)
	if err != nil {
		return err
	}
	return nil
}

func GetOwnerOfBasicNFT(tid string) (std.Address, error) {
	tokenId := grc721.TokenID(tid)
	owner, err := FlipBasicNFT.OwnerOf(tokenId)
	if err != nil {
		return "", err
	}
	return owner, nil
}

func GetUserBasicFlipNFTs(addr std.Address, readyToUse string) string {
	tokenCount := FlipBasicNFT.TokenCount()
	tokenURIs := []string{}
	userNFTsError := ""

	for i := uint64(1); i < uint64(tokenCount+1); i++ {		
		tidString := strconv.FormatUint(i, 10)
		owner, err := FlipBasicNFT.OwnerOf(grc721.TokenID(tidString))
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
		
		// check the owner is the addr
		if owner == addr {
			// if readyToUSe == "yes" we're returning only locked NFTs
			_, exists := LockedNFTs.Get(tidString)
			if readyToUse == "yes" {
				if exists {
					tokenURI := GetBasicTokenURI(tidString)
					tokenURIs = append(tokenURIs, tokenURI)
				}
			} else {
				if !exists {
					tokenURI := GetBasicTokenURI(tidString)
					tokenURIs = append(tokenURIs, tokenURI)
				}
			}
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

func GetBasicNFTsByTokenId(bTokenIDs string) string {
	
	tokenURIs := []string{}
	bTokenIDsArray, _ := ParseString(bTokenIDs)
	
	for i := int64(0); i < int64(len(bTokenIDsArray)); i++ {		
		tidString := strconv.FormatUint(uint64(bTokenIDsArray[i]), 10)
		
		tokenURI := GetBasicTokenURI(tidString)
		tokenURIs = append(tokenURIs, tokenURI)		
	}

	fetchedUserNFTs := &userNFTsResult{
		UserNFTs: tokenURIs,
	}

	userNFTs, err := fetchedUserNFTs.MarshalJSON() 
	if(err != nil){
		panic("error in fetching basic NFTs by tokenIDs")
	}

	return string(userNFTs)
}

func GetAllBasicFlipNFTs(forPlayer std.Address) string {
	tokenCount := FlipBasicNFT.TokenCount()
	tokenURIs := []string{}
	userNFTsError := ""

	for i := uint64(1); i < uint64(tokenCount+1); i++ {		
		tidString := strconv.FormatUint(i, 10)
		owner, err := FlipBasicNFT.OwnerOf(grc721.TokenID(tidString))

		if err != nil {
			userNFTsError += err.Error()
			if err == ErrInvalidTokenId {
				// do we need to do smth else here?
				return "invalid token"
			} 
		}
		
		// return nfts not belonging to forPlayer
		if owner != forPlayer {
			// only locked nfts, unlocked can't be reused
			_, exists := LockedNFTs.Get(tidString)
			
			if exists {
				tokenURI := GetBasicTokenURI(tidString)
				tokenURIs = append(tokenURIs, tokenURI)
			}
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

func GetTotalAvailableBasicNFTs() int {
	return LockedNFTs.Size()
}

func GetAllUnlockedNFTs() int {
	tokenCount := FlipBasicNFT.TokenCount()
	LockedNFTs := LockedNFTs.Size()
	totalUnlocked := int(tokenCount) - LockedNFTs
	return int(totalUnlocked)
}

// util

func GetBasicNFTRealmAddress() std.Address {
	return std.CurrentRealm().Addr()
}


// util
func ParseBasicNFTMetaData(jsonStr string) (BasicNFTMetaData, error) {
	var metadata BasicNFTMetaData

	// Helper function to extract a value from the JSON string by key
	extractValue := func(key string) string {
		startIndex := strings.Index(jsonStr, `"`+key+`":"`)
		if startIndex == -1 {
			return ""
		}
		startIndex += len(key) + 4 // Move to the start of the value

		endIndex := strings.Index(jsonStr[startIndex:], `"`)
		if endIndex == -1 {
			return ""
		}

		return jsonStr[startIndex : startIndex+endIndex]
	}

	metadata.tokenID = extractValue("tokenID")
	metadata.name = extractValue("name")
	metadata.version = extractValue("version")
	metadata.gameId = extractValue("gameId")
	metadata.gameType = extractValue("gameType")
	metadata.gameLevel = extractValue("gameLevel")
	metadata.svgData = extractValue("svgData")
	metadata.airdropName = extractValue("airdropName")
	metadata.airdropParentID = extractValue("airdropParentID")
	metadata.airdropXPos = extractValue("airdropXPos")
	metadata.airdropYPos = extractValue("airdropYPos")

	return metadata, nil
}




