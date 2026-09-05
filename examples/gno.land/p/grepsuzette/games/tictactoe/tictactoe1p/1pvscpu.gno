package tictactoe1p

// a 1P-vs-CPU tictactoe
// extending moul's tictactoe model

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/grepsuzette/games/tictactoe"
)

// 1P-VS-CPU is a tictactoe game
type Game struct {
	*tictactoe.Game
	cpu      std.Address
	Intn     func(n int) int
	pickMove func(g *tictactoe.Game, Intn func(int) int) (x, y int, e error) // AI function
}

// "" is a whole new game, whether CPU goes first is random.
// "---------"` OTOH indicates a game where cpu /declined/ to go first.
// fRand is a user-supplied func returning a random number in [0, n(
// fPickMove is nil (uses averageDifficulty) or a func used for CPU to pick moves
func GameFromRepr(
	text string,
	cpu, human std.Address,
	fRand func(int) int,
	fPickMove func(*tictactoe.Game, func(int) int,
	) (x, y int, e error),
) (*Game, error) {
	addr1, addr2 := decideOrder(cpu, human, text, fRand)

	var g *tictactoe.Game
	var e error
	switch len(text) {
	case 0:
		g, e = tictactoe.NewGame(addr1, addr2)
		g.Start()
	case 9:
		r := []rune(text)
		g, e = tictactoe.RecoverGame(
			addr1,
			addr2,
			string([]rune{
				r[6], r[7], r[8],
				r[3], r[4], r[5],
				r[0], r[1], r[2],
			}),
		)
	default:
		return nil, errors.New("invalid board length")
	}
	if fPickMove == nil {
		fPickMove = averageDifficulty
	}
	return &Game{g, cpu, fRand, fPickMove}, e
}

func (game Game) ToRepr() string {
	return strings.ReplaceAll(game.Output(), "\n", "")
}

func (game Game) IsCpuFirst() bool {
	return game.PlayerByIndex(0) == game.cpu
}

func (game *Game) PlayCPU() (x, y int, e error) {
	switch {
	case game.Turn() != game.cpu:
		return -1, -1, ufmt.Errorf(
			"not my turn (%s), turn is %s's",
			game.cpu.String(), game.Turn().String(),
		)
	case game.IsOver():
		return -1, -1, errors.New("game is over")
	default:
		x, y, _ = game.pickMove(game.Game, game.Intn)
		e = game.Play(game.cpu, x, y)
		return x, y, e
	}
}

// Decide who go first, based on the count of markers.
// No error. A special and important case is the empty board,
// meaning player to go first is random.
// fRand is a user-supplied func returning a random number in [0, n(
func decideOrder(cpu, human std.Address, board string, fRand func(int) int) (addr1, addr2 std.Address) {
	var cpuFirst bool
	if board == "" {
		cpuFirst = fRand(2) == 0
	} else {
		numO := strings.Count(board, "O") + strings.Count(board, "o")
		numX := strings.Count(board, "X") + strings.Count(board, "x")
		cpuFirst = numX != numO
	}
	if cpuFirst {
		return cpu, human
	} else {
		return human, cpu
	}
}
