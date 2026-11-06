package coinflip

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/r/drand/v0"
)

// evmnet round 1. Test block time predates evmnet genesis, so the next safe
// round is round 1.
const sig1 = "11f812d738a36b2210dc88c2d635ad8039588205f42445d6de09e6530165c3462a23aca348c84badcf8df5321ac24577b7963d5b0d780bc4626baedb45cde373"

func TestFlipThroughDrand(cur realm, t *testing.T) {
	player := testutils.TestAddress("player")
	testing.SetRealm(testing.NewUserRealm(player))

	heads := Flip(cross(cur), true)
	tails := Flip(cross(cur), false)

	uassert.AbortsWithMessage(t, cur, "randomness not available yet", func() { Settle(cross(cur), heads) })

	// The drand request was made by this realm, not by the player.
	f := get(heads)
	round, ready := drand.Status(f.reqID)
	uassert.Equal(t, uint64(1), round)
	uassert.False(t, ready)
	uassert.True(t, strings.Contains(drand.Render("request/"+f.reqID), chain.PackageAddress("gno.land/r/drand/coinflip/v0").String()))

	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("relayer")))
	drand.Submit(cross(cur), 1, sig1)

	Settle(cross(cur), heads)
	Settle(cross(cur), tails)
	uassert.AbortsWithMessage(t, cur, "already settled", func() { Settle(cross(cur), heads) })

	s1, wonHeads := Result(heads)
	s2, wonTails := Result(tails)
	uassert.True(t, s1 && s2)

	// Each flip uses its own derived value, so the outcome must match it.
	uassert.Equal(t, drand.Rand(get(heads).reqID).IntN(2) == 0, wonHeads)
	uassert.Equal(t, drand.Rand(get(tails).reqID).IntN(2) == 1, wonTails)

	out := Render("")
	uassert.True(t, strings.Contains(out, "2 flips"), out)
}
