package flippando

import (
	"errors"
	"strconv"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc721f"
	"gno.land/p/demo/avl"
)

var (
	flipBasicNFT	= grc721f.NewBasicNFT("Flippando Basic NFT", "BFLP")
	basicNFTtokenIDCounter	uint64
	lockedNFTs avl.Tree // tokenId -> flipAmount
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
	tid := grc721f.TokenID(tidString)
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
			svgData: gameSVGData,
		}
		flipBasicNFTMetaDataJSON, err := flipBasicNFTMetaData.MarshalJSON()
		if err != nil {
			panic("error in marhasling flipBasicNFTMetaData");
		}
		flipBasicNFT.SetTokenURI(tid, grc721f.TokenURI(flipBasicNFTMetaDataJSON))
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
	return result, nftError
}

func GetBasicTokenURI(tokenID string) string {
	tid := grc721f.TokenID(tokenID)
	tokenURI, err := flipBasicNFT.TokenURI(tid)
	if err != nil {
		return "error"
	}
	return string(tokenURI);
}

func UnlockAndTransferNFT(from, to std.Address, tid string) error {
	tokenId := grc721f.TokenID(tid)
	err := flipBasicNFT.TransferFrom(from, to, tokenId)
	if err != nil {
		return err
	}
	// remove from locked NFTs
	lockedNFTs.Remove(tid)
	return nil
}

// it is used in marketplace
func TransferBasicNFTToBuyer(from, to std.Address, tid string) error {
	tokenId := grc721f.TokenID(tid)
	err := flipBasicNFT.TransferFrom(from, to, tokenId)
	if err != nil {
		return err
	}
	return nil
}

func GetOwnerOfBasicNFT(tid string) std.Address {
	tokenId := grc721f.TokenID(tid)
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

	// Assuming there's a function to get the token ID by index for the given address
	for i := uint64(1); i < uint64(tokenCount+1); i++ {		
		tidString := strconv.FormatUint(i, 10)
		owner, err := flipBasicNFT.OwnerOf(grc721f.TokenID(tidString))
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
			lockedNFT, exists := lockedNFTs.Get(tidString)
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
	tokenCount := flipBasicNFT.TokenCount()
	tokenURIs := []string{}
	bTokenIDsArray, perr := ParseString(bTokenIDs)
	userNFTsError := ""

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
		owner, err := flipBasicNFT.OwnerOf(grc721f.TokenID(tidString))
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
			lockedNFT, exists := lockedNFTs.Get(tidString)
			
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