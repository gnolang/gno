package drand

import (
	"encoding/hex"
	"strings"
	"testing"

	evmnet "gno.land/p/drand/v0"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// Real evmnet beacons (same as p/drand/v0/vectors_test.gno).
const (
	sig2           = "1dbfe644485f31e2a7978fb551b50933b0001e64397fd2b7dc4022e8e2fcdbc325798cea5506deec489c8db3b4697f6b1bb6be8934b95bd9c0db88923487d9ad"
	sig1000        = "06fd5996329504d3a56b482d9222bf7205857d0a9559ddd216ca31a286f6a8cc0a120f021aac2f13553fb164f62bc3a5ca32c76dea88a777b39bcf3cac5fdbd6"
	randomness1000 = "0e6745667465a6f9dce5d5f994656955080be14c469ff17fc4fc588c925a8504"
)

var (
	alice   = testutils.TestAddress("alice")
	bob     = testutils.TestAddress("bob")
	relayer = testutils.TestAddress("relayer")
)

func reset() {
	requests = avl.Tree{}
	pending = avl.Tree{}
	beacons = avl.Tree{}
}

// before returns a unix time that pins RequestAfter to round r. Test block
// time is 2009, before evmnet genesis, so every real round is "future".
func before(r uint64) int64 { return evmnet.TimeOf(r) - 1 }

func TestRequestSubmitGet(cur realm, t *testing.T) {
	reset()

	testing.SetRealm(testing.NewUserRealm(alice))
	id := RequestAfter(cross(cur), before(1000))

	round, ready := Status(id)
	uassert.Equal(t, uint64(1000), round)
	uassert.False(t, ready)
	_, ok := Get(id)
	uassert.False(t, ok, "no randomness before submit")
	uassert.Equal(t, "1000", PendingRounds(10))
	uassert.PanicsWithMessage(t, cur, "randomness not available yet", func() { MustGet(id) })
	uassert.PanicsWithMessage(t, cur, "randomness not available yet", func() { Rand(id) })

	testing.SetRealm(testing.NewUserRealm(relayer))
	Submit(cross(cur), 1000, sig1000)

	_, ready = Status(id)
	uassert.True(t, ready)
	uassert.Equal(t, "", PendingRounds(10))

	raw, ok := Beacon(1000)
	uassert.True(t, ok)
	uassert.Equal(t, randomness1000, hex.EncodeToString(raw[:]))

	v := MustGet(id)
	want := evmnet.Derive(raw, alice.String()+":"+id)
	uassert.True(t, v == want, "per-request derivation")
	uassert.True(t, v != raw, "request value differs from the raw beacon")
}

func TestSubmitRejects(cur realm, t *testing.T) {
	reset()
	testing.SetRealm(testing.NewUserRealm(relayer))
	uassert.AbortsWithMessage(t, cur, "round not requested", func() { Submit(cross(cur), 1000, sig1000) })

	testing.SetRealm(testing.NewUserRealm(alice))
	RequestAfter(cross(cur), before(1000))

	testing.SetRealm(testing.NewUserRealm(relayer))
	uassert.AbortsWithMessage(t, cur, "signature is not hex", func() { Submit(cross(cur), 1000, "zz") })
	uassert.AbortsWithMessage(t, cur, "invalid drand signature", func() { Submit(cross(cur), 1000, sig2) })
	uassert.AbortsWithMessage(t, cur, "invalid drand signature", func() { Submit(cross(cur), 1000, sig1000[:120]) })

	bad := []byte(sig1000)
	if bad[10] == '0' {
		bad[10] = '1'
	} else {
		bad[10] = '0'
	}
	uassert.AbortsWithMessage(t, cur, "invalid drand signature", func() { Submit(cross(cur), 1000, string(bad)) })

	_, ok := Beacon(1000)
	uassert.False(t, ok, "nothing stored after rejected submits")
	uassert.Equal(t, "1000", PendingRounds(10))
}

func TestSubmitTwiceIsNoop(cur realm, t *testing.T) {
	reset()
	testing.SetRealm(testing.NewUserRealm(alice))
	RequestAfter(cross(cur), before(1000))

	testing.SetRealm(testing.NewUserRealm(relayer))
	Submit(cross(cur), 1000, sig1000)
	uassert.NotAborts(t, cur, func() { Submit(cross(cur), 1000, sig1000) })
	// Even a garbage resubmit of a stored round is ignored, not an error.
	uassert.NotAborts(t, cur, func() { Submit(cross(cur), 1000, "zz") })
}

func TestSharedRoundIndependentValues(cur realm, t *testing.T) {
	reset()
	testing.SetRealm(testing.NewUserRealm(alice))
	a1 := RequestAfter(cross(cur), before(1000))
	a2 := RequestAfter(cross(cur), before(1000))
	testing.SetRealm(testing.NewUserRealm(bob))
	b1 := RequestAfter(cross(cur), before(1000))

	uassert.Equal(t, "1000", PendingRounds(10), "one pending entry per round")

	testing.SetRealm(testing.NewUserRealm(relayer))
	Submit(cross(cur), 1000, sig1000)

	va1, vb1, va2 := MustGet(a1), MustGet(b1), MustGet(a2)
	uassert.True(t, va1 != va2, "same requester, different ids")
	uassert.True(t, va1 != vb1, "different requesters")
}

func TestRequestAfterRoundsAndOrder(cur realm, t *testing.T) {
	reset()
	testing.SetRealm(testing.NewUserRealm(alice))
	idLate := RequestAfter(cross(cur), before(1000))
	idEarly := RequestAfter(cross(cur), before(2))
	// Exactly at a round's publication time: that round is not "after".
	idAt := RequestAfter(cross(cur), evmnet.TimeOf(5))

	r, _ := Status(idEarly)
	uassert.Equal(t, uint64(2), r)
	r, _ = Status(idAt)
	uassert.Equal(t, uint64(6), r)
	r, _ = Status(idLate)
	uassert.Equal(t, uint64(1000), r)

	uassert.Equal(t, "2,6,1000", PendingRounds(0), "sorted numerically")
	uassert.Equal(t, "2,6", PendingRounds(2))
}

func TestRequestDefaultIsSafeRound(cur realm, t *testing.T) {
	reset()
	testing.SetRealm(testing.NewUserRealm(alice))
	id := Request(cross(cur))
	r, _ := Status(id)
	uassert.Equal(t, nextSafeRound(), r)
	// A RequestAfter in the past is pulled forward to the safe round.
	id = RequestAfter(cross(cur), 0)
	r2, _ := Status(id)
	uassert.Equal(t, r, r2)
}

func TestUnknownID(t *testing.T) {
	reset()
	_, ok := Get("nope")
	uassert.False(t, ok)
	r, ready := Status("nope")
	uassert.Equal(t, uint64(0), r)
	uassert.False(t, ready)
	uassert.Equal(t, "# request not found\n", Render("request/nope"))
}

func TestRender(cur realm, t *testing.T) {
	reset()
	testing.SetRealm(testing.NewUserRealm(alice))
	id := RequestAfter(cross(cur), before(1000))

	out := Render("")
	uassert.True(t, strings.Contains(out, "- requests: 1"), out)
	uassert.True(t, strings.Contains(out, "- pending rounds: 1"), out)
	uassert.True(t, strings.Contains(Render("request/"+id), "randomness: pending"))

	testing.SetRealm(testing.NewUserRealm(relayer))
	Submit(cross(cur), 1000, sig1000)
	out = Render("")
	uassert.True(t, strings.Contains(out, "| 1000 | `"+randomness1000+"` |"), out)
	v := MustGet(id)
	uassert.True(t, strings.Contains(Render("request/"+id), hex.EncodeToString(v[:])))
}

func TestRand(cur realm, t *testing.T) {
	reset()
	testing.SetRealm(testing.NewUserRealm(alice))
	id := RequestAfter(cross(cur), before(1000))
	testing.SetRealm(testing.NewUserRealm(relayer))
	Submit(cross(cur), 1000, sig1000)

	// Same seed as NewRand(MustGet(id)), and stable across calls.
	r, want, again := Rand(id), evmnet.NewRand(MustGet(id)), Rand(id)
	for i := 0; i < 20; i++ {
		v := r.IntN(1_000_000)
		uassert.Equal(t, want.IntN(1_000_000), v)
		uassert.Equal(t, again.IntN(1_000_000), v)
	}
	// Values stay in range and are not all equal.
	seen := map[int]bool{}
	r = Rand(id)
	for i := 0; i < 50; i++ {
		v := r.IntN(6)
		uassert.True(t, v >= 0 && v < 6)
		seen[v] = true
	}
	uassert.True(t, len(seen) > 1, "a die that always rolls the same face")
}
