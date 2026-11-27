package vesting

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// lcg is a fixed-seed generator so a failure can be replayed.
type lcg struct{ s uint64 }

func (r *lcg) next() uint64 {
	r.s = r.s*6364136223846793005 + 1
	return r.s
}

func (r *lcg) intn(n int64) int64 {
	if n <= 0 {
		return 0
	}
	return int64(r.next() % uint64(n))
}

func TestPropertyInvariants(t *testing.T) {
	rng := lcg{s: 0x123456789abcdef0}
	admin0 := testutils.TestAddress("padmin")
	ben0 := testutils.TestAddress("pben")
	for trial := 0; trial < 40; trial++ {
		terms := randomTerms(&rng)
		revocable := rng.intn(2) == 0
		g, err := New(hostPkg, hostAddr(), admin0, ben0, coinAsset(), terms, revocable, "")
		uassert.NoError(t, err)
		_, err = g.Bind(int64(1000 + trial))
		uassert.NoError(t, err)

		admin := admin0
		ben := ben0
		now := terms.Start - 5
		var prevV int64
		var prevNow int64 = -1
		for step := 0; step < 12; step++ {
			now += 1 + rng.intn(terms.Duration+2)
			v, err := g.Vested(now)
			uassert.NoError(t, err)
			rel, err := g.Releasable(now)
			uassert.NoError(t, err)
			if prevNow >= 0 && now >= prevNow {
				uassert.True(t, v >= prevV)
			}
			prevV = v
			prevNow = now
			checkBooks(t, g, v, rel)

			switch rng.intn(5) {
			case 0:
				got, cerr := g.Claim(ben, now)
				if rel == 0 {
					uassert.ErrorIs(t, cerr, ErrNothing)
				} else {
					uassert.NoError(t, cerr)
					uassert.Equal(t, rel, got)
					uassert.True(t, got <= v)
				}
			case 1:
				if !g.Revoked() && g.Revocable() {
					want, rerr := g.Reclaimable(now)
					uassert.NoError(t, rerr)
					got, rerr := g.Revoke(admin, now)
					uassert.NoError(t, rerr)
					uassert.Equal(t, want, got)
					uassert.True(t, g.Claimed()+got <= terms.Total)
				} else if g.Revoked() {
					_, rerr := g.Revoke(admin, now)
					uassert.ErrorIs(t, rerr, ErrRevoked)
				} else {
					_, rerr := g.Revoke(admin, now)
					uassert.ErrorIs(t, rerr, ErrIrrevocable)
				}
			case 2:
				_, cerr := g.Claim(admin, now)
				if admin != ben {
					uassert.ErrorIs(t, cerr, ErrNotBeneficiary)
				}
			case 3:
				nxt := testutils.TestAddress("b" + itoaSmall(rng.intn(20)))
				if nxt != ben {
					err = g.SetBeneficiary(ben, nxt)
					uassert.NoError(t, err)
					ben = nxt
				}
			default:
				nxt := testutils.TestAddress("a" + itoaSmall(rng.intn(20)))
				if nxt != admin {
					err = g.TransferAdmin(admin, nxt)
					uassert.NoError(t, err)
					admin = nxt
				}
			}
			v, err = g.Vested(now)
			uassert.NoError(t, err)
			rel, err = g.Releasable(now)
			uassert.NoError(t, err)
			checkBooks(t, g, v, rel)
		}
		end, err := terms.End()
		uassert.NoError(t, err)
		v, err := g.Vested(end + 10_000)
		uassert.NoError(t, err)
		if !g.Revoked() {
			uassert.Equal(t, terms.Total, v)
		}
		rel, err := g.Releasable(end + 10_000)
		uassert.NoError(t, err)
		checkBooks(t, g, v, rel)
		if rel > 0 {
			got, err := g.Claim(ben, end+10_000)
			uassert.NoError(t, err)
			uassert.Equal(t, rel, got)
		}
		v, err = g.Vested(end + 10_000)
		uassert.NoError(t, err)
		rel, err = g.Releasable(end + 10_000)
		uassert.NoError(t, err)
		uassert.Equal(t, int64(0), rel)
		checkBooks(t, g, v, rel)
	}
}

func checkBooks(t *testing.T, g *Grant, vested, releasable int64) {
	t.Helper()
	total := g.Schedule().Total
	uassert.True(t, g.Claimed() <= vested)
	uassert.True(t, vested <= total)
	uassert.True(t, releasable >= 0)
	uassert.Equal(t, vested-g.Claimed(), releasable)
	if g.Revoked() {
		uassert.Equal(t, total, g.Claimed()+releasable+g.Returned())
		return
	}
	uassert.Equal(t, int64(0), g.Returned())
	uassert.Equal(t, total, g.Claimed()+releasable+(total-vested))
}

func randomTerms(rng *lcg) Terms {
	start := int64(1_000_000) + rng.intn(1_000_000)
	duration := int64(1)
	switch rng.intn(4) {
	case 0:
		duration = 1
	case 1:
		duration = 2 + rng.intn(30)
	default:
		duration = 31 + rng.intn(400)
	}
	cliff := int64(0)
	switch rng.intn(3) {
	case 0:
		cliff = 0
	case 1:
		cliff = duration
	default:
		cliff = rng.intn(duration + 1)
	}
	total := int64(1)
	switch rng.intn(4) {
	case 0:
		total = 1
	case 1:
		total = 2 + rng.intn(50)
	default:
		total = 51 + rng.intn(1_000_000)
	}
	curve := CurveLinear
	period := int64(0)
	if rng.intn(2) == 0 {
		curve = CurveStep
		period = 1
		if duration > 1 && rng.intn(2) == 0 {
			// Pick a divisor. Walk candidates; duration is small.
			want := 1 + rng.intn(duration)
			for d := want; d >= 1; d-- {
				if duration%d == 0 {
					period = d
					break
				}
			}
		}
	}
	return Terms{Start: start, Cliff: cliff, Duration: duration, Period: period, Total: total, Curve: curve}
}

func itoaSmall(n int64) string {
	if n < 0 {
		n = 0
	}
	if n > 99 {
		n = 99
	}
	if n < 10 {
		return string(byte('0' + n))
	}
	return string([]byte{byte('0' + n/10), byte('0' + n%10)})
}
