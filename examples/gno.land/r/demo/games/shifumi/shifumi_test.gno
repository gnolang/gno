package shifumi

import (
	"math/rand"
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/entropy"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestShifumi(t *testing.T) {
	t.Fatal(hashMove(2, "1"))
	player1 := testutils.TestAddress("alice")
	player2 := testutils.TestAddress("bob")

	std.TestSetOrigCaller(player1)
	gameId := NewGame(player2)

	// Verify game creation
	v, ok := games.Get(seqid.ID(gameId).String())
	if !ok {
		t.Fatal("game not found")
	}
	g := v.(*game)

	urequire.Equal(t, player1.String(), g.player1.String(), ufmt.Sprintf("expected player1 to be %s, got %s", player1.String(), g.player1.String()))
	urequire.Equal(t, player2.String(), g.player2.String(), ufmt.Sprintf("expected player2 to be %s, got %s", player2.String(), g.player2.String()))

	// init randomness
	seed := uint64(entropy.New().Seed())
	r := rand.New(rand.NewPCG(seed, 0xdeadbeef))

	// Player1 commit move (rock)
	std.TestSetOrigCaller(player1)
	move1 := rock
	salt1 := strconv.Itoa(r.IntN(1000))

	// This hash is done off-chain
	commit1 := hashMoveOffchain(move1, salt1)
	CommitMove(gameId, commit1)

	// Player2 commit move (scissors)
	std.TestSetOrigCaller(player2)
	move2 := scissors
	salt2 := strconv.Itoa(r.IntN(1000))

	// This hash is done off-chain
	commit2 := hashMoveOffchain(move2, salt2)
	CommitMove(gameId, commit2)

	// Player1 reveal move
	std.TestSetOrigCaller(player1)
	RevealMove(gameId, move1, salt1)

	// Player2 reveal move
	std.TestSetOrigCaller(player2)
	RevealMove(gameId, move2, salt2)

	// Verify moves
	v, ok = games.Get(seqid.ID(gameId).String())
	urequire.True(t, ok, "game not found")

	// Verify winner
	g = v.(*game)
	urequire.Equal(t, rock, g.move1, ufmt.Sprintf("expected move1 to be %d, got %d", rock, g.move1))
	urequire.Equal(t, scissors, g.move2, ufmt.Sprintf("expected move2 to be %d, got %d", scissors, g.move2))
	urequire.Equal(t, 1, g.winner(), ufmt.Sprintf("expected player1 to win, but result was %d", g.winner()))
}

// hashMoveOffchain function should be executed off-chain by the player,
// and it must implement the same logic as this on-chain function to ensure
// consistency
func hashMoveOffchain(move int, salt string) string {
	return hashMove(move, salt)
}
