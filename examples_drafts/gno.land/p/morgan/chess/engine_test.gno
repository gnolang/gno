package chess

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func movePromo(from, to string, promo Piece) Move {
	m := move(from, to)
	m.Promotion = promo
	return m
}

func move(from, to string) Move {
	return Move{
		From: SquareFromString(strings.ToLower(from)),
		To:   SquareFromString(strings.ToLower(to)),
	}
}

func TestCheckmate(t *testing.T) {
	fp := unsafeFEN("rn1qkbnr/pbpp1ppp/1p6/4p3/2B1P3/5Q2/PPPP1PPP/RNB1K1NR w KQkq - 0 1")
	m := move("f3", "f7")
	newp, ok := fp.ValidateMove(m)
	if !ok {
		t.Fatal("ValidateMove returned false")
	}
	mr := newp.IsFinished()
	if mr != Checkmate {
		t.Fatalf("expected Checkmate (%d), got %d", Checkmate, mr)
	}
}

func TestCheckmateFromFEN(t *testing.T) {
	fp := unsafeFEN("rn1qkbnr/pbpp1Qpp/1p6/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 1")
	mr := fp.IsFinished()
	if mr != Checkmate {
		t.Fatalf("expected Checkmate (%d), got %d", Checkmate, mr)
	}
}

func TestStalemate(t *testing.T) {
	fp := unsafeFEN("k1K5/8/8/8/8/8/8/1Q6 w - - 0 1")
	m := move("b1", "b6")
	newp, ok := fp.ValidateMove(m)
	if !ok {
		t.Fatal("ValidateMove rejected move")
	}
	mr := newp.IsFinished()
	if mr != Stalemate {
		t.Fatalf("expected Stalemate (%d), got %d", Stalemate, mr)
	}
}

// position shouldn't result in check/stalemate
// because pawn can move http://en.lichess.org/Pc6mJDZN#138
func TestNotMate(t *testing.T) {
	fp := unsafeFEN("8/3P4/8/8/8/7k/7p/7K w - - 2 70")
	m := movePromo("d7", "d8", PieceQueen)
	newp, ok := fp.ValidateMove(m)
	if !ok {
		t.Fatal("ValidateMove returned false")
	}
	mr := newp.IsFinished()
	if mr != NotFinished {
		t.Fatalf("expected NotFinished (%d), got %d", NotFinished, mr)
	}
}

func TestXFoldRepetition(t *testing.T) {
	p := NewPosition()
	loop := [...]Move{
		move("g1", "f3"),
		move("g8", "f6"),
		move("f3", "g1"),
		move("f6", "g8"),
	}
	var valid bool
	for i := 0; i < 5; i++ {
		for j, m := range loop {
			p, valid = p.ValidateMove(m)
			if !valid {
				t.Fatalf("move %s not considered valid", m.String())
			}
			fini := p.IsFinished()
			switch {
			case (i == 3 && j == 3) || i == 4:
				// after the fourth full iteration, it should be marked as "drawn" for 5-fold.
				if fini != Drawn5Fold {
					t.Errorf("i: %d j: %d; expect %d got %d", i, j, Drawn5Fold, fini)
				}
			case (i == 1 && j == 3) || i >= 2:
				// After the second full iteration, IsFinished should mark this as "can 3 fold"
				if fini != Can3Fold {
					t.Errorf("i: %d j: %d; expect %d got %d", i, j, Can3Fold, fini)
				}
			default:
				if fini != NotFinished {
					t.Errorf("i: %d j: %d; expect %d got %d", i, j, NotFinished, fini)
				}
			}
		}
	}
}

func assertMoves(p Position, moves ...Move) Position {
	var valid bool
	for _, move := range moves {
		p, valid = p.ValidateMove(move)
		if !valid {
			panic("invalid move")
		}
	}
	return p
}

func TestXFoldRepetition2(t *testing.T) {
	// Like TestXFoldRepetition, but starts after the initial position.

	p := assertMoves(
		NewPosition(),
		move("f2", "f4"),
		move("c7", "c5"),
	)

	loop := [...]Move{
		move("g1", "f3"),
		move("g8", "f6"),
		move("f3", "g1"),
		move("f6", "g8"),
	}
	var valid bool
	for i := 0; i < 5; i++ {
		for j, m := range loop {
			p, valid = p.ValidateMove(m)
			if !valid {
				t.Fatalf("move %s not considered valid", m.String())
			}
			fini := p.IsFinished()
			switch {
			case i == 4:
				if fini != Drawn5Fold {
					t.Errorf("i: %d j: %d; expect %d got %d", i, j, Drawn5Fold, fini)
				}
			case i >= 2:
				if fini != Can3Fold {
					t.Errorf("i: %d j: %d; expect %d got %d", i, j, Can3Fold, fini)
				}
			default:
				if fini != NotFinished {
					t.Errorf("i: %d j: %d; expect %d got %d", i, j, NotFinished, fini)
				}
			}
		}
	}
}

func TestXMoveRule(t *testing.T) {
	p := NewPosition()

	p.HalfMoveClock = 99
	newp := assertMoves(p, move("g1", "f3"))
	fini := newp.IsFinished()
	if fini != Can50Move {
		t.Errorf("want %d got %d", Can50Move, fini)
	}

	p.HalfMoveClock = 149
	newp = assertMoves(p, move("g1", "f3"))
	fini = newp.IsFinished()
	if fini != Drawn75Move {
		t.Errorf("want %d got %d", Drawn75Move, fini)
	}
}

func TestInsufficientMaterial(t *testing.T) {
	fens := []string{
		"8/2k5/8/8/8/3K4/8/8 w - - 1 1",
		"8/2k5/8/8/8/3K1N2/8/8 w - - 1 1",
		"8/2k5/8/8/8/3K1B2/8/8 w - - 1 1",
		"8/2k5/2b5/8/8/3K1B2/8/8 w - - 1 1",
		"8/2k1n1n1/8/8/8/3K4/8/8 w - - 1 1",
		"8/2k1b3/8/8/8/3K1B2/8/8 w - - 1 1",
	}
	for _, f := range fens {
		pos := unsafeFEN(f)
		o := pos.IsFinished()
		if o != CanInsufficient {
			t.Errorf("fen %q: want %d got %d", f, CanInsufficient, o)
		}
	}
}

func TestSufficientMaterial(t *testing.T) {
	fens := []string{
		"8/2k5/8/8/8/3K1B2/4N3/8 w - - 1 1",
		"8/2k5/8/8/8/3KBB2/8/8 w - - 1 1",
		"8/2k5/8/8/4P3/3K4/8/8 w - - 1 1",
		"8/2k5/8/8/8/3KQ3/8/8 w - - 1 1",
		"8/2k5/8/8/8/3KR3/8/8 w - - 1 1",
	}
	for _, f := range fens {
		pos := unsafeFEN(f)
		o := pos.IsFinished()
		if o != NotFinished {
			t.Errorf("fen %q: want %d got %d", f, NotFinished, o)
		}
	}
}

type moveTest struct {
	pos     unsafeFENRes
	m       Move
	postPos unsafeFENRes
}

var (
	invalidMoves = []moveTest{
		// out of turn moves
		{m: move("E7", "E5"), pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: move("E2", "E4"), pos: unsafeFEN("rnbqkbnr/1ppppppp/p7/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")},
		// pawn moves
		{m: move("E2", "D3"), pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: move("E2", "F3"), pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: move("E2", "E5"), pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")},
		{m: move("A2", "A1"), pos: unsafeFEN("8/8/8/8/8/8/p7/8 b - - 0 1")},
		{m: move("E6", "E5"), pos: unsafeFEN(`2b1r3/2k2p1B/p2np3/4B3/8/5N2/PP1K1PPP/3R4 b - - 2 1`)},
		{m: move("H7", "H5"), pos: unsafeFEN(`2bqkbnr/rpppp2p/2n2p2/p5pB/5P2/4P3/PPPP2PP/RNBQK1NR b KQk - 4 6`)},
		// knight moves
		{m: move("E4", "F2"), pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		{m: move("E4", "F3"), pos: unsafeFEN("8/8/8/3pp3/4N3/8/5B2/8 w - - 0 1")},
		// bishop moves
		{m: move("E4", "C6"), pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		{m: move("E4", "E5"), pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		{m: move("E4", "E4"), pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		{m: move("E4", "F3"), pos: unsafeFEN("8/8/8/3pp3/4B3/5N2/8/8 w - - 0 1")},
		// rook moves
		{m: move("B2", "B1"), pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: move("B2", "C3"), pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: move("B2", "B8"), pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		{m: move("B2", "G7"), pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1R6/1B6 w - - 0 1")},
		// queen moves
		{m: move("B2", "B1"), pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: move("B2", "C4"), pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: move("B2", "B8"), pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		{m: move("B2", "G7"), pos: unsafeFEN("8/1p5b/4N3/4p3/8/8/1Q6/1B6 w - - 0 1")},
		// king moves
		{m: move("E4", "F3"), pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: move("E4", "F4"), pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		{m: move("E4", "F5"), pos: unsafeFEN("5r2/8/8/8/4K3/8/8/8 w - - 0 1")},
		// castleing
		{m: move("E1", "B1"), pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: move("E8", "B8"), pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1")},
		{m: move("E1", "C1"), pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R2QK2R w KQkq - 0 1")},
		{m: move("E1", "C1"), pos: unsafeFEN("2r1k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: move("E1", "C1"), pos: unsafeFEN("3rk2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")},
		{m: move("E1", "G1"), pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w Qkq - 0 1")},
		{m: move("E1", "C1"), pos: unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w Kkq - 0 1")},
		// invalid promotion for non-pawn move
		{m: movePromo("B8", "D7", PieceQueen), pos: unsafeFEN("rn1qkb1r/pp3ppp/2p1pn2/3p4/2PP4/2NQPN2/PP3PPP/R1B1K2R b KQkq - 0 7")},
		// en passant on doubled pawn file http://en.lichess.org/TnRtrHxf#24
		{m: move("E3", "F6"), pos: unsafeFEN("r1b2rk1/pp2b1pp/1qn1p3/3pPp2/1P1P4/P2BPN2/6PP/RN1Q1RK1 w - f6 0 13")},
		// can't move piece out of pin (even if checking enemy king) http://en.lichess.org/JCRBhXH7#62
		{m: move("E1", "E7"), pos: unsafeFEN("4R3/1r1k2pp/p1p5/1pP5/8/8/1PP3PP/2K1Rr2 w - - 5 32")},
		// invalid one up pawn capture
		{m: move("E6", "E5"), pos: unsafeFEN(`2b1r3/2k2p1B/p2np3/4B3/8/5N2/PP1K1PPP/3R4 b - - 2 1`)},
		// invalid two up pawn capture
		{m: move("H7", "H5"), pos: unsafeFEN(`2bqkbnr/rpppp2p/2n2p2/p5pB/5P2/4P3/PPPP2PP/RNBQK1NR b KQk - 4 6`)},
		// invalid pawn move d5e4
		{m: move("D5", "E4"), pos: unsafeFEN(`rnbqkbnr/pp2pppp/8/2pp4/3P4/4PN2/PPP2PPP/RNBQKB1R b KQkq - 0 3`)},
	}

	positionUpdates = []moveTest{
		{
			m:       move("E2", "E4"),
			pos:     unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"),
			postPos: unsafeFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"),
		},
		{
			m:       move("E1", "G1"),
			pos:     unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"),
			postPos: unsafeFEN("r3k2r/8/8/8/8/8/8/R4RK1 b kq - 1 1"),
		},
		{
			m:       move("A4", "B3"),
			pos:     unsafeFEN("2r3k1/1q1nbppp/r3p3/3pP3/pPpP4/P1Q2N2/2RN1PPP/2R4K b - b3 0 23"),
			postPos: unsafeFEN("2r3k1/1q1nbppp/r3p3/3pP3/11pP4/PpQ2N2/2RN1PPP/2R4K w - - 0 24"),
		},
		{
			m:       move("E1", "G1"),
			pos:     unsafeFEN("r2qk2r/pp1n1ppp/2pbpn2/3p4/2PP4/1PNQPN2/P4PPP/R1B1K2R w KQkq - 1 9"),
			postPos: unsafeFEN("r2qk2r/pp1n1ppp/2pbpn2/3p4/2PP4/1PNQPN2/P4PPP/R1B2RK1 b kq - 2 9"),
		},
		// half move clock - knight move to f3 from starting position
		{
			m:       move("G1", "F3"),
			pos:     unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"),
			postPos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/5N2/PPPPPPPP/RNBQKB1R b KQkq - 1 1"),
		},
		// half move clock - king side castle
		{
			m:       move("E1", "G1"),
			pos:     unsafeFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"),
			postPos: unsafeFEN("r3k2r/8/8/8/8/8/8/R4RK1 b kq - 1 1"),
		},
		// half move clock - queen side castle
		{
			m:       move("E1", "C1"),
			pos:     unsafeFEN("r3k2r/ppqn1ppp/2pbpn2/3p4/2PP4/1PNQPN2/P2B1PPP/R3K2R w KQkq - 3 10"),
			postPos: unsafeFEN("r3k2r/ppqn1ppp/2pbpn2/3p4/2PP4/1PNQPN2/P2B1PPP/2KR3R b kq - 4 10"),
		},
		// half move clock - pawn push
		{
			m:       move("E2", "E4"),
			pos:     unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"),
			postPos: unsafeFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"),
		},
		// half move clock - pawn capture
		{
			m:       move("E4", "D5"),
			pos:     unsafeFEN("r1bqkbnr/ppp1pppp/2n5/3p4/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"),
			postPos: unsafeFEN("r1bqkbnr/ppp1pppp/2n5/3P4/8/5N2/PPPP1PPP/RNBQKB1R b KQkq - 0 3"),
		},
		// half move clock - en passant
		{
			m:       move("E5", "F6"),
			pos:     unsafeFEN("r1bqkbnr/ppp1p1pp/2n5/3pPp2/8/5N2/PPPP1PPP/RNBQKB1R w KQkq f6 0 4"),
			postPos: unsafeFEN("r1bqkbnr/ppp1p1pp/2n2P2/3p4/8/5N2/PPPP1PPP/RNBQKB1R b KQkq - 0 4"),
		},
		// half move clock - piece captured by knight
		{
			m:       move("C6", "D4"),
			pos:     unsafeFEN("r1bqkbnr/ppp1p1pp/2n5/3pPp2/3N4/8/PPPP1PPP/RNBQKB1R b KQkq - 1 4"),
			postPos: unsafeFEN("r1bqkbnr/ppp1p1pp/8/3pPp2/3n4/8/PPPP1PPP/RNBQKB1R w KQkq - 0 5"),
		},
	}
)

func TestInvalidMoves(t *testing.T) {
	for _, mt := range invalidMoves {
		_, ok := mt.pos.ValidateMove(mt.m)
		if ok {
			t.Errorf("fen %q: unexpected valid move (%s)", mt.pos.orig, mt.m.String())
		}
	}
}

func TestPositionUpdates(t *testing.T) {
	for _, mt := range positionUpdates {
		np, ok := mt.pos.ValidateMove(mt.m)
		if !ok {
			t.Errorf("fen %q: rejected valid move (%s)", mt.pos.orig, mt.m.String())
			continue
		}
		if np.B != mt.postPos.B {
			t.Errorf("%q: boards don't match", mt.pos.orig)
		}
		if np.HalfMoveClock != mt.postPos.HalfMoveClock {
			t.Errorf("%q: hmc doesn't match; want %d got %d", mt.pos.orig, mt.postPos.HalfMoveClock, np.HalfMoveClock)
		}
		if np.HalfMoveClock == 0 && len(np.Hashes) != 1 {
			t.Errorf("%q: hashes not reset", mt.pos.orig)
		}
		if np.Flags != mt.postPos.Flags {
			t.Errorf("%q: flags don't match; want %d got %d", mt.pos.orig, mt.postPos.Flags, np.Flags)
		}
	}
}

func TestPerft(t *testing.T) {
	moves := make([]Move, 0, 10)
	for n, res := range perfResults {
		t.Run(fmt.Sprintf("n%d", n), func(t *testing.T) {
			if testing.Short() {
				t.Skip("skipping perft in short tests")
			}
			res.pos.Moves = append(moves[:0], res.pos.Moves...)
			counts := make([]int, len(res.nodesPerDepth))
			CountMoves(res.pos.Position, len(res.nodesPerDepth), counts)
			t.Logf("counts: %v", counts)
			if !intsMatch(counts, res.nodesPerDepth) {
				t.Errorf("counts don't match: got %v want %v", counts, res.nodesPerDepth)
			}
		})
	}
}

func intsMatch(xx, yy []int) bool {
	if len(xx) != len(yy) {
		return false
	}
	for i := range xx {
		if xx[i] != yy[i] {
			return false
		}
	}
	return true
}

const perftDebug = false

func CountMoves(p Position, depth int, counts []int) {
	total := 0
	l := len(counts) - depth
	p.GenMoves(func(newp Position, m Move) error {
		counts[l]++
		if depth > 1 {
			countMoves(newp, depth-1, counts)
		}
		delta := counts[len(counts)-1] - total
		if perftDebug {
			fmt.Printf("%s%s: %d\n", m.From.String(), m.To.String(), delta)
		}
		total += delta
		return nil
	})
}

func countMoves(p Position, depth int, counts []int) {
	l := len(counts) - depth
	p.GenMoves(func(newp Position, m Move) error {
		counts[l]++
		if depth > 1 {
			countMoves(newp, depth-1, counts)
		}
		return nil
	})
}

type perfTest struct {
	pos           unsafeFENRes
	nodesPerDepth []int
}

/* https://www.chessprogramming.org/Perft_Results */
var perfResults = []perfTest{
	{pos: unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"), nodesPerDepth: []int{
		20, 400, 8902, // 197281,
		// 4865609, 119060324, 3195901860, 84998978956, 2439530234167, 69352859712417
	}},
	{pos: unsafeFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"), nodesPerDepth: []int{
		48, 2039, 97862,
		// 4085603, 193690690
	}},
	{pos: unsafeFEN("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"), nodesPerDepth: []int{
		14, 191, 2812, 43238, // 674624,
		// 11030083, 178633661
	}},
	{pos: unsafeFEN("r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1"), nodesPerDepth: []int{
		6, 264, 9467, // 422333,
		// 15833292, 706045033
	}},
	{pos: unsafeFEN("r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1"), nodesPerDepth: []int{
		6, 264, 9467, // 422333,
		// 15833292, 706045033
	}},
	{pos: unsafeFEN("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8"), nodesPerDepth: []int{
		44, 1486, 62379,
		// 2103487, 89941194
	}},
	{pos: unsafeFEN("r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10"), nodesPerDepth: []int{
		46, 2079, // 89890,
		// 3894594, 164075551, 6923051137, 287188994746, 11923589843526, 490154852788714
	}},
}

// ---
// testing utility functions

// FEN decoding: see https://www.chessprogramming.org/Forsyth-Edwards_Notation
// copied mostly from notnil/chess and adapted to our own system.

type unsafeFENRes struct {
	Position
	orig string
}

func unsafeFEN(fen string) unsafeFENRes {
	p, e := decodeFEN(fen)
	if e != nil {
		panic(e)
	}
	return unsafeFENRes{p, fen}
}

// Decodes FEN into Board and previous moves.
func decodeFEN(fen string) (p Position, err error) {
	fen = strings.TrimSpace(fen)
	parts := strings.Split(fen, " ")
	if len(parts) != 6 {
		err = fmt.Errorf("chess: fen invalid notation %s must have 6 sections", fen)
		return
	}

	p = NewPosition()

	// fen board
	var ok bool
	p.B, ok = fenBoard(parts[0])
	if !ok {
		err = fmt.Errorf("chess: invalid fen board %s", parts[0])
		return
	}

	// do castling rights first (more convenient to set prev)
	if parts[2] != "KQkq" {
		p.Flags = castleRightsToPositionFlags(parts[2])
	}

	// color to play
	color := Color(parts[1] == "b")
	if color == Black {
		// add fake move to make len(prev) odd
		p.Moves = append(p.Moves, Move{})
	}

	// en passant
	if parts[3] != "-" {
		f, e := parseEnPassant(parts[3])
		if e != nil {
			err = e
			return
		}
		p.Flags |= f
	}

	halfMove, _ := strconv.Atoi(parts[4])
	p.HalfMoveClock = uint16(halfMove)

	// parts[5]: full move counter, probably never implementing

	return
}

// generates board from fen format: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
func fenBoard(boardStr string) (Board, bool) {
	rankStrs := strings.Split(boardStr, "/")
	if len(rankStrs) != 8 {
		return Board{}, false
	}
	var b Board
	for idx, pieces := range rankStrs {
		rank := (7 - Square(idx)) << 3
		file := Square(0)
		for _, ch := range pieces {
			if ch >= '1' && ch <= '8' {
				delta := byte(ch) - '0'
				file += Square(delta)
				if file > 8 {
					return b, false
				}
				continue
			}
			piece := p[byte(ch)]
			if piece == PieceEmpty || file >= 8 {
				return b, false
			}
			b[rank|file] = piece
			file++
		}
		if file != 8 {
			return b, false
		}
	}
	return b, true
}

func castleRightsToPositionFlags(cr string) (pf PositionFlags) {
	pf = NoCastleWQ | NoCastleWK | NoCastleBQ | NoCastleBK
	if cr == "-" {
		return
	}
	for _, ch := range cr {
		switch ch {
		case 'K':
			pf &^= NoCastleWK
		case 'Q':
			pf &^= NoCastleWQ
		case 'k':
			pf &^= NoCastleBK
		case 'q':
			pf &^= NoCastleBQ
		}
	}
	return
}

func parseEnPassant(strpos string) (PositionFlags, error) {
	eppos := SquareFromString(strpos)
	if eppos == SquareInvalid {
		return 0, fmt.Errorf("invalid pos: %s", eppos)
	}
	row, col := eppos.Split()
	if row != 5 && row != 2 {
		return 0, fmt.Errorf("invalid en passant pos: %s", eppos)
	}
	return EnPassant | PositionFlags(col), nil
}
