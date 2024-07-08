package tictactoe

import (
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestGame(t *testing.T) {
	var (
		addr1 = testutils.TestAddress("addr1")
		addr2 = testutils.TestAddress("addr2")
		addr3 = testutils.TestAddress("addr3")
	)

	game, err := NewGame(addr1, addr1)
	assertErr(t, err)

	game, err = NewGame(addr2, addr3)
	assertNoErr(t, err)
	assertFalse(t, game.IsOver())
	assertFalse(t, game.IsDraw())
	assertErr(t, game.Play(addr3, 0, 0))                                 // addr2's turn
	assertErr(t, game.Play(addr2, -1, 0))                                // invalid location
	assertErr(t, game.Play(addr2, 3, 0))                                 // invalid location
	assertErr(t, game.Play(addr2, 0, -1))                                // invalid location
	uassert.ErrorContains(t, game.Play(addr2, 0, 3), "game not started") // invalid location
	// assertNoErr(t,
	game.start()
	uassert.NoError(t, game.Play(addr2, 1, 1)) // first move
	assertErr(t, game.Play(addr2, 2, 2))       // addr3's turn
	assertErr(t, game.Play(addr3, 1, 1))       // slot already used
	assertNoErr(t, game.Play(addr3, 0, 0))     // second move
	assertNoErr(t, game.Play(addr2, 1, 2))     // third move
	assertNoErr(t, game.Play(addr3, 0, 1))     // fourth move
	assertFalse(t, game.IsOver())
	assertNoErr(t, game.Play(addr2, 1, 0)) // fifth move (win)
	assertTrue(t, game.IsOver())
	assertFalse(t, game.IsDraw())

	expected := `
XO-
XO-
-O-
`
	got := game.Output()
	assertEqualString(t, expected, got)
}

func assertEqualString(t *testing.T, expected, got string) {
	t.Helper()
	expected = strings.TrimSpace(expected)
	got = strings.TrimSpace(got)
	if expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got err: %v", err)
	}
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error, got nil")
	}
}

func assertTrue(t *testing.T, val bool) {
	t.Helper()
	if !val {
		t.Errorf("expected true, got false")
	}
}

func assertFalse(t *testing.T, val bool) {
	t.Helper()
	if val {
		t.Errorf("expected false, got true")
	}
}
