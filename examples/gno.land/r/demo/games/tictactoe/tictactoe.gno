package tictactoe

import (
	"errors"
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
func Challenge(opponent std.Address) string {
	caller := std.GetOrigCaller()
	g, err := NewGame(caller, opponent)
	if err != nil {
		panic(err)
	}
	// TODO: handle "sent gnots" for bets.

	gameCtr++
	gameID := strconv.Itoa(gameCtr)

	games.Set(gameID, g)
	return gameID
}

// Join joins a previously created game.
//
// Caller should have be the opponent of gameID.
func Join(gameID string) {
	caller := std.GetOrigCaller()

	g, err := getGameByID(gameID)
	if err != nil {
		panic("no such game: " + err.Error())
	}
	// TODO: handle "sent".
	// TODO: determines starting player randomly.
	// TODO: check for already accepted.

	if g.player2 != caller {
		panic("only invited opponent can join the game.")
	}

	g.start()
	games.Set(gameID, g)
}

func getGameByID(id string) (*Game, error) {
	obj, found := games.Get(id)
	if !found {
		return nil, errors.New("game not found.")
	}
	return obj.(*Game), nil
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
		_ = gameID
		// FIXME: continue implementation
		// games.Get()
	}
	return "404"
}
