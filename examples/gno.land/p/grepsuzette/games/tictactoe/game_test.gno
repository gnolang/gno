package tictactoe

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	addr1 = testutils.TestAddress("addr1")
	addr2 = testutils.TestAddress("addr2")
	addr3 = testutils.TestAddress("addr3")
)

func TestGame(t *testing.T) {
	game, err := NewGame(addr1, addr1)
	uassert.Error(t, err)

	game, err = NewGame(addr2, addr3)
	uassert.NoError(t, err)

	uassert.False(t, game.IsOver())
	uassert.False(t, game.IsDraw())
	game.Start()
	uassert.Error(t, game.Play(addr3, 0, 0))   // addr2's turn
	uassert.Error(t, game.Play(addr2, -1, 0))  // invalid location
	uassert.Error(t, game.Play(addr2, 3, 0))   // invalid location
	uassert.Error(t, game.Play(addr2, 0, -1))  // invalid location
	uassert.Error(t, game.Play(addr2, 0, 3))   // invalid location
	uassert.NoError(t, game.Play(addr2, 1, 1)) // first move
	uassert.Error(t, game.Play(addr2, 2, 2))   // addr3's turn
	uassert.Error(t, game.Play(addr3, 1, 1))   // slot already used
	uassert.NoError(t, game.Play(addr3, 0, 0)) // second move
	uassert.NoError(t, game.Play(addr2, 1, 2)) // third move
	uassert.NoError(t, game.Play(addr3, 0, 1)) // fourth move
	uassert.False(t, game.IsOver())
	uassert.NoError(t, game.Play(addr2, 1, 0)) // fifth move (win)
	uassert.True(t, game.IsOver())
	uassert.False(t, game.IsDraw())

	expected := `-O-
XO-
XO-
`
	got := game.Output()
	uassert.Equal(t, expected, got)
}

func TestRecoverGame(t *testing.T) {
	for _, o := range []struct {
		repr, err string
	}{
		{"", "error"},
		{"--", "error"},
		{"---", "error"},
		{"-----", "error"},
		{"--------", "error"},
		{"---------", ""},
		{"XX-------", "error"},
		{"OO-------", "error"},
		{"XO-X-----", "error"}, // O plays first
		{"XO-O-----", ""},      // valid from there on
		{"XOXO-----", ""},
		{"XOXOO----", ""},
		{"XOXOO-X--", ""},
		{"XOXOOOX--", ""}, // circles won but the function doesn't care
		{"XOXOOOX-X", ""},
		{"XOXOOOXOX", ""},       // circles won a second time
		{"XOXOOOXOXX", "error"}, // too long (10 squares)
	} {
		g, e := RecoverGame(addr1, addr2, o.repr)
		if o.err == "error" {
			uassert.Error(t, e, "repr=", o.repr)
		} else {
			uassert.NoError(t, e, "repr=", o.repr)
			uassert.True(t, g != nil, "repr=", o.repr)
		}
	}
}
