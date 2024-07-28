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
	lockedNFTs avl.Tree // tokenId -> flipAmount
	childOf avl.Tree // tokenId -> parent (composite) tokenId
	// hardcoding the airdrop account and restricting minting to it
	validAirdropAccount std.Address = std.Address("g1mphljw4duhzwqs72hv9grm4k6ufkvts3g7t3t4")
	mintingLock bool
)

func CreateBasicFlipNFT(player std.Address, gameId string) (string, string) {
	// - generates a GRC721 token with the solved board SVG as a tokenURI
	// - returns token metadata as string, to be used in client
	result := ""
	nftError := ""
	
	games := GetUserGames(player)
	// safety checks
	if(len(games) == 0){
		result = "{failedSafetyCheck: userHasGames}"
		nftError = "{error: user has zero games}"
	}
	currentGame := games[0]
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
		//FlipBasicNFT.SetApprovalForAll(std.CurrentRealm().Addr(), true)
		//FlipBasicNFT.SetApprovalForAll(std.PrevRealm().Addr(), true)
		
		// add NFT to the lockedNFTs avl.Tree
		var flipAmount int64
		if len(currentGame.SolvedGameBoard) == 16 {
			flipAmount = 1000 // 1000 uflip = 1 FLIP
		} else if len(currentGame.SolvedGameBoard) == 64 {
			flipAmount = 4000 // 4000 uflip = 4 FLIP
		}
		// mint fungible FLIP token
		MintAndLockFLIP(player, tidString, std.NewCoin("FLIP", flipAmount))
		lockedNFTs.Set(tidString, flipAmount)

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
		return "error"
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
	/*
	err := FlipBasicNFT.Approve(std.CurrentRealm().Addr(), tokenId)
        if err != nil {
            return errors.New(ufmt.Sprintf("Failed to approve NFT transfer: %s ", err.Error()))
		}*/
	terr := FlipBasicNFT.TransferFrom(from, to, tokenId)
	if terr != nil {
		return terr
	}
	// remove from locked NFTs
	lockedNFTs.Remove(tid)
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

func GetOwnerOfBasicNFT(tid string) std.Address {
	tokenId := grc721.TokenID(tid)
	owner, err := FlipBasicNFT.OwnerOf(tokenId)
	if err != nil {
		return "can't get owner"
	}
	return owner
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
			_, exists := lockedNFTs.Get(tidString)
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
	//tokenCount := FlipBasicNFT.TokenCount()
	tokenURIs := []string{}
	bTokenIDsArray, _ := ParseString(bTokenIDs)
	//userNFTsError := ""

	// Assuming there's a function to get the token ID by index for the given address
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

	// Assuming there's a function to get the token ID by index for the given address
	for i := uint64(1); i < uint64(tokenCount+1); i++ {		
		tidString := strconv.FormatUint(i, 10)
		owner, err := FlipBasicNFT.OwnerOf(grc721.TokenID(tidString))
		// debug
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
			_, exists := lockedNFTs.Get(tidString)
			
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
	return lockedNFTs.Size()
}

func GetAllUnlockedNFTs() int {
	tokenCount := FlipBasicNFT.TokenCount()
	lockedNFTs := lockedNFTs.Size()
	totalUnlocked := int(tokenCount) - lockedNFTs
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

// extra
// TODO mve this into a separate package

func AirdropBasicFlipNFT( 
	toPlayer std.Address,
	 airdropName string, 
	 airdropParentID string,
	 airdropXPos string,
	 airdropYPos string,
	 gameType string, 
	 gameLevel string,
	 svgData string) (string, string) {
	
	// - toPlayer is the user receiving the NFT, and the locked FLIP
	// - generates a GRC721 token with the airdropped board SVG as a tokenURI
	// - returns token metadata as string, to be used in client
	result := ""
	nftError := ""
	
	if validAirdropAccount.String() != std.GetOrigCaller().String() {
		return "", "airdrop account not correct"
	}

	count := FlipBasicNFT.TokenCount()
	basicNFTtokenIDCounter = count + 1
	tidString := strconv.FormatUint(basicNFTtokenIDCounter, 10)
	tid := grc721.TokenID(tidString)
	mintResultError := FlipBasicNFT.Mint(toPlayer, tid)
	

	if mintResultError == nil {
		gameId := "airdrop/" + airdropName + "/1"

		flipBasicNFTMetaData := &BasicNFTMetaData{
			tokenID: tidString,
			name: "Flippando - the Game",
			version: "1.0.0",
			gameId: gameId, // airdrop name + airdrop edition
			gameType: gameType,
			gameLevel: gameLevel,
			svgData: svgData,
			airdropName: airdropName,
			airdropParentID: airdropParentID,
			airdropXPos: airdropXPos,
			airdropYPos: airdropYPos,
		}
		flipBasicNFTMetaDataJSON, err := flipBasicNFTMetaData.MarshalJSON()
		if err != nil {
			panic("error in marhasling flipBasicNFTMetaData");
		}
		FlipBasicNFT.SetTokenURI(tid, grc721.TokenURI(flipBasicNFTMetaDataJSON))
		// set approval for the realm owner so we can later transfer the NFT
		FlipBasicNFT.Approve(std.CurrentRealm().Addr(), tid)
		
		// add NFT to the lockedNFTs avl.Tree
		var flipAmount int64
		if gameLevel == "16" {
			flipAmount = 1000
		} else if gameLevel == "64" {
			flipAmount = 4000
		}
		// mint fungible FLIP token
		MintAndLockFLIP(toPlayer, tidString, std.NewCoin("FLIP", flipAmount))
		lockedNFTs.Set(tidString, flipAmount)
		// mint FLIP for the account airdropping, to create liquidity
		merr := Mint(airdropAccount, std.NewCoin("FLIP", flipAmount))
		if merr != nil {
			return "", ufmt.Sprintf("error in calling Mint in airdrop function, %s", merr.Error())
		}
		result = string(flipBasicNFTMetaDataJSON)
	} 

	if mintResultError != nil {
		nftError = ufmt.Sprintf("error in calling Mint in airdrop function, %s", mintResultError.Error())
	}
	
	res := &nftResult{
		result: result,
		nftError: nftError,
	}
	
	createNFTResult, err := res.MarshalJSON()
	if(err != nil){
		panic("error in marshaling result")
	}
	return string(createNFTResult), nftError
}


