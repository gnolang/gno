// Package chess implements a simple chess engine, designed to implement all
// of the official FIDE rules with the intention of validating moves for a
// "chess server" realm (gno.land/r/demo/chess).
//
// To implement the rules, the FIDE "Laws of Chess" are used as a reference:
// https://www.fide.com/FIDE/handbook/LawsOfChess.pdf
//
// This package was designed with a focus on clarity and on using this code as
// a didactic tool. Any contributions to the code should respect this.
package chess

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"gno.land/p/morgan/chess/zobrist"
)

// PositionFlags. The lower 4 bits indicate an en passant column; the upper
// 4 indicate castling rights.
type PositionFlags byte

const (
	EnPassant PositionFlags = 1 << (iota + 3)
	NoCastleWQ
	NoCastleWK
	NoCastleBQ
	NoCastleBK

	MaskEnPassant = 7 // low 4 bits
)

// CastlingRights returns FEN castling rights.
// https://www.chessprogramming.org/Forsyth-Edwards_Notation#Castling_ability
func (p PositionFlags) CastlingRights() string {
	s := ""
	if p&NoCastleWK == 0 {
		s += "K"
	}
	if p&NoCastleWQ == 0 {
		s += "Q"
	}
	if p&NoCastleBK == 0 {
		s += "k"
	}
	if p&NoCastleBQ == 0 {
		s += "q"
	}
	if s == "" {
		return "-"
	}
	return s
}

// Position contains the information about a chessboard, and surrounding
// context: the previous moves, the castling rights and "en passant" column.
//
// NOTE: the position of a piece is encoded in a [Square].
type Position struct {
	B     Board
	Moves []Move
	Flags PositionFlags

	// Halfmoves since the last pawn move or capture.
	// https://www.chessprogramming.org/Halfmove_Clock
	HalfMoveClock uint16
	// Used to calculate repeating positions (3- 5-fold repetition).
	// Zobrist hashing: https://www.chessprogramming.org/Zobrist_Hashing
	// Reset together with halfmove clock.
	Hashes []uint64
}

// NewPosition returns a new Position, set up with the initial board position.
func NewPosition() Position {
	return Position{
		B:      NewBoard(),
		Moves:  make([]Move, 0, 80), // typical chess game is ~40 moves, 80 half-moves
		Hashes: []uint64{zobrist.InitialPosition},
	}
}

// Color of the "next" move after p.Moves. (White for even len(p.Moves),
// Black otherwise)
func (p Position) Color() Color { return Color(len(p.Moves)&1 == 1) }

// uintSlice attaches the methods of sort.Interface to []uint64, sorting in increasing order.
type uintSlice []uint64

func (x uintSlice) Len() int           { return len(x) }
func (x uintSlice) Less(i, j int) bool { return x[i] < x[j] }
func (x uintSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// maxHashCount counts the maximum number of repeating hashes in p.Hashes.
func (p Position) maxHashCount() int {
	if len(p.Hashes) == 0 {
		return 0
	}
	sort.Sort(uintSlice(p.Hashes))
	var last uint64
	var cur, max int
	for _, v := range p.Hashes {
		if last != v {
			last = v
			if cur >= max {
				max = cur
			}
			cur = 0
		}
		cur++
	}
	if cur >= max {
		max = cur
	}
	return max
}

func sign(n int8) int8 {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	default:
		return 0
	}
}

func abs(n int8) int8 {
	return n * sign(n)
}

// EncodeFEN encodes p into FEN.
// https://www.chessprogramming.org/Forsyth-Edwards_Notation
func (p Position) EncodeFEN() string {
	var s string
	emptyCount := 0
	// FEN has different ordering from us, as [0] is a black rook while for us
	// is a white rook. So we need to invert the order of rows.
	for i := 56; i >= 0; i++ {
		v := p.B[i]
		if v == PieceEmpty {
			emptyCount++
			if i%8 == 7 {
				s += strconv.Itoa(emptyCount) + "/"
				emptyCount = 0
				i -= 16
			}
			continue
		}
		if emptyCount > 0 {
			s += strconv.Itoa(emptyCount)
			emptyCount = 0
		}
		s += v.String()
		if i%8 == 7 {
			s += "/"
			i -= 16
		}
	}
	// remove trailing slash
	s = s[:len(s)-1]

	strs := []string{
		s,                        // 0: piece placement
		"w",                      // 1: side to move
		p.Flags.CastlingRights(), // 2: castling ability
		"-",                      // 3: e.p. target square

		strconv.Itoa(int(p.HalfMoveClock)), // 4: halfmove clock
		strconv.Itoa(len(p.Moves)/2 + 1),   // 5: fullmove counter
	}

	var epFile byte
	if p.Flags&EnPassant > 0 {
		epFile = 'a' + byte(p.Flags&MaskEnPassant)
	}

	if p.Color() == Black {
		strs[1] = "b"
		if epFile != 0 {
			strs[3] = string(epFile) + "3"
		}
	} else if epFile != 0 {
		strs[3] = string(epFile) + "6"
	}

	return strings.Join(strs, " ")
}

// ValidateMove checks whether the given move is legal in Chess.
//
// Caller must guarantee m.To and m.From to be valid (<64).
func (p Position) ValidateMove(m Move) (newP Position, valid bool) {
	if p.B[m.To].StripColor() == PieceKing {
		return
	}

	return p.validateMove(m)
}

// validateMove allows for m to be a "king-capture" move, which is illegal in
// chess, but it is useful for InCheck.
// This is commonly known in chess programming as a "pseudo-legal" move.
func (oldp Position) validateMove(m Move) (newPos Position, ok bool) {
	p := oldp

	piece := p.B[m.From]

	// piece moved must be of player's color
	color := p.Color()
	if piece == PieceEmpty || piece.Color() != color ||
		// additionally, check piece has actually moved
		m.From == m.To {
		return
	}
	// destination must not be occupied by piece of same color
	if to := p.B[m.To]; to != PieceEmpty && to.Color() == color {
		return
	}

	// one of the two necessarily != 0 (consequence of m.From != m.To).
	delta := m.From.Sub(m.To)
	dr, dc := delta[0], delta[1]

	// Keep old castling rights; remove en passant info.
	newFlags := p.Flags &^ (MaskEnPassant | EnPassant)
	// Marked as true for succesful promotions.
	var promoted bool

	isDiag := func() bool {
		// move diagonally (|dr| == |dc|)
		if abs(dr) != abs(dc) {
			return false
		}
		signr, signc := sign(dr), sign(dc)
		// squares crossed must be empty
		for i := int8(1); i < abs(dr); i++ {
			if p.B[m.From.Move(i*signr, i*signc)] != PieceEmpty {
				return false
			}
		}
		return true
	}
	isHorizVert := func() bool {
		// only one of dr, dc must be 0 (horiz/vert movement)
		if dr != 0 && dc != 0 {
			return false
		}
		// squares crossed must be empty
		for i := int8(1); i < abs(dr); i++ {
			if p.B[m.From.Move(i*sign(dr), 0)] != PieceEmpty {
				return false
			}
		}
		for i := int8(1); i < abs(dc); i++ {
			if p.B[m.From.Move(0, i*sign(dc))] != PieceEmpty {
				return false
			}
		}
		return true
	}

	switch piece.StripColor() {
	case PieceRook:
		if !isHorizVert() {
			return
		}
		// if rook has moved from a starting position, this disables castling
		// on the side of the rook. flag accordingly in the move.
		var fg Square
		if color == Black {
			fg = 7 << 3
		}
		switch m.From {
		case fg: // a-col rook (either side)
			if color == White {
				newFlags |= NoCastleWQ
			} else {
				newFlags |= NoCastleBQ
			}
		case fg | 7: // h-col rook (either side)
			if color == White {
				newFlags |= NoCastleWK
			} else {
				newFlags |= NoCastleBK
			}
		}

	case PieceKnight:
		// move L-shaped
		// rationale: if you only have positive integers, the only way you can
		// obtain x * y == 2 is if x,y are either 1,2 or 2,1.
		if abs(dc*dr) != 2 {
			return
		}

	case PieceBishop:
		if !isDiag() {
			return
		}

	case PieceQueen:
		if !isHorizVert() && !isDiag() {
			return
		}

	case PieceKing:
		// castling
		if abs(dc) == 2 && dr == 0 {
			// determine if castle is a valid form of castling for the given color
			ctype := m.isCastle(color)
			if ctype == 0 {
				return
			}

			if false ||
				// check that there are no previous moves which disable castling
				p.castlingDisabled(color, ctype) ||
				// check that we have the exact board set ups we need
				// + make sure that the original and crossed squares are not in check
				!p.checkCastlingSetup(ctype) {
				return
			}

			// perform rook move here
			p.B = p.B.castleRookMove(color, ctype)
			// add NoCastle flags to prevent any further castling
			if color == White {
				newFlags |= NoCastleWQ | NoCastleWK
			} else {
				newFlags |= NoCastleBQ | NoCastleBK
			}
			break
		}
		// move 1sq in all directions
		if dc < -1 || dc > 1 || dr < -1 || dr > 1 {
			return
		}
		// king has moved: disable castling.
		if color == White {
			newFlags |= NoCastleWQ | NoCastleWK
		} else {
			newFlags |= NoCastleBQ | NoCastleBK
		}

	case PiecePawn:
		// determine direction depending on color
		dir := int8(1)
		if color == Black {
			dir = -1
		}

		switch {
		case dc == 0 && dr == dir: // 1sq up
			// destination must be empty (no captures allowed)
			if p.B[m.To] != PieceEmpty {
				return
			}
		case dc == 0 && dr == dir*2: // 2sq up (only from starting row)
			wantRow := Square(1)
			if color == Black {
				wantRow = 6
			}
			// check starting row, and that two squares are empty
			if (m.From>>3) != wantRow ||
				p.B[m.From.Move(int8(dir), 0)] != PieceEmpty ||
				p.B[m.To] != PieceEmpty {
				return
			}
			_, col := m.To.Split()
			newFlags |= EnPassant | PositionFlags(col)
		case abs(dc) == 1 && dr == dir: // capture on diag
			// must be a capture
			if p.B[m.To] == PieceEmpty {
				if sq := p.checkEnPassant(color, m.To); sq != SquareInvalid {
					// remove other pawn
					p.B[sq] = PieceEmpty
					break
				}
				return
			}
			// p.B[m.To] is necessarily an opponent piece; we check & return
			// p.B[m.To].Color == color at the beginning of the fn.
		default: // not a recognized move
			return
		}

		row := m.To >> 3
		if (color == White && row == 7) ||
			(color == Black && row == 0) {
			switch m.Promotion {
			case 0:
				// m.To is a king? then this is a pseudo-move check.
				// assume queen in that case.
				if p.B[m.To].StripColor() != PieceKing {
					// no promotion given, invalid
					return
				}
				m.Promotion = PieceQueen
			case PieceQueen, PieceBishop, PieceKnight, PieceRook:
			default:
				return
			}
			promoted = true
			p.B[m.From] = m.Promotion | color.Piece()
		}
	}

	// reject moves with promotion if there's nothing to promote
	if m.Promotion != 0 && !promoted {
		return
	}

	if p.B[m.To].StripColor() == PieceKing {
		// King captures don't check for our own king in check;
		// these are only "theoretical" moves.
		return Position{}, true
	}

	// perform board mutation
	capture := p.B[m.To] != PieceEmpty
	p.B[m.From], p.B[m.To] = PieceEmpty, p.B[m.From]
	p.Flags = newFlags
	p.Moves = append([]Move{}, p.Moves...)
	p.Moves = append(p.Moves, m)

	// halfmove clock + hashes logic
	if piece.StripColor() == PiecePawn || capture {
		// reset both
		p.HalfMoveClock = 0
		p.Hashes = nil
	} else {
		p.HalfMoveClock++
	}
	ep := byte(255)
	if p.Flags&EnPassant != 0 {
		ep = byte(p.Flags & MaskEnPassant)
	}
	// Not ideal, but avoids GenMoves potentially polluting "real moves" hashes.
	p.Hashes = append([]uint64{}, p.Hashes...)
	p.Hashes = append(
		p.Hashes,
		// color is inverted, because we consider the present move as already
		// done (hence, it is the other player to move).
		zobrist.Hash(toZobristBoard(p.B), bool(!color), byte(p.Flags)>>4, ep),
	)

	// is our king in check, as a result of the current move?
	if p.B.InCheck(color) {
		return
	}
	return p, true
}

func toZobristBoard(src Board) zobrist.Board {
	var zb zobrist.Board
	for pos, piece := range src {
		zb[pos] = zobrist.Piece(piece)
	}
	return zb
}

// used by InCheck to simulate a move by black player.
var blackPrevMoves = make([]Move, 1)

// InCheck checks whether the king with the given color is in check.
// If such king does not exist on the board, InCheck returns false.
//
// A king is in check if the move from a piece of the other color
// towards the king is valid, ignoring any checks on the other color's king.
//
// NOTE: the last remark is important:
// https://lichess.org/analysis/4k3/8/4b3/8/8/8/K3R3/8_w_-_-_0_1?color=white
// -- this is still a check for white, even if _technically_ black couldn't
// move the bishop (as that would check its own king)
func (b Board) InCheck(color Color) bool {
	pWant := PieceKing | color.Piece()
	kingp := b.findPiece(pWant)
	if kingp == SquareInvalid {
		return false
	}

	pos := Position{B: b}
	if color == White {
		// color == White -> simulate a move by black player -> pos.Moves odd
		pos.Moves = blackPrevMoves
	}

	for sq, piece := range b {
		if piece == PieceEmpty || piece.Color() == color {
			continue
		}
		_, ok := pos.validateMove(Move{
			From: Square(sq),
			To:   kingp,
			// validateMove (unexp) understands that moves to capture a king are
			// pseudo moves, so it doesn't check for checking on its own king,
			// or promotion.
		})
		if ok {
			return true
		}
	}

	return false
}

// Board is a representation of a chess board.
// Details on how to transform a chess algebraic position into an index
// can be found at [Square].
type Board [64]Piece

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

func (b Board) findPiece(pWant Piece) Square {
	for sq, p := range b {
		if p == pWant {
			return Square(sq)
		}
	}
	return SquareInvalid
}

func (p Position) checkCastlingSetup(typ byte) bool {
	// set up correct row and piece flags according to color
	c := p.Color()
	b := p.B
	var fg Square
	var pfg Piece
	if c == Black {
		fg, pfg = 7<<3, PieceBlack
	}

	// _cross are the squares that the king starts from,
	// crosses and "lands". they are recorded as they must all be
	// not in check by any opponent piece.
	var _cross [3]Square

	if typ == 'K' {
		if !(b[fg|4] == pfg|PieceKing &&
			b[fg|5] == PieceEmpty &&
			b[fg|6] == PieceEmpty &&
			b[fg|7] == pfg|PieceRook) {
			return false
		}
		_cross = [3]Square{fg | 4, fg | 5, fg | 6}
	} else {
		if !(b[fg|4] == pfg|PieceKing &&
			b[fg|3] == PieceEmpty &&
			b[fg|2] == PieceEmpty &&
			b[fg|1] == PieceEmpty &&
			b[fg|0] == pfg|PieceRook) {
			return false
		}
		_cross = [3]Square{fg | 4, fg | 3, fg | 2}
	}

	testb := p.B
	for _, sq := range _cross {
		testb[sq] = pfg | PieceKing
		if testb.InCheck(c) {
			return false
		}
		testb[sq] = PieceEmpty
	}

	return true
}

func (b Board) castleRookMove(c Color, typ byte) Board {
	var fg Square
	var pfg Piece
	if c == Black {
		fg, pfg = 7<<3, PieceBlack
	}

	if typ == 'K' {
		b[fg|7], b[fg|5] = PieceEmpty, PieceRook|pfg
	} else {
		b[fg|0], b[fg|3] = PieceEmpty, PieceRook|pfg
	}
	return b
}

func (p Position) castlingDisabled(color Color, kind byte) bool {
	if kind != 'K' && kind != 'Q' {
		return false
	}

	// Determine what flag we're looking for.
	var want PositionFlags
	switch {
	case color == White && kind == 'K':
		want = NoCastleWK
	case color == White && kind == 'Q':
		want = NoCastleWQ
	case color == Black && kind == 'K':
		want = NoCastleBK
	case color == Black && kind == 'Q':
		want = NoCastleBQ
	}

	return p.Flags&want != 0
}

func (p Position) checkEnPassant(c Color, sq Square) Square {
	row, col := sq.Split()
	if p.Flags&EnPassant == 0 ||
		(c == White && row != 5) ||
		(c == Black && row != 2) {
		return SquareInvalid
	}
	wantCol := byte(p.Flags & MaskEnPassant)

	if col != wantCol {
		return SquareInvalid
	}

	if c == White {
		return Square(4<<3 | col)
	}
	return Square(3<<3 | col)
}

// InsufficientMaterial tests for insufficient material on the board, in which
// case it returns true.
//
// See this reference for the rules used:
// https://www.chess.com/article/view/how-chess-games-can-end-8-ways-explained#insufficient-material
//
//   - king vs king
//   - king+N/B vs king
//   - king+NN vs king
//   - king+N/B vs king+N/B
func (bd Board) InsufficientMaterial() bool {
	// strategy:
	// store the pieces which could count for an insufficient material
	// scenario in w, b.
	// if we encounter any pawn, queen, or rook, material is sufficient.
	// if we encounter 2 bishops, material is sufficient.
	// afterwards, we verify that w and b are one of the possible insuf material
	// scenarios.

	const (
		imN  byte = 1 << iota // knight
		imN2                  // second knight
		imB                   // bishop
	)
	var w, b byte
	for _, p := range bd {
		strip := p.StripColor()
		switch strip {
		case PieceQueen, PiecePawn, PieceRook:
			return false
		case PieceKing, PieceEmpty:
			continue
		}
		// strip is one of PieceBishop PieceKnight
		t := &w
		if p.Color() == Black {
			t = &b
		}
		if strip == PieceBishop {
			if *t&imB != 0 {
				return false
			}
			*t |= imB
		} else {
			if *t&imN2 != 0 {
				// third knight (ie. from pawn promotion)
				return false
			}
			if *t&imN != 0 {
				*t |= imN2
			} else {
				*t |= imN
			}
		}
	}
	// make it so that w is bigger than b
	if b > w {
		w, b = b, w
	}
	switch [2]byte{w, b} {
	case [2]byte{0, 0}, // king vs king
		[2]byte{imN, 0}, // king+N/B vs king
		[2]byte{imB, 0},
		[2]byte{imN | imN2, 0}, // king+NN vs king
		[2]byte{imN, imN},      // king+N/B vs king+N/B
		[2]byte{imB, imB},
		[2]byte{imB, imN}: // imB > imN, so [2]byte{imN, imB} cannot happen
		return true
	}
	return false
}

// GenMoves implements a rudimentary move generator.
// This is not used beyond aiding in determing stalemate and doing perft tests.
// Each generated move is passed to cb.
// If cb returns an error, it is returned without processing further moves.
func (p Position) GenMoves(cb func(Position, Move) error) error {
	color := p.Color()
	for sq, piece := range p.B {
		if piece == PieceEmpty || piece.Color() != color {
			continue
		}

		from := Square(sq)

		pstrip := piece.StripColor()
		// If the piece is a pawn, and they are on the second last row, we know
		// that whatever move they do (advance, or take diagonally) they're going
		// to promote.
		prom := pstrip == PiecePawn &&
			((color == White && from>>3 == 6) ||
				(color == Black && from>>3 == 1))

		// delta generator needs to know if p is black
		if pstrip == PiecePawn && color == Black {
			pstrip |= Black.Piece()
		}

		var err error
		deltaGenerator(pstrip, func(delta Delta) byte {
			// create move; if the resulting square is oob, continue
			m := Move{
				From: from,
				To:   from.Apply(delta),
			}
			if m.To == SquareInvalid ||
				(p.B[m.To] != PieceEmpty && p.B[m.To].Color() == color) {
				return deltaGenStopLinear
			}

			// handle promotion case
			if prom {
				m.Promotion = PieceQueen
			}

			// if it's a valid move, call cb on it
			newp, ok := p.ValidateMove(m)
			if !ok {
				return deltaGenOK
			}
			if err = cb(newp, m); err != nil {
				return deltaGenStop
			}

			// if we've promoted, handle the cases where we've promoted to a non-queen.
			if !prom {
				return deltaGenOK
			}

			for _, promPiece := range [...]Piece{PieceRook, PieceKnight, PieceBishop} {
				newp.B[m.To] = promPiece | color.Piece()
				m.Promotion = promPiece
				if err = cb(newp, m); err != nil {
					return deltaGenStop
				}
			}
			return deltaGenOK
		})
		if err != nil {
			return err
		}
	}
	return nil
}

const (
	// carry on normally
	deltaGenOK = iota
	// if the generator is doing a linear attack (ie. rook, bishop, queen),
	// then stop that (there is a piece of same colour in the way.)
	deltaGenStopLinear
	// abort generation asap.
	deltaGenStop
)

/*func init() {
	for i := PiecePawn; i <= PieceKing; i++ {
		println("generator ", i.String())
		deltaGenerator(i, func(d Delta) byte {
			println("  ", d[0], d[1])
			return deltaGenOK
		})
	}
}*/

// deltaGenerator generates the possible ways in which p can move.
// the callback may return one of the three deltaGen* values.
func deltaGenerator(p Piece, cb func(d Delta) byte) {
	doLinear := func(d Delta) bool {
		for i := int8(1); i <= 7; i++ {
			switch cb(d.Mul(i)) {
			case deltaGenStopLinear:
				return false
			case deltaGenStop:
				return true
			}
		}
		return false
	}
	rotate := func(d Delta, lin bool) bool {
		for i := 0; i < 4; i++ {
			if lin {
				if doLinear(d) {
					return true
				}
			} else {
				if cb(d) == deltaGenStop {
					return true
				}
			}

			d = d.Rot()
		}
		return false
	}

	// In the following, we use logical OR's to do conditional evaluation
	// (if the first item returns true, the second won't be evaluated)
	switch p {
	case PiecePawn, PiecePawn | PieceBlack:
		dir := int8(1)
		if p.Color() == Black {
			dir = -1
		}
		// try moving 1sq forward; if we get StopLinear, don't try to do 2sq.
		fw := cb(Delta{dir, 0})
		if fw == deltaGenStop {
			return
		}
		if fw != deltaGenStopLinear {
			if cb(Delta{dir * 2, 0}) == deltaGenStop {
				return
			}
		}

		_ = cb(Delta{dir, 1}) == deltaGenStop ||
			cb(Delta{dir, -1}) == deltaGenStop

	case PieceRook:
		rotate(Delta{0, 1}, true)
	case PieceBishop:
		rotate(Delta{1, 1}, true)
	case PieceKnight:
		_ = rotate(Delta{1, 2}, false) ||
			rotate(Delta{2, 1}, false)
	case PieceQueen:
		_ = rotate(Delta{0, 1}, true) ||
			rotate(Delta{1, 1}, true)
	case PieceKing:
		_ = rotate(Delta{0, 1}, false) ||
			rotate(Delta{1, 1}, false) ||
			cb(Delta{0, 2}) == deltaGenStop ||
			cb(Delta{0, -2}) == deltaGenStop
	}
}

// Outcome is returned by IsFinished, and indicates if the game has finished,
// and additionally if either of the player can claim draw by the 50-move rule
// or by 3-fold repetition.
//
// Either one of the sequential outcomes will be set (values 1-5), indicating
// that the game is terminated, or the low 3 bits will be 0, and
// the Can* flags may be set.
type Outcome byte

const (
	NotFinished Outcome = iota
	Checkmate
	Stalemate
	Drawn75Move
	Drawn5Fold

	Can50Move       Outcome = 8
	Can3Fold        Outcome = 16
	CanInsufficient Outcome = 32
)

var errFound = errors.New("found")

// IsFinished determines if the game at the given position can be considered
// "finished".
// See [Outcome] for the possible results.
func (p Position) IsFinished() Outcome {
	err := p.GenMoves(func(Position, Move) error {
		return errFound
	})
	// If there is any legal move, this is not any kind of mate.
	if err != nil {
		mhc := p.maxHashCount()
		switch {
		case mhc >= 5:
			return Drawn5Fold
		case p.HalfMoveClock >= 150:
			return Drawn75Move
		}
		var o Outcome
		if mhc >= 3 {
			o |= Can3Fold
		}
		if p.HalfMoveClock >= 100 {
			o |= Can50Move
		}
		if p.B.InsufficientMaterial() {
			o |= CanInsufficient
		}
		return o
	}

	// No legal moves. Is the king in check?
	if p.B.InCheck(p.Color()) {
		return Checkmate
	}
	return Stalemate
}

// Color determines a player's color -- either white or black.
type Color bool

const (
	White Color = false
	Black Color = true
)

// Piece returns the color as a piece to be OR'd into a Piece;
// ie. 0 on White, and [PieceBlack] on black.
func (c Color) Piece() Piece {
	if c == White {
		return 0
	}
	return PieceBlack
}

// Piece represents a piece on the board.
type Piece byte

// PieceFromChar returns the piece corresponding to the given character.
// White pieces are uppercase, and black pieces are lowercase.
// If a piece is invalid, PieceInvalid is returned.
func PieceFromChar(b byte) Piece {
	return p[b]
}

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

var pstring = [PieceBlack | PieceKing + 1]byte{
	PiecePawn:                'P',
	PieceRook:                'R',
	PieceKnight:              'N',
	PieceBishop:              'B',
	PieceQueen:               'Q',
	PieceKing:                'K',
	PieceBlack | PiecePawn:   'p',
	PieceBlack | PieceRook:   'r',
	PieceBlack | PieceKnight: 'n',
	PieceBlack | PieceBishop: 'b',
	PieceBlack | PieceQueen:  'q',
	PieceBlack | PieceKing:   'k',
}

func (p Piece) String() string {
	if int(p) >= len(pstring) {
		return ""
	}
	v := pstring[p]
	if v == 0 {
		return ""
	}
	return string(v)
}

// Possible values of Piece. Within the context of Board, Piece is assumed to
// be white, unless p&PieceBlack != 0. Note PieceBlack is not a valid piece; it
// must be bitwise OR'd to a non-empty piece.
const (
	PieceEmpty Piece = iota

	PiecePawn
	PieceRook
	PieceKnight
	PieceBishop
	PieceQueen
	PieceKing

	PieceBlack Piece = 8 // bit-flag
)

// Color returns the color of the piece.
func (p Piece) Color() Color { return Color(p&PieceBlack != 0) }

// Piece returns the given Piece without color information.
func (p Piece) StripColor() Piece { return p &^ PieceBlack }

// Switch switches the color of the given piece.
func (p Piece) Switch() Piece {
	if p.Color() == Black {
		return p &^ PieceBlack
	}
	return p | PieceBlack
}

// Delta represents a 2d vector for indicating a movement from one square
// to another. The first value indicates the change in column, the second the
// change in rows.
type Delta [2]int8

// Valid ensures the two values of delta are valid.
func (d Delta) Valid() bool {
	return d[0] >= -7 && d[0] <= 7 &&
		d[1] >= -7 && d[1] <= 7 &&
		!(d[0] == 0 && d[1] == 0)
}

// Rot applies a 90 degree anti-clockwise rotation to d.
func (d Delta) Rot() Delta {
	// Rationale: this is just matrix-vector multiplication.
	// 90 deg rotation is just the matrix {0, -1; 1, 0}.
	return Delta{d[1], -d[0]}
}

// Mul multiplies both values by n, otherwise known as scalar product.
func (d Delta) Mul(n int8) Delta {
	return Delta{d[0] * n, d[1] * n}
}

// Square encodes piece position information, in chess the "square" the piece is on.
// Indexing 0 as the LSB, bits 0-3 indicate the column and bits 4-6 indicate
// the row. For instance, square 44 (decimal) is:
//
//	44 = 0b00 101  100  = d5
//	          ^row ^col
//
// (note: in algebraic notation, this is swapped: the letter represents the
// column, and the number represents the row).
type Square byte

// SquareInvalid is returned by some Square-related methods to indicate
// invalid parameters.
const SquareInvalid Square = 255

// String returns p in algebraic notation.
func (q Square) String() string {
	if q >= 64 {
		return "<invalid>"
	}
	return string(q&7+'a') + string(q>>3+'1')
}

// SquareFromString returns Square, reading the human-readable algebraic
// notation in s. s must be 2 bytes long, with the first byte a letter included
// between ['a'; 'h'], and the second a number included between ['1';'8'].
// If s is invalid, SquareInvalid is returned.
func SquareFromString(s string) Square {
	if len(s) != 2 {
		return SquareInvalid
	}
	col, row := s[0]-'a', s[1]-'1'
	// because s[0] is a byte, if s[0] < 'a' then the above will underflow and
	// row will be >= 8 (same for col).
	if row >= 8 || col >= 8 {
		return SquareInvalid
	}
	return Square(row<<3 | col)
}

// Move changes the square of q, moving it vertically according to dr
// (delta row) and horizontally according to dc (delta column).
// If the resulting square is not on the board, then SquareInvalid is returned.
func (q Square) Move(dr, dc int8) Square {
	if q == SquareInvalid || !(Delta{dr, dc}).Valid() {
		return SquareInvalid
	}

	row, col := int8(q>>3), int8(q&7)
	row += dr
	col += dc

	nr, nc := Square(row), Square(col)
	if nr >= 8 || nc >= 8 {
		return SquareInvalid
	}
	return nr<<3 | nc
}

// Apply applies the given delta to the square.
// It is shorthand for q.Move(d[0], d[1]).
func (q Square) Apply(d Delta) Square { return q.Move(d[0], d[1]) }

// Split splits Square into its components.
// This function does not check if p is invalid.
func (q Square) Split() (row, col byte) {
	return byte(q >> 3), byte(q & 7)
}

// SplitI works like [Square.Split], but returns int8's instead
// of bytes.
func (q Square) SplitI() (row, col int8) {
	return int8(q >> 3), int8(q & 7)
}

// Sub calculates the difference between the two squares.
// q is the originating square, s is the ending square. The difference in
// rows and columns from q to s is returned; for instance, d1.Sub(a4) yields
// Delta{3, -3}.
func (q Square) Sub(s Square) Delta {
	fr, fc := q.SplitI()
	tr, tc := s.SplitI()
	return Delta{tr - fr, tc - fc}
}

// Move represents a chess game move.
type Move struct {
	From, To  Square
	Promotion Piece
}

// String returns a string representation of Move.
// It is in the form of "Long Algebraic Notation".
// https://backscattering.de/chess/uci/#move-lan
func (m Move) String() string {
	p := ""
	if m.Promotion != 0 {
		p = string(m.Promotion.String()[0] + ('a' - 'A'))
	}
	return m.From.String() + m.To.String() + p
}

var (
	castleWhiteQ = Move{From: SquareFromString("e1"), To: SquareFromString("c1")}
	castleWhiteK = Move{From: SquareFromString("e1"), To: SquareFromString("g1")}
	castleBlackQ = Move{From: SquareFromString("e8"), To: SquareFromString("c8")}
	castleBlackK = Move{From: SquareFromString("e8"), To: SquareFromString("g8")}
)

// returns 0, 'K' or 'Q'.
func (m Move) isCastle(c Color) (kind byte) {
	if c == White {
		switch m {
		case castleWhiteQ:
			return 'Q'
		case castleWhiteK:
			return 'K'
		}
	} else {
		switch m {
		case castleBlackQ:
			return 'Q'
		case castleBlackK:
			return 'K'
		}
	}
	return 0
}
