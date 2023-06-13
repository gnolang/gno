package game2048

import (
	"std"

	"gno.land/r/demo/users"
)

func Play() {
	caller := std.GetOrigCaller()

	player, found := players.Get(caller.String())
	if found {
		return
	}

	player = newPlayer(caller)
	if !players.Set(caller.String(), player) {
		panic("cannot set player")
	}
}

func Move(move string) {
	caller := std.GetOrigCaller()

	player, found := players.Get(caller.String())
	if !found {
		panic("player not found")
	}

	switch move {
	case "up":
		player.Board.MoveUp()
	case "down":
		player.Board.MoveDown()
	case "left":
		player.Board.MoveLeft()
	case "right":
		player.Board.MoveRight()
	default:
		panic("invalid move")
	}

	// Calculate score if game is over or win
	if player.Board.IsGameOver() || player.Board.IsWin() {
		player.Score += player.Board.CalculateScore()
	}
}

func Reset() {
	caller := std.GetOrigCaller()

	player, found := players.Get(caller.String())
	if !found {
		panic("player not found")
	}

	player.Board.Reset()
}

func GetScore(user users.AddressOrName) int {
	player, found := players.Get(user.Resolve().String())
	if !found {
		panic("player not found")
	}

	return player.Score
}
