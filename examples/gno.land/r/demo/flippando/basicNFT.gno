package flippando

import (
	"errors"
	"strconv"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc721f"
)

var (
	flipBasicNFT	= grc721f.NewBasicNFT("Flippando Basic NFT", "BFLP")
	basicNFTtokenIDCounter	uint64
)

func CreateBasicFlipNFT(player std.Address, gameId string) (string, string) {
	// - generates a GRC721 token with the solved board SVG as a tokenURI
	// - returns tokenID as int64, to be used in lockgin the fungible token in nft2flip
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
		// mint fungible FLIP token
		MintLocked(player, tidString, uint64(1))
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

func GetTokenURI(tokenID string) string {
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

func GetUserBasicFlipNFTs(addr std.Address) string {
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
		
		// If the owner matches the address we're looking for, add the TokenID to the slice.
		if owner == addr {
			tokenURI := GetTokenURI(tidString)
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

func GetBasicNFTsByTokenId(bTokenIDs string) string {
	tokenCount := flipBasicNFT.TokenCount()
	tokenURIs := []string{}
	bTokenIDsArray, perr := ParseString(bTokenIDs)
	userNFTsError := ""

	// Assuming there's a function to get the token ID by index for the given address
	for i := int64(0); i < int64(len(bTokenIDsArray)); i++ {		
		tidString := strconv.FormatUint(uint64(bTokenIDsArray[i]), 10)
		
		tokenURI := GetTokenURI(tidString)
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

func GetAllBasicFlipNFTs(excludePlayer std.Address) string {
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
		
		// If excludePlayer is "" return all nfts
		if len(excludePlayer) == 0 {
			tokenURI := GetTokenURI(tidString)
			tokenURIs = append(tokenURIs, tokenURI)
		} else {
			// if excludePlayer is set, return nfts not belonging to excludePlayer
			if owner != excludePlayer {
				tokenURI := GetTokenURI(tidString)
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