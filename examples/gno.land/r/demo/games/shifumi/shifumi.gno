package shifumi

import (
	"chain/runtime"
	"errors"
	"strconv"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"

	"gno.land/r/sys/users"
)

const (
	empty = iota
	rock
	paper
	scissors
	last
)

type game struct {
	player1, player2 address // shifumi is a 2 players game
	move1, move2     int     // can be empty, rock, paper, or scissors
}

var games avl.Tree
var id seqid.ID

func (g *game) play(player address, move int) error {
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
func NewGame(_ realm, player address) int {
	games.Set(id.Next().String(), &game{player1: runtime.PreviousRealm().Address(), player2: player})
	return int(id)
}

// Play executes a move for the game at index idx, where move can be:
// 1 (rock), 2 (paper), 3 (scissors).
func Play(_ realm, idx, move int) {
	v, ok := games.Get(seqid.ID(idx).String())
	if !ok {
		panic("game not found")
	}
	if err := v.(*game).play(runtime.PreviousRealm().Address(), move); err != nil {
		panic(err)
	}
}

func Render(path string) string {
	mov1 := []string{"", " 🤜  ", " 🫱  ", " 👉  "}
	mov2 := []string{"", " 🤛  ", " 🫲  ", " 👈  "}
	win := []string{"pending", "draw", "player1", "player2"}

	output := `# 👊  ✋  ✌️  Shifumi
Actions:
* [NewGame](shifumi$help&func=NewGame) opponentAddress
* [Play](shifumi$help&func=Play) gameIndex move (1=rock, 2=paper, 3=scissors)

 game  | player1 |     | player2 |       | win 
 --- | --- | --- | --- | --- | ---
`
	// Output the 100 most recent games.
	maxGames := 100
	for n := int(id); n > 0 && int(id)-n < maxGames; n-- {
		v, ok := games.Get(seqid.ID(n).String())
		if !ok {
			continue
		}
		g := v.(*game)
		output += strconv.Itoa(n) + " | " +
			shortName(g.player1) + " | " + mov1[g.move1] + " | " +
			shortName(g.player2) + " | " + mov2[g.move2] + " | " +
			win[g.winner()+1] + "\n"
	}
	return output
}

func shortName(addr address) string {
	user := users.ResolveAddress(addr)
	if user != nil {
		return user.Name()
	}
	if len(addr) < 10 {
		return string(addr)
	}
	return string(addr)[:10] + "..."
}
