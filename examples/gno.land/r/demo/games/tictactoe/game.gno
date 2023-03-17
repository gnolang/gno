package tictactoe

import "std"

type Game struct {
	player1, player2 std.Address
	turn             bool
	board            [9]rune
	gameOver         bool
}

func NewGame(player1, player2 std.Address) *Game {
	return &Game{player1: player1, player2: player2}
}

func (g *Game) Play(player std.Address, posX, poxY int) error {
	// check if turn
	// check if pos already done
	// check if gameover
	// play
	// check if win
	// change turn
	return nil
}

func (g *Game) IsOver() bool {
	return g.gameOver
}

func (g Game) Turn() std.Address {
	if g.turn {
		return g.player1
	}
	return g.player2
}

func (g Game) Output() string {
	panic("not implemented")
}
