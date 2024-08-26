package flippando

import (
	"math"
	"strconv"
	"std"
	"testing"
	"encoding/json"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/flippandoserver"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/avl"
)

// todo refactor this so it actually works
/*
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

}*/

func testMarshalJSON(t *testing.T) {
    fbn := BasicNFTMetaData{
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



