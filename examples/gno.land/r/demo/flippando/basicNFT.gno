package flippando

import (
	"strconv"
	"std"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/avl"
)

var (
	flipBasicNFT	= grc721.NewBasicNFT("Flippando Basic NFT", "BFLP")
	basicNFTtokenIDCounter	uint64
	lockedNFTs avl.Tree // tokenId -> flipAmount
	childOf avl.Tree // tokenId -> parent (composite) tokenId
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
	
	count := flipBasicNFT.TokenCount()
	basicNFTtokenIDCounter = count + 1
	tidString := strconv.FormatUint(basicNFTtokenIDCounter, 10)
	tid := grc721.TokenID(tidString)
	mintResultError := flipBasicNFT.Mint(player, tid)
	if mintResultError == nil {
		gameSVGData, errSVG := GenerateGrid(currentGame.SolvedGameBoard, currentGame.TileType)
		if errSVG != nil {
			result = "{failedSafetyCheck: svgDataCorrupted}"
			nftError = "SVG data not generated"
		}
		flipBasicNFTMetaData := &basicNFTMetaData{
			tokenID: tidString,
			name: "Flippando - the Game",
			version: "1.0.0",
			gameId: currentGame.ID,
			gameType: currentGame.TileType,
			gameLevel: strconv.Itoa(len(currentGame.SolvedGameBoard)),
			svgData: gameSVGData,
		}
		flipBasicNFTMetaDataJSON, err := flipBasicNFTMetaData.MarshalJSON()
		if err != nil {
			panic("error in marhasling flipBasicNFTMetaData");
		}
		flipBasicNFT.SetTokenURI(tid, grc721.TokenURI(flipBasicNFTMetaDataJSON))
		// set approval for the realm owner so we can later transfer the NFT
		flipBasicNFT.Approve(std.CurrentRealm().Addr(), tid)
		// mint fungible FLIP token
		MintLocked(player, tidString, uint64(1))
		// add NFT to the lockedNFTs avl.Tree
		var flipAmount uint64
		if len(currentGame.SolvedGameBoard) == 16 {
			flipAmount = 1
		} else if len(currentGame.SolvedGameBoard) == 64 {
			flipAmount = 4
		}

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
	return createNFTResult, nftError
}

func GetBasicTokenURI(tokenID string) string {
	tid := grc721.TokenID(tokenID)
	tokenURI, err := flipBasicNFT.TokenURI(tid)
	if err != nil {
		return "error"
	}
	return string(tokenURI);
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
	err := flipBasicNFT.TransferFrom(from, to, tokenId)
	if err != nil {
		return err
	}
	// remove from locked NFTs
	lockedNFTs.Remove(tid)
	childOf.Set(tid, parentTid)
	return nil
}

// it is used in marketplace
func TransferBasicNFTToBuyer(from, to std.Address, tid grc721.TokenID) error {
	err := flipBasicNFT.TransferFrom(from, to, tid)
	if err != nil {
		return err
	}
	return nil
}

func GetOwnerOfBasicNFT(tid string) std.Address {
	tokenId := grc721.TokenID(tid)
	owner, err := flipBasicNFT.OwnerOf(tokenId)
	if err != nil {
		return "can't get owner"
	}
	return owner
}

func GetUserBasicFlipNFTs(addr std.Address, readyToUse string) string {
	tokenCount := flipBasicNFT.TokenCount()
	tokenURIs := []string{}
	userNFTsError := ""

	for i := uint64(1); i < uint64(tokenCount+1); i++ {		
		tidString := strconv.FormatUint(i, 10)
		owner, err := flipBasicNFT.OwnerOf(grc721.TokenID(tidString))
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
	//tokenCount := flipBasicNFT.TokenCount()
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
	tokenCount := flipBasicNFT.TokenCount()
	tokenURIs := []string{}
	userNFTsError := ""

	// Assuming there's a function to get the token ID by index for the given address
	for i := uint64(1); i < uint64(tokenCount+1); i++ {		
		tidString := strconv.FormatUint(i, 10)
		owner, err := flipBasicNFT.OwnerOf(grc721.TokenID(tidString))
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
	tokenCount := flipBasicNFT.TokenCount()
	lockedNFTs := lockedNFTs.Size()
	totalUnlocked := int(tokenCount) - lockedNFTs
	return int(totalUnlocked)
}

func AirdropBasicFlipNFT(airdropAccount, 
	toPlayer std.Address,
	 airdropName string, 
	 gameType string, 
	 gameLevel string,
	 svgData string) (string, string) {
	// - airdropAccount is the airdrop account, it will receive the lockedNFTs tokens when
	//	someone solves a composite NFT - the initial NFT is  to this account, then transferred
	//	to toPlayer
	// - toPlayer is the user receiving the NFT
	// - generates a GRC721 token with the airdropped board SVG as a tokenURI
	// - returns token metadata as string, to be used in client
	result := ""
	nftError := ""
	
	count := flipBasicNFT.TokenCount()
	basicNFTtokenIDCounter = count + 1
	tidString := strconv.FormatUint(basicNFTtokenIDCounter, 10)
	tid := grc721.TokenID(tidString)
	mintResultError := flipBasicNFT.Mint(airdropAccount, tid)
	if mintResultError == nil {
		flipBasicNFTMetaData := &basicNFTMetaData{
			tokenID: tidString,
			name: "Flippando - the Game",
			version: "1.0.0",
			gameId: "airdrop/hackerville/1", // airdrop name + airdrop edition
			gameType: gameType,
			gameLevel: gameLevel,
			svgData: svgData,
		}
		flipBasicNFTMetaDataJSON, err := flipBasicNFTMetaData.MarshalJSON()
		if err != nil {
			panic("error in marhasling flipBasicNFTMetaData");
		}
		flipBasicNFT.SetTokenURI(tid, grc721.TokenURI(flipBasicNFTMetaDataJSON))
		// set approval for the realm owner so we can later transfer the NFT
		flipBasicNFT.Approve(std.CurrentRealm().Addr(), tid)
		// mint fungible FLIP token
		MintLocked(airdropAccount, tidString, uint64(1))
		// add NFT to the lockedNFTs avl.Tree
		var flipAmount uint64
		if gameLevel == "16" {
			flipAmount = 1
		} else if gameLevel == "64" {
			flipAmount = 4
		}

		lockedNFTs.Set(tidString, flipAmount)
		result = string(flipBasicNFTMetaDataJSON)
	}
	// transfer the minted NFT to toPlayer
	err := flipBasicNFT.TransferFrom(airdropAccount, toPlayer, tid)
	if err != nil {
		panic("error in transferring to toPLayer")
	}
	
	res := &nftResult{
		result: result,
		nftError: nftError,
	}
	
	createNFTResult, err := res.MarshalJSON()
	if(err != nil){
		panic("error in minting")
	}
	return createNFTResult, nftError
}
