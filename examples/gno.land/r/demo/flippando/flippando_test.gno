// add tests for realm operations

package flippando

import (
	"math"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/flippandoserver"
)

func TestStartGame(t *testing.T) {
	player := std.Address("someAddress")
	tileType := "someType"
	boardSize := 16

	expectedGameJSON := `{"id":"000000001","solvedGameBoard":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"gameTiles":[1,3,5,8],"gameStatus":"initialized"}`

	got := StartGame(player, tileType, boardSize)
	if got != expectedGameJSON {
		t.Errorf("StartGame() = %v; want %v", got, expectedGameJSON)
	}
}

func TestCreateNFT(t *testing.T) {
    player := std.Address("someAddress")
    gameId := "someGameID"
    
    // Assume tokenIDCounter is a global variable and its initial value is 0
    tokenIDCounter = 0
    
    createNFT(player, gameId)
    
    // Check the new value of tokenIDCounter, it should be incremented by 1
    if tokenIDCounter != 1 {
        t.Errorf("tokenIDCounter = %v; want 1", tokenIDCounter)
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
}*/

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
}

func testMarshalJSON(t *testing.T) {
    fbn := FlippandoBasicNFT{
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
