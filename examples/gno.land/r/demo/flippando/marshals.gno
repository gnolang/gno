package flippando

import (
	"strconv"
	"bytes"
)

type GameBoard struct {
	Board     []int64  `json:"board"`
	GameTiles []int64  `json:"gameTiles"`
	BoardSize int    `json:"boardSize"`
}

type Game struct {
	ID 			string `json:"id"`
	GameStatus 	string	`json:"gameStatus"`
	GameBoard 	[]int64	`json:"gameBoard"`
	SolvedGameBoard	[]int64	`json:"solvedGameBoard"`
	GameTiles	[]int64	`json:"gameTiles"`
	TileType	string	`json:"tileType"`
}

type GameResult struct {
	ID			string `json:"id"`
	GameBoard 	string `json:"gameBoard"`
	TileType	string `json:"tileTpye"`
}

type UserGamesResult struct {
	UserGames []Game `json:"userGames"`
}

type basicNFTMetaData struct {
	tokenID	string	`json:"tokenID"`
	name	string `json:"name"`
	version	string	`json:"version"`
	gameId	string`json:"gameId"`
	gameType string `json:"gameType"`
	gameLevel string `json:"gameLevel"`
	svgData	string `json:"svgData"`
}

type compositeNFTMetaData struct {
	tokenID string `json:"tokenID"`
	name	string `json:"name"`
	version	string	`json:"version"`
	canvasWidth	string `json:canvasWidth"`
	canvasHeight string `json:canvasHeight"`
	bTokenIDs []int64  `json:"bTokenIDs"`
}

type listingCompositeResult struct {
	compositeTokenID string `json:"tokenID"`
    tokenURI    string `json:"tokenURI"`
    seller           string `json:"seller"`
    price            string `json:price"` 
}

type listingBasicResult struct {
	basicTokenID string `json:"tokenID"`
    tokenURI    string `json:"tokenURI"`
    seller           string `json:"seller"`
    price            string `json:price"` 
}

type marketplaceListings struct {
	marketplaceListings	[]string `json:"marketplaceListings"`
}

type nftResult struct {
	result	string `json:"result"`
	nftError	string	`json:"-"`
}

type buyNFTResult struct {
	flipPaid	string `json:"flipPaid"`
	flipBurned	string `json:"flipBurned"`
	buyerror	string `json:"error"`
}

type userNFTs struct {
	tokenIDs	[]int64 `json:"tokenIDs"`
	nftError	string	`json:"-"`
}

type userNFTsResult struct {
	UserNFTs	[]string `json:"userNFTs"`
}

type userFLIPBalances struct {
	lockedBalance string `json:"lockedBalance"`
	availableBalance string `json:"availableBalance"`
}

type inTransit struct {
	tokenId   int
	inTransit bool
}


func (g Game) MarshalJSON() ([]byte,  error) {
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"id":"`)
	b.WriteString(g.ID)
	b.WriteString(`",`)

	b.WriteString(`"gameBoard":[`)
	b.WriteString(writeIntSlice(g.GameBoard))
	b.WriteString(`],`)

	b.WriteString(`"solvedGameBoard":[`)
	b.WriteString(writeIntSlice(g.SolvedGameBoard))
	b.WriteString(`],`)

	b.WriteString(`"gameTiles":[`)
	b.WriteString(writeIntSlice(g.GameTiles))
	b.WriteString(`],`)
	
	b.WriteString(`"gameStatus":"`)
	b.WriteString(g.GameStatus)
	b.WriteString(`",`)

	b.WriteString(`"tileType":"`)
	b.WriteString(g.TileType)
	b.WriteString(`"`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
}

func (gb GameBoard) MarshalJSON() ([]byte,  error) {
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"BoardSize":`)
	b.WriteString(strconv.Itoa(gb.BoardSize)) // converting int to string
	b.WriteString(`,`)
	
	
	b.WriteString(`"Board":[`)
	b.WriteString(writeIntSlice(gb.Board))
	b.WriteString(`],`)
	
	b.WriteString(`"GameTiles":[`)
	b.WriteString(writeIntSlice(gb.GameTiles))
	b.WriteString(`]`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
}

func (ugr UserGamesResult) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteByte('{')
	b.WriteString(`"userGames":[`)
	for i, game := range ugr.UserGames {
		if i > 0 {
			b.WriteByte(',')
		}
		gameJSON, err := game.MarshalJSON()
		if err != nil {
			return nil, err
		}
		b.Write(gameJSON)
	}
	b.WriteString(`]`)
	b.WriteByte('}')

	return b.Bytes(), nil
}

func (fbn basicNFTMetaData) MarshalJSON() ([]byte,  error) {
	
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"tokenID":"`)
	b.WriteString(fbn.tokenID)
	b.WriteString(`",`)

	b.WriteString(`"name":"`)
	b.WriteString(fbn.name)
	b.WriteString(`",`)

	b.WriteString(`"version":"`)
	b.WriteString(fbn.version)
	b.WriteString(`",`)

	b.WriteString(`"gameId":"`)
	b.WriteString(fbn.gameId)
	b.WriteString(`",`)

	b.WriteString(`"gameType":"`)
	b.WriteString(fbn.gameType)
	b.WriteString(`",`)

	b.WriteString(`"gameLevel":"`)
	b.WriteString(fbn.gameLevel)
	b.WriteString(`",`)
	
	b.WriteString(`"svgData":"`)
	b.WriteString(fbn.svgData) 
	b.WriteString(`"`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
	
}

func (cbn compositeNFTMetaData) MarshalJSON() ([]byte,  error) {
	
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"tokenID":"`)
	b.WriteString(cbn.tokenID)
	b.WriteString(`",`)

	b.WriteString(`"name":"`)
	b.WriteString(cbn.name)
	b.WriteString(`",`)

	b.WriteString(`"version":"`)
	b.WriteString(cbn.version)
	b.WriteString(`",`)

	b.WriteString(`"canvasWidth":"`)
	b.WriteString(cbn.canvasWidth)
	b.WriteString(`",`)

	b.WriteString(`"canvasHeight":"`)
	b.WriteString(cbn.canvasHeight)
	b.WriteString(`",`)
	
	b.WriteString(`"bTokenIDs":[`)
	b.WriteString(writeIntSlice(cbn.bTokenIDs))
	b.WriteString(`]`)

	b.WriteByte('}')

	return b.Bytes(), nil
	
}

func (lrc listingCompositeResult)MarshalJSON() ([]byte,  error){

	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"tokenID":"`)
	b.WriteString(lrc.compositeTokenID)
	b.WriteString(`",`)

	// tokenURI is already a string
	b.WriteString(`"tokenURI":`)
	b.WriteString(lrc.tokenURI)
	b.WriteString(`,`)

	b.WriteString(`"seller":"`)
	b.WriteString(lrc.seller)
	b.WriteString(`",`)

	b.WriteString(`"price":"`)
	b.WriteString(lrc.price)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (lrb listingBasicResult)MarshalJSON() ([]byte,  error){

	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"tokenID":"`)
	b.WriteString(lrb.basicTokenID)
	b.WriteString(`",`)

	// tokenURI is already a string
	b.WriteString(`"tokenURI":`)
	b.WriteString(lrb.tokenURI)
	b.WriteString(`,`)

	b.WriteString(`"seller":"`)
	b.WriteString(lrb.seller)
	b.WriteString(`",`)

	b.WriteString(`"price":"`)
	b.WriteString(lrb.price)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (ml marketplaceListings) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"marketplaceListings":[`)
	for i, listing := range ml.marketplaceListings {
		if i > 0 {
			b.WriteString(`,`)
		}
		b.WriteString(listing)
	}
	b.WriteString(`]}`)

	return b.Bytes(), nil
}

func (n nftResult)MarshalJSON() ([]byte,  error) {
	
	var b bytes.Buffer

	b.WriteByte('{')

	b.WriteString(`"result":`)
	b.WriteString(n.result)
	//b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil

}

func (bn buyNFTResult)MarshalJSON() ([]byte, error){
	
	var b bytes.Buffer

	b.WriteByte('{')

	b.WriteString(`"flipPaid":"`)
	b.WriteString(bn.flipPaid)
	b.WriteString(`",`)

	b.WriteString(`"flipBurned":"`)
	b.WriteString(bn.flipBurned)
	b.WriteString(`",`)

	b.WriteString(`"error":"`)
	b.WriteString(bn.buyerror)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (t userNFTs)MarshalJSON() ([]byte,  error) {
	
	var b bytes.Buffer

	b.WriteByte('{')

	b.WriteString(`"tokenIDs":[`)
	b.WriteString(writeIntSlice(t.tokenIDs))
	b.WriteString(`],`)

	b.WriteString(`"error":"`)
	b.WriteString(t.nftError)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil

}

func (u userFLIPBalances)MarshalJSON() ([]byte,  error) {
	
	var b bytes.Buffer

	b.WriteByte('{')

	b.WriteString(`"lockedBalance":"`)
	b.WriteString(u.lockedBalance)
	b.WriteString(`",`)

	b.WriteString(`"availableBalance":"`)
	b.WriteString(u.availableBalance)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil

}

func (r userNFTsResult) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteString(`{"userNFTs":[`)
	for i, nft := range r.UserNFTs {
		if i > 0 {
			b.WriteString(`,`)
		}
		b.WriteString(nft)
	}
	b.WriteString(`]}`)

	return b.Bytes(), nil
}