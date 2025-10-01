// Realm chess implements a Gno chess server.
package chess

import (
	"chain/runtime"
	"errors"
	"time"

	"gno.land/p/morgan/chess"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
)

// realm state
var (
	// (not "games" because that's too useful a variable name)
	gameStore     avl.Tree // string (game ID) -> *Game
	gameIDCounter seqid.ID

	// Value must be sorted by game ID, descending
	user2Games avl.Tree // address -> []*Game
)

// Game represents a chess game.
type Game struct {
	ID string `json:"id"`

	White    address        `json:"white"`
	Black    address        `json:"black"`
	Position chess.Position `json:"position"`
	State    GameState      `json:"state"`
	Winner   Winner         `json:"winner"`

	Creator     address   `json:"creator"`
	CreatedAt   time.Time `json:"created_at"`
	DrawOfferer *address  `json:"draw_offerer"` // set on draw offers
	Concluder   *address  `json:"concluder"`    // set on non-auto draws, and aborts

	Time *TimeControl `json:"time"`
}

func (g Game) json() string {
	s, err := g.MarshalJSON()
	checkErr(err)
	return string(s)
}

// Winner represents the "direct" outcome of a game
// (white, black or draw?)
type Winner byte

const (
	WinnerNone Winner = iota
	WinnerWhite
	WinnerBlack
	WinnerDraw
)

var winnerString = [...]string{
	WinnerNone:  "none",
	WinnerWhite: "white",
	WinnerBlack: "black",
	WinnerDraw:  "draw",
}

// GameState represents the current game state.
type GameState byte

const (
	GameStateInvalid = iota

	GameStateOpen

	// "automatic" endgames following moves
	GameStateCheckmated
	GameStateStalemate
	GameStateDrawn75Move
	GameStateDrawn5Fold

	// single-party draws
	GameStateDrawn50Move
	GameStateDrawn3Fold
	GameStateDrawnInsufficient

	// timeout by either player
	GameStateTimeout
	// aborted within first two moves
	GameStateAborted
	// resignation by either player
	GameStateResigned
	// draw by agreement
	GameStateDrawnByAgreement
)

var gameStatesSnake = [...]string{
	GameStateInvalid:           "invalid",
	GameStateOpen:              "open",
	GameStateCheckmated:        "checkmated",
	GameStateStalemate:         "stalemate",
	GameStateDrawn75Move:       "drawn_75_move",
	GameStateDrawn5Fold:        "drawn_5_fold",
	GameStateDrawn50Move:       "drawn_50_move",
	GameStateDrawn3Fold:        "drawn_3_fold",
	GameStateDrawnInsufficient: "drawn_insufficient",
	GameStateTimeout:           "timeout",
	GameStateAborted:           "aborted",
	GameStateResigned:          "resigned",
	GameStateDrawnByAgreement:  "drawn_by_agreement",
}

// IsFinished returns whether the game is in a finished state.
func (g GameState) IsFinished() bool {
	return g != GameStateOpen
}

// NewGame initialized a new game with the given opponent.
// opponent may be a bech32 address or "@user" (r/demo/users).
//
// seconds and increment specifies the time control for the given game.
// seconds is the amount of time given to play to each player; increment
// is by how many seconds the player's time should be increased when they make a move.
// seconds <= 0 means no time control (correspondence).
//
// XXX: Disabled for GnoChess production temporarily. (prefixed with x for unexported)
// Ideally, we'd need this to work either by not forcing users not to have
// parallel games OR by introducing a "request" system, so that a game is not
// immediately considered "open" when calling NewGame.
func xNewGame(cur realm, opponentRaw string, seconds, increment int) string {
	assertOriginCall()
	if seconds >= 0 && increment < 0 {
		panic("negative increment invalid")
	}

	opponent := parsePlayer(opponentRaw)
	caller := runtime.PreviousRealm().Address()
	assertUserNotInLobby(caller)

	return newGame(caller, opponent, seconds, increment).json()
}

func getUserGames(user address) []*Game {
	val, exist := user2Games.Get(user.String())
	if !exist {
		return nil
	}
	return val.([]*Game)
}

func assertGamesFinished(games []*Game) {
	for _, g := range games {
		if g.State.IsFinished() {
			continue
		}
		err := g.claimTimeout()
		if err != nil {
			panic("can't start new game: game " + g.ID + " is not yet finished")
		}
	}
}

func newGame(caller, opponent address, seconds, increment int) *Game {
	games := getUserGames(caller)
	// Ensure player has no ongoing games.
	assertGamesFinished(games)
	assertGamesFinished(getUserGames(opponent))

	if caller == opponent {
		panic("can't create a game with yourself")
	}

	isBlack := determineColor(games, caller, opponent)

	// id is zero-padded to work well with avl's alphabetic order.
	id := gameIDCounter.Next().String()
	g := &Game{
		ID:        id,
		White:     caller,
		Black:     opponent,
		Position:  chess.NewPosition(),
		State:     GameStateOpen,
		Creator:   caller,
		CreatedAt: time.Now(),
		Time:      NewTimeControl(seconds, increment),
	}
	if isBlack {
		g.White, g.Black = g.Black, g.White
	}

	gameStore.Set(g.ID, g)
	addToUser2Games(caller, g)
	addToUser2Games(opponent, g)

	return g
}

func addToUser2Games(addr address, game *Game) {
	var games []*Game
	v, ok := user2Games.Get(string(addr))
	if ok {
		games = v.([]*Game)
	}
	// game must be at top, because it is the latest ID
	games = append([]*Game{game}, games...)
	user2Games.Set(string(addr), games)
}

func determineColor(games []*Game, caller, opponent address) (isBlack bool) {
	// fast path for no games
	if len(games) == 0 {
		return false
	}

	// Determine color of player. If the player has already played with
	// opponent, invert from last game played among them.
	// Otherwise invert from last game played by the player.
	isBlack = games[0].White == caller

	// "try" to save gas if the user has really a lot of past games
	if len(games) > 256 {
		games = games[:256]
	}
	for _, game := range games {
		if game.White == opponent || game.Black == opponent {
			return game.White == caller
		}
	}
	return
}

// GetGame returns a game, knowing its ID.
func GetGame(id string) string {
	return getGame(id, false).json()
}

func getGame(id string, wantOpen bool) *Game {
	graw, ok := gameStore.Get(id)
	if !ok {
		panic("game not found")
	}
	g := graw.(*Game)
	if wantOpen && g.State.IsFinished() {
		panic("game is already finished")
	}
	return g
}

// MakeMove specifies a move to be done on the given game, specifying in
// algebraic notation the square where to move the piece.
// If the piece is a pawn which is moving to the last row, a promotion piece
// must be specified.
// Castling is specified by indicating the king's movement.
func MakeMove(cur realm, gameID, from, to string, promote chess.Piece) string {
	assertOriginCall()
	g := getGame(gameID, true)

	// determine if this is a black move
	isBlack := len(g.Position.Moves)%2 == 1

	caller := runtime.PreviousRealm().Address()
	if (isBlack && g.Black != caller) ||
		(!isBlack && g.White != caller) {
		// either not a player involved; or not the caller's turn.
		panic("you are not allowed to make a move at this time")
	}

	// game is time controlled? add move to time control
	if g.Time != nil {
		valid := g.Time.AddMove()
		if !valid && len(g.Position.Moves) < 2 {
			g.State = GameStateAborted
			g.Concluder = &caller
			g.Winner = WinnerNone
			return g.json()
		}
		if !valid {
			g.State = GameStateTimeout
			if caller == g.White {
				g.Winner = WinnerBlack
			} else {
				g.Winner = WinnerWhite
			}
			g.saveResult()
			return g.json()
		}
	}

	// validate move
	m := chess.Move{
		From: chess.SquareFromString(from),
		To:   chess.SquareFromString(to),
	}
	if m.From == chess.SquareInvalid || m.To == chess.SquareInvalid {
		panic("invalid from/to square")
	}
	if promote > 0 && promote <= chess.PieceKing {
		m.Promotion = promote
	}
	newp, ok := g.Position.ValidateMove(m)
	if !ok {
		panic("illegal move")
	}

	// add move and record new board
	g.Position = newp

	o := newp.IsFinished()
	if o == chess.NotFinished {
		// opponent of draw offerer has made a move. take as implicit rejection of draw.
		if g.DrawOfferer != nil && *g.DrawOfferer != caller {
			g.DrawOfferer = nil
		}

		return g.json()
	}

	switch {
	case o == chess.Checkmate && isBlack:
		g.State = GameStateCheckmated
		g.Winner = WinnerBlack
	case o == chess.Checkmate && !isBlack:
		g.State = GameStateCheckmated
		g.Winner = WinnerWhite
	case o == chess.Stalemate:
		g.State = GameStateStalemate
		g.Winner = WinnerDraw

	case o == chess.Drawn75Move:
		g.State = GameStateDrawn75Move
		g.Winner = WinnerDraw
	case o == chess.Drawn5Fold:
		g.State = GameStateDrawn5Fold
		g.Winner = WinnerDraw
	}
	g.DrawOfferer = nil
	g.saveResult()

	return g.json()
}

func (g *Game) claimTimeout() error {
	// no assert origin call or caller check: anyone can claim a game to have
	// finished in timeout.

	if g.Time == nil {
		return errors.New("game is not time controlled")
	}

	// game is time controlled? add move to time control
	to := g.Time.TimedOut()
	if !to {
		return errors.New("game is not timed out")
	}

	if nmov := len(g.Position.Moves); nmov < 2 {
		g.State = GameStateAborted
		if nmov == 1 {
			g.Concluder = &g.Black
		} else {
			g.Concluder = &g.White
		}
		g.Winner = WinnerNone
		return nil
	}

	g.State = GameStateTimeout
	if len(g.Position.Moves)&1 == 0 {
		g.Winner = WinnerBlack
	} else {
		g.Winner = WinnerWhite
	}
	g.DrawOfferer = nil
	g.saveResult()

	return nil
}

// ClaimTimeout should be called when the caller believes the game has resulted
// in a timeout.
func ClaimTimeout(cur realm, gameID string) string {
	g := getGame(gameID, true)
	err := g.claimTimeout()
	checkErr(err)

	return g.json()
}

func Abort(cur realm, gameID string) string {
	assertOriginCall()
	g := getGame(gameID, true)
	err := abort(g)
	if err != nil {
		panic(err.Error())
	}
	return g.json()
}

func abort(g *Game) error {
	if len(g.Position.Moves) >= 2 {
		return errors.New("game can no longer be aborted; if you wish to quit, resign")
	}

	caller := runtime.PreviousRealm().Address()
	if caller != g.White && caller != g.Black {
		return errors.New("you are not involved in this game")
	}
	g.State = GameStateAborted
	g.Concluder = &caller
	g.DrawOfferer = nil
	g.Winner = WinnerNone

	return nil
}

func Resign(cur realm, gameID string) string {
	assertOriginCall()
	g := getGame(gameID, true)
	err := resign(g)
	if err != nil {
		panic(err.Error())
	}
	return g.json()
}

func resign(g *Game) error {
	if len(g.Position.Moves) < 2 {
		return abort(g)
	}
	caller := runtime.PreviousRealm().Address()
	switch caller {
	case g.Black:
		g.State = GameStateResigned
		g.Winner = WinnerWhite
	case g.White:
		g.State = GameStateResigned
		g.Winner = WinnerBlack
	default:
		return errors.New("you are not involved in this game")
	}
	g.DrawOfferer = nil
	g.saveResult()

	return nil
}

// DrawOffer creates a draw offer in the current game, if one doesn't already
// exist.
func DrawOffer(cur realm, gameID string) string {
	assertOriginCall()
	g := getGame(gameID, true)
	caller := runtime.PreviousRealm().Address()

	switch {
	case caller != g.Black && caller != g.White:
		panic("you are not involved in this game")
	case g.DrawOfferer != nil:
		panic("a draw offer in this game already exists")
	}

	g.DrawOfferer = &caller
	return g.json()
}

// DrawRefuse refuse a draw offer in the given game.
func DrawRefuse(cur realm, gameID string) string {
	assertOriginCall()
	g := getGame(gameID, true)
	caller := runtime.PreviousRealm().Address()

	switch {
	case caller != g.Black && caller != g.White:
		panic("you are not involved in this game")
	case g.DrawOfferer == nil:
		panic("no draw offer present")
	case *g.DrawOfferer == caller:
		panic("can't refuse an offer you sent yourself")
	}

	g.DrawOfferer = nil
	return g.json()
}

// Draw implements draw by agreement, as well as "single-party" draw:
// - Threefold repetition (§9.2)
// - Fifty-move rule (§9.3)
// - Insufficient material (§9.4)
// Note: stalemate happens as a consequence of a Move, and thus is handled in that function.
func Draw(cur realm, gameID string) string {
	assertOriginCall()
	g := getGame(gameID, true)

	caller := runtime.PreviousRealm().Address()
	if caller != g.Black && caller != g.White {
		panic("you are not involved in this game")
	}

	// accepted draw offer (do early to avoid gas for g.Position.IsFinished())
	if g.DrawOfferer != nil && *g.DrawOfferer != caller {
		g.State = GameStateDrawnByAgreement
		g.Winner = WinnerDraw
		g.Concluder = &caller

		g.saveResult()

		return g.json()
	}

	o := g.Position.IsFinished()
	switch {
	case o&chess.Can50Move != 0:
		g.State = GameStateDrawn50Move
	case o&chess.Can3Fold != 0:
		g.State = GameStateDrawn3Fold
	case o&chess.CanInsufficient != 0:
		g.State = GameStateDrawnInsufficient
	default:
		panic("this game can't be automatically drawn")
	}
	g.Concluder = &caller
	g.Winner = WinnerDraw
	g.DrawOfferer = nil

	g.saveResult()

	return g.json()
}

func checkErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// Replacement for OriginCall using std.PreviousRealm().
func assertOriginCall() {
	if !runtime.PreviousRealm().IsUser() {
		panic("invalid non-origin call")
	}
}
