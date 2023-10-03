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
