// add tests for realm operations

package flippando

import (
	"math"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/flippandoserver"
)

func TestStartGame(t *testing.T) {
	Owner := "Alice"
	GameId := "game123"
	tileType := "1"
	boardSize := 4

	gameId, gameStatus, gameBoard := StartGame(Owner, GameId, tileType, boardSize)

	//if g.Owner != Owner {
	//	t.Errorf("Expected owner %s, but got %s", Owner, g.Owner)
	//}

	if gameId != GameId {
		t.Errorf("Expected gameId %s, but got %s", gameId, GameId)
	}

	//if g.tileType != tileType {
	//	t.Errorf("Expected tileType %s, but got %s", tileType, g.tileType)
	//}

	if len(gameBoard) != boardSize {
		t.Errorf("Expected board size %d, but got %d", boardSize, len(gameBoard))
	}

	//expectedGameTilesLength := int(sqrt(float64(boardSize)))
	//f len(g.gameTiles) != expectedGameTilesLength {
	//	t.Errorf("Expected gameTiles length %d, but got %d", expectedGameTilesLength, len(g.gameTiles))
	//}

	//if g.boardSize != boardSize {
	//	t.Errorf("Expected boardSize %d, but got %d", boardSize, g.boardSize)
	//}

	if gameStatus != "initiated" {
		t.Errorf("Expected status 'initiated', but got %s", gameStatus)
	}
}

func TestCreateNFT(t *testing.T) {
    // Setup
    gameId := "testGameID"

    // Execute
    defer func() {
        if r := recover(); r != nil {
            t.Errorf("Function panicked: %v", r)
        }
    }()
    createNFT(gameId)

    // Additional tests for the functions called within createNFT
    t.Run("TestNewBasicNFT", testNewBasicNFT)
    t.Run("TestGenerateColorSquareGrid", testGenerateColorSquareGrid)
    t.Run("TestMarshalJSON", testMarshalJSON)
}

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
