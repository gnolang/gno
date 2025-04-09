package dice_roller

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
)

var (
	player1       = testutils.TestAddress("alice")
	player2       = testutils.TestAddress("bob")
	unknownPlayer = testutils.TestAddress("unknown")
)

// resetGameState resets the game state for testing
func resetGameState() {
	games = avl.Tree{}
	gameId = seqid.ID(0)
	players = avl.Tree{}
}

// TestNewGame tests the initialization of a new game
func TestNewGame(t *testing.T) {
	resetGameState()

	testing.SetOriginCaller(player1)
	gameID := NewGame(player2)

	// Verify that the game has been correctly initialized
	g, err := getGame(gameID)
	urequire.NoError(t, err)
	urequire.Equal(t, player1.String(), g.player1.String())
	urequire.Equal(t, player2.String(), g.player2.String())
	urequire.Equal(t, 0, g.roll1)
	urequire.Equal(t, 0, g.roll2)
}

// TestPlay tests the dice rolling functionality for both players
func TestPlay(t *testing.T) {
	resetGameState()

	testing.SetOriginCaller(player1)
	gameID := NewGame(player2)

	g, err := getGame(gameID)
	urequire.NoError(t, err)

	// Simulate rolling dice for player 1
	roll1 := Play(gameID)

	// Verify player 1's roll
	urequire.NotEqual(t, 0, g.roll1)
	urequire.Equal(t, g.roll1, roll1)
	urequire.Equal(t, 0, g.roll2) // Player 2 hasn't rolled yet

	// Simulate rolling dice for player 2
	testing.SetOriginCaller(player2)
	roll2 := Play(gameID)

	// Verify player 2's roll
	urequire.NotEqual(t, 0, g.roll2)
	urequire.Equal(t, g.roll1, roll1)
	urequire.Equal(t, g.roll2, roll2)
}

// TestPlayAgainstSelf tests the scenario where a player plays against themselves
func TestPlayAgainstSelf(t *testing.T) {
	resetGameState()

	testing.SetOriginCaller(player1)
	gameID := NewGame(player1)

	// Simulate rolling dice twice by the same player
	roll1 := Play(gameID)
	roll2 := Play(gameID)

	g, err := getGame(gameID)
	urequire.NoError(t, err)
	urequire.Equal(t, g.roll1, roll1)
	urequire.Equal(t, g.roll2, roll2)
}

// TestPlayInvalidPlayer tests the scenario where an invalid player tries to play
func TestPlayInvalidPlayer(t *testing.T) {
	resetGameState()

	testing.SetOriginCaller(player1)
	gameID := NewGame(player1)

	// Attempt to play as an invalid player
	testing.SetOriginCaller(unknownPlayer)
	urequire.PanicsWithMessage(t, "invalid player", func() {
		Play(gameID)
	})
}

// TestPlayAlreadyPlayed tests the scenario where a player tries to play again after already playing
func TestPlayAlreadyPlayed(t *testing.T) {
	resetGameState()

	testing.SetOriginCaller(player1)
	gameID := NewGame(player2)

	// Player 1 rolls
	Play(gameID)

	// Player 1 tries to roll again
	urequire.PanicsWithMessage(t, "already played", func() {
		Play(gameID)
	})
}

// TestPlayBeyondGameEnd tests that playing after both players have finished their rolls fails
func TestPlayBeyondGameEnd(t *testing.T) {
	resetGameState()

	testing.SetOriginCaller(player1)
	gameID := NewGame(player2)

	// Play for both players
	testing.SetOriginCaller(player1)
	Play(gameID)
	testing.SetOriginCaller(player2)
	Play(gameID)

	// Check if the game is over
	_, err := getGame(gameID)
	urequire.NoError(t, err)

	// Attempt to play more should fail
	testing.SetOriginCaller(player1)
	urequire.PanicsWithMessage(t, "game over", func() {
		Play(gameID)
	})
}
