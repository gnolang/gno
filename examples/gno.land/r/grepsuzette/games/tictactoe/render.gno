package tictactoe

// Stateless human VS CPU Tic-tac-toe
// Markdown + HTML1.0 + gnolang
// no javascript, no wallet needed.

import (
	"math/rand"
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/grepsuzette/games/tictactoe/tictactoe1p"
	"gno.land/p/grepsuzette/querystring"
	"gno.land/p/grepsuzette/ternary"
)

const (
	cpuAddress   = std.Address("gCPU")
	humanAddress = std.Address("gHUMAN")
	urlParrot    = "https://raw.githubusercontent.com/grepsuzette/gfx/master/parrot.png"
	statusWon    = "Looks like you've won!"
	statusLost   = "Sorry mate, you lost!"
	statusDraw   = "It's a draw..."
)

var prng *rand.Rand

func Intn(n int) int {
	if prng == nil {
		// Note: our PRNG is not stateful as calling Render is not going
		// to modify this stateless realm. We initialize it here when still nil
		// this creates the randomness we need (seeded from blockchain's height)
		prng = rand.New(rand.NewPCG(uint64(std.GetHeight()), uint64(9)))
	}
	return prng.IntN(n)
}

func Render(path string) string {
	game, x, y, debug := parsePath(path)
	if x != -1 && y != -1 {
		game.Play(humanAddress, x, y)
	}
	cpuX, cpuY, e := game.PlayCPU()
	output := parrotTalk(*game)
	output += render(*game, cpuX, cpuY)
	if debug {
		output += ufmt.Sprintf(
			"--- played x=%d y=%d cpuX=%d cpuY=%d height=%d path=%s turn:%d error:%s",
			x, y, cpuX, cpuY, std.GetHeight(), path, game.Game.TurnNumber(), e,
		)
	}
	return output
}

// Lower-case render is simply called by Render.
func render(game tictactoe1p.Game, cpuX, cpuY int) string {
	status, class := statusAndClass(game)
	repr := game.ToRepr()
	output := css()
	output += `<form id="board" method="GET" class="` + class + `">` + "\n"
	output += "\t" + `<input type="hidden" name="state" value="` + repr + `">` + "\n"
	output += `<div id="left"><img src="` + urlParrot + `" id="parrot" width="120" height="120" align="left" /></div>` + "\n"
	output += `<div id="right">`

	for y := 2; y >= 0; y-- {
		output += "\t"
		for x := 0; x <= 2; x++ {
			runeAtXY := game.At(x, y)
			occupied := runeAtXY != rune(0)
			if occupied {
				highlighted := x == cpuX && y == cpuY
				output += button(x, y, ternary.String(runeAtXY == rune(1), "O", "X"), true, highlighted, repr)
			} else if game.IsOver() {
				output += button(x, y, "&nbsp;", true, false, repr)
			} else {
				// tile is free, but :hover must show
				cpu1st := game.Game.PlayerByIndex(0) == cpuAddress
				output += button(x, y, ternary.String(cpu1st, "X", "O"), false, false, repr)
			}
		}
		output += "<br />\n"
	}
	output += "</div>\n"
	output += "</form>\n"
	if game.IsOver() {
		output += ufmt.Sprintf(
			"\n%s [ %s | %s ]\n",
			status,
			"[New game](/r/grepsuzette/games/tictactoe)",
			"[Back to grepsuzette/games](/r/grepsuzette/games)",
		)
	} else {
		output += "\n[Okay country roads, take me home](/r/grepsuzette/games)"
	}
	return output
}

func button(x, y int, char string, occupied, highlighted bool, repr string) string {
	return ufmt.Sprintf(
		`<button name="move" value="%c%d" %s class="button tile %s">%s</button>`,
		rune('a'+x), y+1,
		ternary.String(occupied, "disabled", ""),
		ternary.String(highlighted, "highlighted", ""),
		char,
	)
}

// return status and class(es).
// it's empty when the game is not over.
func statusAndClass(game tictactoe1p.Game) (status string, classes string) {
	if game.IsOver() {
		switch {
		case game.Winner() == humanAddress:
			classes = "over won"
			status = statusWon
		case game.Winner() == cpuAddress:
			classes = "over lost"
			status = statusLost
		default:
			classes = "over draw"
			status = statusDraw
		}
	}
	return
}

// Decompose path, into valid game and played coordinates (or -1,-1)
// The path is like "move=c3?state=O--X-----"
//
//   - state is empty or 9 characters,
//     imagine the following but without the \n: "-OX\nO--\n---".
//     O always plays first.
//
//   - move coordinate ranges from "a1" to "c3" but can be empty.
//
//   - debug &debug enables printing of a debug line.
func parsePath(path string) (game *tictactoe1p.Game, x, y int, debug bool) {
	h, e := querystring.ToMap(path)
	if e != nil {
		panic(e)
	}
	// nil: default AI
	game, e = tictactoe1p.GameFromRepr(h["state"], cpuAddress, humanAddress, Intn, nil)
	if e != nil {
		panic(e)
	}
	_, debug = h["debug"]
	x, y = -1, -1
	for i, xx := range h["move"] {
		switch i {
		case 0:
			x = int(xx) - int('a')
		case 1:
			y = int(xx) - int('1')
		default:
			panic("invalid move: " + h["move"])
		}
	}
	return game, x, y, debug
}

func parrotTalk(game tictactoe1p.Game) string {
	var a []string
	switch {
	case !game.IsOver():
		a = []string{
			"Let's play Tic-tac-toe!",
			"Cap'n Cluck, am here and ready to rule the high seas of tic-tac-toe!",
			"Aarrr, we'll see who'll be top parrot on this jolly board.",
			"The stakes are higher than the mast of a sunken galleon!",
		}
	case game.IsDraw():
		a = []string{
			"Three cheers for Yers Truly, Cap'n! It's a draw.",
			"It's a draw, sailor...",
			"This match ends in neither victory nor defeat, but a draw!",
		}
	case game.Winner().String() == cpuAddress.String():
		a = []string{
			"Alrighty, mateys! Who's the bravest birdy of the seven seas!",
			"Avast, ye scallywags, for I won this game",
			"Remember, me hearties, even in victory, Cap'n Cluck remains a humble scallywag.",
			"Aarrr, parrot power prevails again!",
			"Cap'n Cluck claims the prize!",
			"The high seas crown me king o' the game!",
		}
	case game.Winner().String() != cpuAddress.String():
		a = []string{
			"Ye bested a bold parrot on this day!",
			"Cap'n Cluck, beaten by a buccaneer? Nay, 'tis unbirdable!",
			"Aarrr, this be the day I, Cap'n Cluck, ate me parroted pirate's words! But fear not, for I shall be back!",
			"Ye bested this parrot, but amongst feathery fiends, revenge is ripe!",
			"Despite this defeat, remember, I, Cap'n Cluck, am not a chicken when it comes to tic-tac-toe!",
		}
	}
	s := a[Intn(len(a))]
	return ufmt.Sprintf("<div><b>%s</b><br />"+strings.Repeat("&nbsp;", 17)+"/</div>", s)
}

func css() string {
	return `
<style type="text/css">
	/* responsive stuffs */
	#board { 
		display: flex; 
		flex-direction: row;
	}
	#board > #left {
		flex-shrink: 1;
	}
	#board > #right {
		flex-shrink: 0;
		flex-basis: auto;
		width: 9.5em;
	}
	/* buttons, and parrot */
	#board a.tile.button { 
	  appearance: button;
	  box-sizing: border-box;
	  margin: 0;
	  font-weight: bold;
	  display: inline-block;
	  background-color: #eee;
	  border-color: rgb(227, 227, 227);
	  border-style: outset;
	  border-width: 1px;
	  border-collapse: separate;
	  text-decoration: none;
	  text-align: center;
	  line-height: 3em;
	}
    #board button, #board a.tile.button { 
        width: 3em; height: 3em; 
        margin-right: 2px; margin-bottom: 2px; 
        cursor: pointer; 
        color: initial;
		border-radius: 5px;
    }
	#board.won button, #board.won a.tile.button  { box-shadow: aquamarine 1px 1px 12px 6px; }
	#board.lost button, #board.lost a.tile.button  { box-shadow: rgb(255, 200, 190) 0px 1px 34px 2px; }
	#board.draw button , #board.draw a.tile.button { box-shadow: rgb(200, 200, 200) 0px 1px 30px 8px; }
	#board button, #board a.tile.button { color: #888; }
	#board button:not([disabled]):hover, #board a.tile.button:not(.disabled):hover { 
        border-style: ridge; 
        box-shadow: inset -10px -10px 15px rgba(255, 255, 255, 0.5), 
                    inset 10px 10px 15px rgba(70, 70, 70, 0.12);
		border-radius: 5px;
    }
    #board button[disabled], #board a.tile.button.disabled { cursor: default; }
	#board button:not([disabled]), #board a.tile.button:not(.disabled) { color: transparent; }
	#board button:not([disabled]):hover , #board a.tile.button:not(.disabled):hover { color: #888; text-transform: uppercase; }
	#board button.highlighted, #board a.tile.button.highlighted { color: chocolate; }
    img#parrot {
        -webkit-transform: scaleX(-1);
        transform: scaleX(-1);
        padding-left: 10px; 
		margin-right: 0.5em;
    }
</style>
`
}
