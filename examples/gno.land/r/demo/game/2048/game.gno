package game2048

import (
	"std"
)

func Play() Player {
	caller := std.GetOrigCaller()

	player, found := players.Get(caller)
	if found {
		return player
	}

	player = newPlayer(caller)
	if !players.Set(caller, player) {
		panic("cannot set player")
	}

	return player
}

func Move(move string) {
	caller := std.GetOrigCaller()

	player, found := players.Get(caller)
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

	player, found := players.Get(caller)
	if !found {
		panic("player not found")
	}

	player.Board.Reset()
}
