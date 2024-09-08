package diceroller

import (
	"std"
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
	id = seqid.ID(0)
}

// TestNewGame tests the initialization of a new game
func TestNewGame(t *testing.T) {
	resetGameState()

	std.TestSetOrigCaller(player1)
	gameID := NewGame(player2, 3)

	// Verify that the game has been correctly initialized
	g, err := gameByID(gameID)
	urequire.NoError(t, err)
	urequire.Equal(t, player1.String(), g.player1.String())
	urequire.Equal(t, player2.String(), g.player2.String())
	urequire.Equal(t, 0, g.roll1)
	urequire.Equal(t, 0, g.roll2)
	urequire.Equal(t, 0, g.score1)
	urequire.Equal(t, 0, g.score2)
	urequire.Equal(t, 0, g.round)
	urequire.Equal(t, 3, g.maxRounds)
}

// TestPlay tests the dice rolling functionality for both players
func TestPlay(t *testing.T) {
	resetGameState()

	std.TestSetOrigCaller(player1)
	gameID := NewGame(player2, 3)

	g, err := gameByID(gameID)
	urequire.NoError(t, err)

	// Simulate rolling dice for player 1
	roll1 := Play(gameID)

	// Verify player 1's roll and score
	urequire.NotEqual(t, 0, g.roll1)
	urequire.Equal(t, g.score1, roll1)
	urequire.Equal(t, 0, g.roll2) // Player 2 hasn't rolled yet
	urequire.Equal(t, 0, g.round)

	// Simulate rolling dice for player 2
	std.TestSetOrigCaller(player2)
	roll2 := Play(gameID)

	// Verify player 2's roll and score
	urequire.Equal(t, 0, g.roll1)
	urequire.Equal(t, 0, g.roll2)
	urequire.Equal(t, g.score1, roll1)
	urequire.Equal(t, g.score2, roll2)
	urequire.Equal(t, 1, g.round) // Round should be incremented now
}

// TestPlayAgainstSelf tests the scenario where a player plays against themselves
func TestPlayAgainstSelf(t *testing.T) {
	resetGameState()

	std.TestSetOrigCaller(player1)
	gameID := NewGame(player1, 1)

	// Simulate rolling dice twice by the same player
	roll1 := Play(gameID)
	roll2 := Play(gameID)

	g, err := gameByID(gameID)
	urequire.NoError(t, err)
	urequire.Equal(t, 0, g.roll1)
	urequire.Equal(t, g.score1, roll1)
	urequire.Equal(t, 0, g.roll2)
	urequire.Equal(t, g.score2, roll2)
}

// TestPlayInvalidPlayer tests the scenario where an invalid player tries to play
func TestPlayInvalidPlayer(t *testing.T) {
	resetGameState()

	std.TestSetOrigCaller(player1)
	gameID := NewGame(player1, 3)

	// Attempt to play as an invalid player
	std.TestSetOrigCaller(unknownPlayer)
	urequire.PanicsWithMessage(t, "invalid player", func() {
		Play(gameID)
	})
}

// TestPlayAlreadyPlayed tests the scenario where a player tries to play again in the same round
func TestPlayAlreadyPlayed(t *testing.T) {
	resetGameState()

	std.TestSetOrigCaller(player1)
	gameID := NewGame(player2, 3)

	// Player 1 rolls
	std.TestSetOrigCaller(player1)
	Play(gameID)

	// Player 1 tries to roll again
	urequire.PanicsWithMessage(t, "already played", func() {
		Play(gameID)
	})
}

// TestPlayUntilGameEnds tests the scenario where the game is played until it ends
func TestPlayUntilGameEnds(t *testing.T) {
	resetGameState()

	std.TestSetOrigCaller(player1)
	gameID := NewGame(player2, 10) // Set the game to have 10 rounds

	for round := 0; round < 10; round++ {
		// Play each round
		std.TestSetOrigCaller(player1)
		Play(gameID)
		std.TestSetOrigCaller(player2)
		Play(gameID)
	}

	// Check if the game is over
	g, err := gameByID(gameID)
	urequire.NoError(t, err)
	urequire.Equal(t, 10, g.round)
	urequire.NotEqual(t, "Game still in progress", g.winner())

	// Verify the winner is correctly determined
	winner := g.winner()
	if winner == "Player 1 wins!" {
		urequire.True(t, g.score1 > g.score2)
	} else if winner == "Player 2 wins!" {
		urequire.True(t, g.score2 > g.score1)
	} else if winner == "It's a draw!" {
		urequire.True(t, g.score1 == g.score2)
	}
}

// TestPlayBeyondMaxRounds tests that playing beyond the maximum number of rounds fails
func TestPlayBeyondMaxRounds(t *testing.T) {
	resetGameState()

	std.TestSetOrigCaller(player1)
	gameID := NewGame(player2, 1) // Only 1 round

	// Play the single round
	std.TestSetOrigCaller(player1)
	Play(gameID)
	std.TestSetOrigCaller(player2)
	Play(gameID)

	// Check if the game is over
	g, err := gameByID(gameID)
	urequire.NoError(t, err)
	urequire.Equal(t, 1, g.round)
	urequire.NotEqual(t, "Game still in progress", g.winner())

	// Attempt to play more rounds should fail
	std.TestSetOrigCaller(player1)
	urequire.PanicsWithMessage(t, "game over", func() {
		Play(gameID)
	})
}
