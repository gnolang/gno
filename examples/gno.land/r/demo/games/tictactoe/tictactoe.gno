package tictactoe

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
)

var (
	gameCtr int      // game counter
	games   avl.Tree // int(gameID) -> Game
)

// Challenge creates a new game, and return a gameID.
//
// If ugnot are sent, they will be stored for bet and require the opponent to send the same amount.
func Challenge(opponent std.Address) int {
	gameCtr++

	caller := std.GetOrigCaller()
	game, err := NewGame(caller, opponent)
	if err != nil {
		panic(err)
	}
	// TODO: handle "sent gnots" for bets.

	gameID := strconv.Itoa(gameCtr)
	games.Set(gameID, game)
	return gameCtr
}

// Join joins a previously created game.
//
// Caller should have be the opponent of gameID.
func Join(gameID int) {
	// TODO: handle "sent".
	// TODO: determines starting player randomly.
}

func Render(path string) string {
	path = strings.TrimSpace(path)
	parts := strings.Split(path, "/")
	partN := len(parts)

	switch {
	case partN == 0:
		// TODO: leaderboard
		// TODO: "new game" link
		// TODO: "join challenge" link
		// TODO: vanity metrics
		// TODO: last N active games
		return "home"
	case partN == 2 && parts[0] == "game":
		gameID := parts[1]
		// FIXME: continue implementation
		// games.Get()
	}
	return "404"
}
