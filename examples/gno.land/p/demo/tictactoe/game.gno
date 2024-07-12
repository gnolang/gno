package tictactoe

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
)

// this file is @moul's work in #613
// a few changes and bugfixes have been made

type Game struct {
	player1, player2 std.Address
	board            [9]rune // 0=empty, 1=player1, 2=player2
	turnCtr          int
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
		turnCtr:   -1,
	}
	return &g, nil
}

// Partially recover a game
// The game is guaranteed to be legit in terms of number of tiles 1 and 2
// No winning detection is implemented here however
func RecoverGame(player1, player2 std.Address, board string) (*Game, error) {
	g, e := NewGame(player1, player2)
	if e != nil {
		return nil, e
	}
	if len(board) != 9 {
		return nil, ufmt.Errorf("invalid board length: %d", len(board))
	}
	num1, num2 := 0, 0
	runes := [9]rune{}
	for i, c := range board {
		switch c {
		case rune(0), '_', '-':
			runes[i] = rune(0)
		case rune(1), 'O', 'o':
			num1 += 1
			runes[i] = rune(1)
		case rune(2), 'X', 'x':
			num2 += 1
			runes[i] = rune(2)
		default:
			return nil, errors.New("invalid rune")
		}
	}
	if num1 != num2 && num1 != num2+1 {
		return nil, errors.New("invalid number of x and o")
	}
	g.board = runes
	g.turnCtr = num1 + num2
	g.winnerIdx = -1
	return g, nil
}

// start sets turnCtr to 0.
func (g *Game) Start() {
	if g.turnCtr != -1 {
		panic("game already started")
	}
	g.turnCtr = 0
}

func (g *Game) Play(player std.Address, posX, posY int) error {
	if !g.Started() {
		return errors.New("game not started")
	}

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
		return ufmt.Errorf("slot already used (%d, %d)", posX, posY)
	}

	// play
	playerVal := rune(g.turnCtr%2) + 1 // player1=1, player2=2
	g.board[idx] = playerVal

	// check if win
	if g.checkLastMoveWon(posX, posY) {
		g.winnerIdx = g.turnCtr
	}

	// change turn
	g.turnCtr++
	return nil
}

func (g Game) WouldWin(side rune, x, y int) bool {
	idx := xyToIdx(x, y)
	if g.board[idx] != rune(0) {
		panic("tile should be empty")
	}
	// place rune temporarily
	g.board[idx] = side
	b := g.checkLastMoveWon(x, y)
	g.board[idx] = rune(0)
	return b
}

func (g Game) checkLastMoveWon(posX, posY int) bool {
	// assumes the game wasn't won yet, and that the move was already applied.

	// check vertical line
	{
		a := g.At(posX, 0)
		b := g.At(posX, 1)
		c := g.At(posX, 2)
		if a == b && b == c {
			return true
		}
	}

	// check horizontal line
	{
		a := g.At(0, posY)
		b := g.At(1, posY)
		c := g.At(2, posY)
		if a == b && b == c {
			return true
		}
	}

	// diagonals
	{
		tl := g.At(0, 0)
		tr := g.At(0, 2)
		bl := g.At(2, 0)
		br := g.At(2, 2)
		c := g.At(1, 1)
		if posX == posY && tl == c && c == br {
			return true
		}
		if posX+posY == 2 && tr == c && c == bl {
			return true
		}
	}
	return false
}

func (g Game) At(posX, posY int) rune { return g.board[xyToIdx(posX, posY)] }
func (g Game) Winner() std.Address    { return g.PlayerByIndex(g.winnerIdx) }
func (g Game) Turn() std.Address      { return g.PlayerByIndex(g.turnCtr) }
func (g Game) TurnNumber() int        { return g.turnCtr }
func (g Game) IsDraw() bool           { return g.turnCtr > 8 && g.winnerIdx == -1 }
func (g Game) Started() bool          { return g.turnCtr >= 0 }

func (g Game) IsOver() bool {
	// draw
	if g.turnCtr > 8 {
		return true
	}

	// winner
	return g.Winner() != std.Address("")
}

func (g Game) Output() string {
	output := ""

	for y := 2; y >= 0; y-- {
		for x := 0; x < 3; x++ {
			val := g.At(x, y)
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

func (g Game) PlayerByIndex(idx int) std.Address {
	switch idx % 2 {
	case 0:
		return g.player1
	case 1:
		return g.player2
	default:
		return std.Address("")
	}
}

func xyToIdx(x, y int) int { return y*3 + x }
