package flippando

import (
	"strconv"
	"std"
	"errors"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	// hardcoding the airdrop account and restricting minting to it
	validAirdropAccount std.Address = std.Address("g1mphljw4duhzwqs72hv9grm4k6ufkvts3g7t3t4")
	basicNFTtokenIDCounter uint64
	mintedNFTs avl.Tree // owner -> []string
)

func AirdropBasicFlipNFT( 
	playerRecipient string,
	airdropName string, 
	airdropParentID string,
	airdropXPos string,
	airdropYPos string,
	gameType string, 
	gameLevel string,
	svgData string) (string, string) {
	
	// - toPlayer is the user receiving the NFT, and the locked FLIP
	toPlayer := std.Address(playerRecipient)
	
	// - generates a GRC721 token with the airdropped board SVG as a tokenURI
	// - returns token metadata as string, to be used in client
	result := ""
	nftError := ""

	// Check if the function is called by the valid airdrop account
	if validAirdropAccount.String() != std.GetOrigCaller().String() {
		return "", "airdrop account not correct"
	}

	// Check if the user has already minted the maximum number of NFTs
	const maxNFTs = 20
	userMintedNFTs, exists := mintedNFTs.Get(toPlayer.String())
	var mintedList []string
	if exists {
		mintedList = userMintedNFTs.([]string)
		if len(mintedList) >= maxNFTs {
			return "", "user has already minted the maximum number of NFTs"
		}
	} else {
		mintedList = []string{}
		mintedNFTs.Set(toPlayer.String(), mintedList)
	}

	// Mint the NFT
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
			gameId: gameId, 
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
			panic("error in marshalling flipBasicNFTMetaData")
		}
		FlipBasicNFT.SetTokenURI(tid, grc721.TokenURI(flipBasicNFTMetaDataJSON))
		FlipBasicNFT.Approve(GetBasicNFTRealmAddress(), tid)

		// Add NFT to the LockedNFTs avl.Tree
		var flipAmount int64
		if gameLevel == "16" {
			flipAmount = 1000
		} else if gameLevel == "64" {
			flipAmount = 4000
		}
		MintAndLockFLIP(toPlayer, tidString, std.NewCoin("FLIP", flipAmount))
		LockedNFTs.Set(tidString, flipAmount)

		// Mint FLIP for the account airdropping, to create liquidity
		merr := Mint(airdropAccount, std.NewCoin("FLIP", flipAmount))
		if merr != nil {
			return "", ufmt.Sprintf("error in calling Mint in airdrop function, %s", merr.Error())
		}

		// Add minted NFT to the user's minted NFTs list with the format "basicTokenId:airdropParentId:airdropXPos:airdropYPos"
		mintedEntry := tidString + ":" + airdropParentID + ":" + airdropXPos + ":" + airdropYPos
		if exists {
			mintedList = append(mintedList, mintedEntry)
			mintedNFTs.Set(toPlayer.String(), mintedList)
		} else {
			mintedNFTs.Set(toPlayer.String(), []string{mintedEntry})
		}

		result = string(flipBasicNFTMetaDataJSON)
	} else {
		nftError = ufmt.Sprintf("error in calling Mint in airdrop function, %s", mintResultError.Error())
	}

	res := &nftResult{
		result: result,
		nftError: nftError,
	}
	
	createNFTResult, err := res.MarshalJSON()
	if err != nil {
		panic("error in marshaling result")
	}
	return string(createNFTResult), nftError
}

// getters, utils

// GetAllMintedAirdropNFTs returns all minted airdrop NFTs in JSON format
func GetAllMintedAirdropNFTs() string {
	var allTokenIDs []string
	iterator := mintedNFTs.Iterator()
	for iterator.HasNext() {
		_, userMintedNFTs := iterator.Next()
		mintedList := userMintedNFTs.([]string)
		allTokenIDs = append(allTokenIDs, mintedList...)
	}

	res := &mintedAirdropNFTs{
		TokenIDs: allTokenIDs,
	}
	
	mintedNFTResult, err := res.MarshalJSON()
	if err != nil {
		panic("error in marshalling result")
	}
	return string(mintedNFTResult)
}

// MarshalJSON marshals the mintedAirdropNFTs struct into JSON
func (man mintedAirdropNFTs) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteByte('{')

	b.WriteString(`"tokenIDs":[`)
	for i, s := range man.TokenIDs {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`"` + s + `"`)
	}
	b.WriteString(`]`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

// mintedAirdropNFTs struct
type mintedAirdropNFTs struct {
	TokenIDs []string `json:"tokenIDs"`
}
