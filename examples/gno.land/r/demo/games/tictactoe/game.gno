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

func NewGame(player1, player2 std.Address) (*Game, error) {
	if player1 == player2 {
		return nil, errors.New("cannot fight against self")
	}

	g := Game{
		player1:   player1,
		player2:   player2,
		winnerIdx: -1,
	}
	return &g, nil
}

func (g *Game) Play(player std.Address, posX, posY int) error {
	if g.Turn() != player {
		return errors.New("invalid turn")
	}

	if g.IsOver() {
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
	playerVal := rune(g.turnIdx%2) + 1 // player1=1, player2=2
	g.board[idx] = playerVal

	// check if win
	if g.checkLastMoveWon(posX, posY) {
		g.winnerIdx = g.turnIdx
	}

	// change turn
	g.turnIdx++
	return nil
}

func (g Game) checkLastMoveWon(posX, posY int) bool {
	// assumes the game wasn't won yet, and that's the move was already applied.

	// check vertical line
	{
		a := g.at(posX, 0)
		b := g.at(posX, 1)
		c := g.at(posX, 2)
		if a == b && b == c {
			return true
		}
	}

	// check horizontal line
	{
		a := g.at(0, posY)
		b := g.at(1, posY)
		c := g.at(2, posY)
		if a == b && b == c {
			return true
		}
	}

	// diagonals
	{
		tl := g.at(0, 0)
		tr := g.at(0, 2)
		bl := g.at(2, 0)
		br := g.at(2, 2)
		c := g.at(1, 1)
		if tl == c && c == br {
			return true
		}
		if tr == c && c == bl {
			return true
		}
	}
	return false
}

func (g Game) at(posX, posY int) rune { return g.board[xyToIdx(posX, posY)] }
func (g Game) Winner() std.Address    { return g.playerByIndex(g.winnerIdx) }
func (g Game) Turn() std.Address      { return g.playerByIndex(g.turnIdx) }
func (g Game) IsDraw() bool           { return g.turnIdx >= 8 && g.winnerIdx == 0 }

func (g Game) IsOver() bool {
	// draw
	if g.turnIdx >= 8 {
		return true
	}

	// winner
	return g.Winner() != std.Address("")
}

func (g Game) Output() string {
	output := ""

	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			val := g.at(x, y)
			switch val {
			case 0:
				output += "-"
			case 1:
				output += "O"
			case 2:
				output += "X"
			}
		}
		output += "\n"
	}

	return output
}

func (g Game) playerByIndex(idx int) std.Address {
	switch idx % 2 {
	case 0:
		return g.player1
	case 1:
		return g.player2
	default:
		return std.Address("")
	}
}

func xyToIdx(x, y int) int { return x*3 + y }
