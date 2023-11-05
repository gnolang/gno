package flippando

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc721f"
)

var (
	flipBasicNFT	= grc721f.NewBasicNFT("Flippando Basic NFT", "BFLP")
	basicNFTtokenIDCounter	uint64
)

func CreateBasicFlipNFT(player std.Address, gameId string) (int64, string, string) {
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
	gameSVGData, errSVG := GenerateColorSquareGrid(currentGame.SolvedGameBoard, 4)
	if(errSVG != nil){
		result = "{failedSafetyCheck: svgDataCorrupted}"
		nftError = "{error: SVG data not generated}"
	}
	flipBasicNFTMetaData := &basicNFTMetaData{
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
	count := flipBasicNFT.TokenCount()
	basicNFTtokenIDCounter = count + 1
	tid := grc721f.TokenID(ufmt.Sprintf("%d", basicNFTtokenIDCounter))
	mintResultError := flipBasicNFT.SafeMint(player, tid)
	if(mintResultError == nil){
		flipBasicNFT.SetTokenURI(tid, grc721f.TokenURI(flipBasicNFTMetaDataJSON))
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
	return int64(basicNFTtokenIDCounter), result, nftError
}

func GetUserBasicFlipNFTs(addr std.Address) string {
	tokenCount, err := flipBasicNFT.BalanceOf(addr)
	tokenCount := flipBasicNFT.TokenCount()
	userNFTsError := ""

	// Assuming there's a function to get the token ID by index for the given address
	//for i := int64(0); i < int64(tokenCount); i++ {
	for i := int64(0); i < int64(10); i++ {		
		owner, err := flipBasicNFT.OwnerOf(grc721f.TokenID(i))
		if err != nil {
			if err == ErrInvalidTokenId {
				// do we need to do smth else here?
				return "invalid token"
			} 
		}
		// If the owner matches the address we're looking for, add the TokenID to the slice.
		if owner == addr {
			tokenIDs = append(tokenIDs, i)
			userNFTsError = "false"
		}
	}

	fetchedUserNFTs := &userNFTs{
		tokenIDs: tokenIDs,
		nftError: userNFTsError,
	}

	userNFTsResult, err := fetchedUserNFTs.MarshalJSON() 
	if(err != nil){
		panic("error in fetching user NFTs")
	}

	return string(userNFTsResult)
}