package flippando

import (
	"strconv"
	"bytes"
	"errors"
	"std"
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
	TileType	string	`json:"-"`
}

type basicNFTMetaData struct {
	name	string `json:"name"`
	version	string	`json:"version"`
	gameId	string`json:"gameId"`
	gameType string `json:"gameType"`
	svgData	string `json:"svgData"`
}

type compositeNFTMetaData struct {
	name	string `json:"name"`
	version	string	`json:"version"`
	tokenIDs []int64  `json:"tokenIDs"`
}

type nftResult struct {
	result	string `json:"result"`
	nftError	string	`json:"-"`
}

type userNFTs struct {
	tokenIDs	[]int64 `json:"tokenIDs"`
	nftError	string	`json:"error"`
}

type inTransit struct {
	tokenId   int
	inTransit bool
}


func (g Game) MarshalJSON() (_ []byte, err error) {
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
	b.WriteString(`"`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
}

func (gb GameBoard) MarshalJSON() (_ []byte, err error) {
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

func (fbn basicNFTMetaData) MarshalJSON() (_ []byte, err error) {
	
	var b bytes.Buffer

	b.WriteByte('{')
	
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
	
	b.WriteString(`"svgData":"`)
	b.WriteString(fbn.svgData) 
	b.WriteString(`"`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
	
}

func (cbn compositeNFTMetaData) MarshalJSON() (_ []byte, err error) {
	
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"name":"`)
	b.WriteString(cbn.name)
	b.WriteString(`",`)

	b.WriteString(`"version":"`)
	b.WriteString(cbn.version)
	b.WriteString(`",`)

	b.WriteString(`"tokenIDs":[`)
	b.WriteString(writeIntSlice(cbn.tokenIDs))
	b.WriteString(`]`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
	
}

func (n nftResult)MarshalJSON() (_ []byte, err error) {
	
	var b bytes.Buffer

	b.WriteByte('{')

	b.WriteString(`"result":"`)
	b.WriteString(n.result)
	b.WriteString(`"`)

	b.WriteByte('}')

	return b.Bytes(), nil

}

func (t userNFTs)MarshalJSON() (_ []byte, err error) {
	
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