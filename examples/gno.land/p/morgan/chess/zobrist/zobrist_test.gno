package zobrist

import (
	"testing"
)

// piece character to internal piece
var p = [256]Piece{
	'P': PiecePawn,
	'R': PieceRook,
	'N': PieceKnight,
	'B': PieceBishop,
	'Q': PieceQueen,
	'K': PieceKing,

	'p': PieceBlack | PiecePawn,
	'r': PieceBlack | PieceRook,
	'n': PieceBlack | PieceKnight,
	'b': PieceBlack | PieceBishop,
	'q': PieceBlack | PieceQueen,
	'k': PieceBlack | PieceKing,
}

// NewBoard returns a Board normally set up at the initial position for standard
// chess.
func NewBoard() Board {
	return Board{
		// row 1
		p['R'], p['N'], p['B'], p['Q'],
		p['K'], p['B'], p['N'], p['R'],
		// row 2
		p['P'], p['P'], p['P'], p['P'],
		p['P'], p['P'], p['P'], p['P'],

		// rows 3, 4, 5, 6
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0,

		// row 7
		p['p'], p['p'], p['p'], p['p'],
		p['p'], p['p'], p['p'], p['p'],
		// row 8
		p['r'], p['n'], p['b'], p['q'],
		p['k'], p['b'], p['n'], p['r'],
	}
}

func TestInitialPosition(t *testing.T) {
	h := Hash(NewBoard(), false, 0, 255)
	if h != InitialPosition {
		t.Fatalf("InitialPosition is invalid: set to %d, should be %d", InitialPosition, h)
	}
}
