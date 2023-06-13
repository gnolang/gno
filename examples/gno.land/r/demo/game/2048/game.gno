package game2048

import (
	"std"

	"gno.land/r/demo/users"
)

func Play() {
	caller := std.GetOrigCaller()

	player := getPlayer(caller)
	if player != nil {
		return
	}

	player = newPlayer(caller)
	if !players.Set(caller.String(), player) {
		panic("cannot set player")
	}
}

func Move(move string) {
	caller := std.GetOrigCaller()

	player := getPlayer(caller)
	if player == nil {
		panic("player does not exist")
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

	player := getPlayer(caller)
	if player == nil {
		panic("player does not exist")
	}

	player.Board.Reset()
}

func GetScore(user users.AddressOrName) int {
	player := getPlayer(user.Resolve())
	if player == nil {
		panic("player does not exist")
	}

	return player.Score
}

func getPlayer(addr std.Address) *Player {
	player, found := players.Get(addr.String())
	if !found {
		return nil
	}

	return player.(*Player)
}

func leaderboard() []*Player {
	// TODO: implement
	return nil
}
