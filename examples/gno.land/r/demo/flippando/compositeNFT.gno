package flippando

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/grc/grc721"
)

var (
	flipCompositeNFT	= grc721.NewBasicNFT("Flippando Composite NFT", "CFLP")
	compositeNFTtokenIDCounter	uint64
)
/*
func CreateCompositeFlipNFT(player std.Address, buildingBlocks []int64) string {
	// - generates a GRC721 token with the buildingBlocks as tokenIDs in metadata
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
		result := "{failedSafetyCheck: gameCompletion}"
		nftError := "{error: game is not finished}"
	}
	gameSVGData, errSVG := GenerateColorSquareGrid(currentGame.SolvedGameBoard, 4)
	if(errSVG != nil){
		result := "{failedSafetyCheck: svgDataCorrupted}"
		nftError := "{error: SVG data not generated}"
	}
	flipBasicNFTMetaData := &basicNFTMetaData{
		name: "Flippando - the Game",
		version: "1.0.0",
		gameId: currentGame.ID,
		gameType: currentGame.TileType,
		svgData: gameSVGData,
	}
	flipBasicNFTMetaDataJSON, err := basicNFTMetaData.MarshalJSON()
	if err != nil {
		panic("error in marhasling flipBasicNFTMetaData");
	}
	
	basicNFTtokenIDCounter = flipBasicNFT.TokenCount() + 1
	tid := grc721.TokenID(ufmt.Sprintf("%d", i))
	mintResultError := flipBasicNFT.SafeMint(player, tid)
	if(mintResultError == nil){
		flipBasicNFT.SetTokenURI(tid, grc721.TokenURI(flippandoBasicNFTJSON))
		result = string(flipBasicNFTMetaDataJSON)
		nftError = "{error: false}"
	}
	res := &nftResult{
		result: result,
		nftError: nftError,
	}
	
	createNFTResult, err := res.MarshalJSON()
	if(err != nil){
		return "error"
	}
	return string(createNFTResult)
}

func GetUserBasicFlipNFTs(addr std.Address) string {
	tokenCount := flipBasicNFT.BalanceOf(addr)
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