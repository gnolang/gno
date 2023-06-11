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

	return newPlayer(caller)
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

	// Calculate score
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
