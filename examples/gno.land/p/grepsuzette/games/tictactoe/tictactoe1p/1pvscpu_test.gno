package tictactoe1p

import (
	"math/rand"
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

const (
	cpuAddress   = std.Address("CPU")
	humanAddress = std.Address("HUMAN")
)

var prng = rand.New(rand.NewPCG(uint64(3), uint64(7)))

func TestRand(t *testing.T) {
	n0 := prng.IntN(10)
	for i := 0; i < 128; i++ {
		if prng.IntN(10) != n0 {
			return
		}
	}
	t.Errorf("no randomness")
}

func TestGameFromRepr(t *testing.T) {
	for _, o := range []struct{ repr, expect string }{
		{repr: "", expect: "---------"},
		{repr: "-X--OX-O-", expect: "-X--OX-O-"},
		{repr: "_x__ox_o_", expect: "-X--OX-O-"},
		{repr: "O--------", expect: "O--------"},
		{repr: "O---X----", expect: "O---X----"},
		{repr: "-X--O--O-", expect: "-X--O--O-"},
		{repr: "OXOXOXOXO", expect: "OXOXOXOXO"},
		{repr: "X--------", expect: "error"}, // O always begin
		{repr: "OOOOOX---", expect: "error"},
	} {
		game, e := GameFromRepr(o.repr, cpuAddress, humanAddress, prng.IntN, nil)
		if o.expect == "error" {
			uassert.Error(t, e)
		} else {
			uassert.NoError(t, e)
			uassert.Equal(t, o.expect, game.ToRepr())
		}
	}
}

func TestPlayer1Alternates(t *testing.T) {
	// ensure CPU goes first or second, in different games
	var isCpuFirst0 bool
	for i := 0; i < 128; i++ {
		g, e := GameFromRepr("", cpuAddress, humanAddress, prng.IntN, nil)
		uassert.NoError(t, e)
		if i == 0 {
			isCpuFirst0 = g.IsCpuFirst()
		} else if isCpuFirst0 != g.IsCpuFirst() {
			return // ok
		}
	}
	t.FailNow()
}

func TestPickMove(t *testing.T) {
	// study expected VS statistical result
	// CPU sees own winning move ~80% of times
	//     is otherwise blind to other side
	for _, o := range []struct {
		board     string
		expectedX int
		expectedY int
	}{
		{"O-OX-X---", 1, 2},
		{"O-OX-XO--", 1, 1},
		{"X-O-XOO--", 2, 0},
		{"OX-XO----", 2, 0},
	} {
		// play over and over, see most frequent moves
		h := map[struct{ x, y int }]int{}
		for i := 0; i < 64; i++ {
			game, e := GameFromRepr(o.board, cpuAddress, humanAddress, prng.IntN, nil)
			if !uassert.NoError(t, e, "GameFromRepr") {
				t.FailNow()
			}
			x, y, e := game.pickMove(game.Game, game.Intn)
			h[struct{ x, y int }{x, y}] += 1
		}
		// assert most frequent move
		var hiTemperature int
		var hotMove struct{ x, y int }
		for coord, v := range h {
			if v > hiTemperature {
				hiTemperature = v
				hotMove = coord
			}
		}
		uassert.Equal(t, o.expectedX, hotMove.x, "for x of repr", o.board)
		uassert.Equal(t, o.expectedY, hotMove.y, "for y of repr", o.board)
	}
}
