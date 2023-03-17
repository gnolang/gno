package tictactoe

import (
	"errors"
	"std"
)

type Game struct {
	player1, player2 std.Address
	board            [9]rune // 0=empty, 1=player1, 2=player2
	turnIdx          int
	winnerIdx        int
}

func NewGame(player1, player2 std.Address) *Game {
	return &Game{player1: player1, player2: player2}
}

func (g *Game) Play(player std.Address, posX, posY int) error {
	if g.Turn() != player {
		return errors.New("invalid turn")
	}

	if !g.IsOver() {
		return errors.New("game over")
	}

	// are posX and posY valid
	if posX < 0 || posY < 0 || posX > 2 || posY > 2 {
		return errors.New("posX and posY should be 0, 1 or 2")
	}

	// is slot already used?
	idx := xyToIdx(posX, posY)
	if g.board[idx] != 0 {
		return errors.New("slot already used")
	}

	// play
	playerVal := rune(g.turnIdx) + 1 // player1=1, player2=2
	g.board[idx] = playerVal

	// check if win

	// change turn
	g.turnIdx = (g.turnIdx + 1) % 2
	return nil
}

func xyToIdx(x, y int) int {
	return x*3 + y
}

func (g *Game) Winner() std.Address { return g.playerByIndex(g.winnerIdx) }
func (g *Game) IsOver() bool        { return g.Winner() != std.Address("") }
func (g Game) Turn() std.Address    { return g.playerByIndex(g.turnIdx) }

func (g Game) Output() string {
	panic("not implemented")
}

func (g Game) playerByIndex(idx int) std.Address {
	switch idx {
	case 0:
		return g.player1
	case 1:
		return g.player2
	default:
		return std.Address("")
	}
}
