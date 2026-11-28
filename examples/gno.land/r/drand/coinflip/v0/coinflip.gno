// Package coinflip is a minimal consumer of r/drand/v0: call Flip with a
// guess, then Settle once the drand round is in.
//
// It shows the one rule every consumer must follow: record the player's
// choice first, then request randomness, in the same transaction. The round
// the request is pinned to is still unpublished, so nobody can pick a guess
// knowing the outcome.
package coinflip

import (
	"strconv"
	"strings"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/drand/v0"
)

type flip struct {
	player  address
	heads   bool
	reqID   string
	settled bool
	won     bool
}

var (
	ids   seqid.ID
	flips avl.Tree // seqid -> *flip
)

// Flip records a guess and requests drand. It returns the flip id.
func Flip(cur realm, heads bool) string {
	f := &flip{player: cur.Previous().Address(), heads: heads}
	f.reqID = drand.Request(cross(cur)) // after the guess is stored
	id := ids.Next().String()
	flips.Set(id, f)
	return id
}

// Settle resolves a flip once its randomness is available. Anyone may call
// it; the outcome only depends on the drand beacon.
func Settle(cur realm, id string) {
	f := get(id)
	if f.settled {
		panic("already settled")
	}
	f.won = (drand.Rand(f.reqID).IntN(2) == 0) == f.heads // 0 = heads
	f.settled = true
}

// Result returns whether the flip is settled and whether the player won.
func Result(id string) (settled, won bool) {
	f := get(id)
	return f.settled, f.won
}

func get(id string) *flip {
	v := flips.Get(id)
	if v == nil {
		panic("unknown flip")
	}
	return v.(*flip)
}

func Render(path string) string {
	var b strings.Builder
	b.WriteString("# coinflip\n\nA demo consumer of [drand](/r/drand/v0).\n\n")
	b.WriteString("| id | player | guess | result |\n|---|---|---|---|\n")
	n := 0
	flips.ReverseIterate("", "", func(id string, v any) bool {
		f := v.(*flip)
		guess := "tails"
		if f.heads {
			guess = "heads"
		}
		result := "pending"
		if f.settled {
			result = "lost"
			if f.won {
				result = "won"
			}
		}
		b.WriteString(ufmt.Sprintf("| %s | %s | %s | %s |\n", id, f.player.String(), guess, result))
		n++
		return n >= 20
	})
	b.WriteString("\n" + strconv.Itoa(flips.Size()) + " flips\n")
	return b.String()
}
