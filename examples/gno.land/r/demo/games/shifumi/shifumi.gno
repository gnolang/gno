package shifumi

import (
	"errors"
	"std"
	"strconv"
)

const (
	empty = iota
	rock
	paper
	scissors
	last
)

type game struct {
	player1, player2 std.Address  // shifumi is a 2 players game
	move1, move2 int              // can be empty, rock, paper, or scissors
}

var games = []*game{}

func (g *game) play(player std.Address, move int) error {
	if !(move > empty && move < last) {
		return errors.New("invalid move")
	}
	if player != g.player1 && player != g.player2 {
		return errors.New("invalid player")
	}
	if player == g.player1 && g.move1 == empty {
		g.move1 = move
		return nil
	}
	if player == g.player2 && g.move2 == empty {
		g.move2 = move
		return nil
	}
	return errors.New("already played")
}

func (g *game) winner() int {
	if g.move1 == empty || g.move2 == empty {
		return -1
	}
	if g.move1 == g.move2 {
		return 0
	}
	if g.move1 == rock && g.move2 == scissors ||
		g.move1 == paper && g.move2 == rock ||
		g.move1 == scissors && g.move2 == paper {
		return 1
	}
	return 2
}

// NewGame creates a new game where player1 is the caller and player2 the argument.
// A new game index is returned.
func NewGame(player std.Address) int {
	games = append(games, &game{player1: std.GetOrigCaller(), player2: player})
	return len(games)-1
}

// Play executes a move for the game at index idx, where move can be:
// 1 (rock), 2 (paper), 3 (scissors).
func Play(idx, move int) {
	if err := games[idx].play(std.GetOrigCaller(), move); err != nil {
		panic(err)
	}
}

func Render(path string) string {
	mov1 := []string{ "", " 🤜  ", " 🫱  ", " 👉  "}
	mov2 := []string{ "", " 🤛  ", " 🫲  ", " 👈  "}
	win := []string{ "pending", "draw", "player1", "player2" }

	output := `# 👊  ✋  ✌️  Shifumi
Actions:
* [NewGame](r/demo/games/shifumi?help&__func=NewGame) opponentAddress
* [Play](r/demo/games/shifumi?help&__func=Play) gameIndex move (1=rock, 2=paper, 3=scissors)

 game  | player1 |     | player2 |       | win 
 --- | --- | --- | --- | --- | ---
`
	for i := len(games)-1; i >= 0; i-- {
		g := games[i]
		output += strconv.Itoa(i) + " | " +
			short(g.player1) + " | " + mov1[g.move1] + " | " +
			short(g.player2) + " | " + mov2[g.move2] + " | " +
			win[g.winner()+1] + "\n"
	}
	return output
}

func short(addr std.Address) string {
	if len(addr) < 10 {
		return string(addr)
	}
	return string(addr)[:10] + "..."
}
