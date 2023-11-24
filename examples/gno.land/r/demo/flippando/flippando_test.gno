package flippando

import (
	"math"
	"std"
	"testing"
	"encoding/json"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/flippandoserver"
)

// todo refactor this so it actually works
func TestStartGame(t *testing.T) {
	player := std.Address("someAddress")
	tileType := "someType"
	boardSize := 16

	expectedGameJSON := `{"id":"000000001","gameBoard":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"solvedGameBoard":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"gameTiles":[1,3,5,8],"gameStatus":"initialized"}`

	got := StartGame(player, tileType, boardSize)
	if got != expectedGameJSON {
		t.Errorf("StartGame() = %v; want %v", got, expectedGameJSON)
	}
}

func TestCreateNFT(t *testing.T) {
    mockPlayer := std.Address("someAddress")
    mockGameId := "someGameID"
    
	CreateBasicNFT(mockPlayer, mockGameId)

	metadata, err := CreateBasicFlipNFT(mockPlayer, mockGameId)
		if err != "" {
			t.Fatalf("Expected no error, got %s", err)
		}
		if metadata == "" {
			t.Error("Expected metadata to be non-empty")
		}

}

/*
func testNewBasicNFT(t *testing.T) {
    name := "TestNFT"
	symbol := "TNFT"

	nft := NewBasicNFT(name, symbol)

	if nft.name != name || nft.symbol != symbol {
		t.Errorf("Expected name %s and symbol %s, got name %s and symbol %s", name, symbol, nft.name, nft.symbol)
	}

	if nft.owners.Size() != 0 || nft.balances.Size() != 0 || nft.tokenApprovals.Size() != 0 ||
		nft.tokenURIs.Size() != 0 || nft.operatorApprovals.Size() != 0 {
		t.Error("Expected all AVL trees to be initialized and empty")
	}
}

func testGenerateColorSquareGrid(t *testing.T) {
    solvedBoard := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	size := 4
	expectedOutput, err := GenerateColorSquareGrid(solvedBoard, size)
	if err != nil {
		t.Errorf("GenerateColorSquareGrid() error = %v", err)
		return
	}
	output, err := GenerateColorSquareGrid(solvedBoard, size)
	if err != nil {
		t.Errorf("GenerateColorSquareGrid() error = %v", err)
		return
	}
	if output != expectedOutput {
		t.Errorf("GenerateColorSquareGrid() = %v; want %v", output, expectedOutput)
	}
}*/

func testMarshalJSON(t *testing.T) {
    fbn := basicNFTMetaData{
		name:     "Flippando - the Game",
		version:  "1.0.0",
		gameId:   "game123",
		gameType: "classic",
		svgData:  "someSVGData",
	}
	expectedJSON := `{"name":"Flippando - the Game","version":"1.0.0","gameId":"game123","gameType":"classic","svgData":"someSVGData"}`
	jsonData, err := fbn.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(jsonData) != expectedJSON {
		t.Errorf("MarshalJSON() = %s; want %s", string(jsonData), expectedJSON)
	}
}
