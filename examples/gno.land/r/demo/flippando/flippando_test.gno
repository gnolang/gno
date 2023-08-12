// add tests for realm operations

package flippando

import (
	"math"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestStartGame(t *testing.T) {
	owner := "Alice"
	gameId := "game123"
	tileType := "1"
	boardSize := 4

	g := startGame(owner, gameId, tileType, boardSize)

	if g.owner != owner {
		t.Errorf("Expected owner %s, but got %s", owner, g.owner)
	}

	if g.gameId != gameId {
		t.Errorf("Expected gameId %s, but got %s", gameId, g.gameId)
	}

	if g.tileType != tileType {
		t.Errorf("Expected tileType %s, but got %s", tileType, g.tileType)
	}

	if len(g.board) != boardSize {
		t.Errorf("Expected board size %d, but got %d", boardSize, len(g.board))
	}

	expectedGameTilesLength := int(sqrt(float64(boardSize)))
	if len(g.gameTiles) != expectedGameTilesLength {
		t.Errorf("Expected gameTiles length %d, but got %d", expectedGameTilesLength, len(g.gameTiles))
	}

	if g.boardSize != boardSize {
		t.Errorf("Expected boardSize %d, but got %d", boardSize, g.boardSize)
	}

	if g.status != "initiated" {
		t.Errorf("Expected status 'initiated', but got %s", g.status)
	}
}
