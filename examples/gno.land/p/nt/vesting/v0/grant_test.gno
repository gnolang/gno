package vesting

import (
	"strconv"
	"strings"
	"testing"

	"chain"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

const hostPkg = "gno.land/r/nt/vesting/v0"

func hostAddr() address { return chain.PackageAddress(hostPkg) }

func coinAsset() Asset { return Asset{Kind: AssetCoin, Denom: "ugnot"} }

func linear(start, cliff, duration, total int64) Terms {
	return Terms{Start: start, Cliff: cliff, Duration: duration, Total: total, Curve: CurveLinear}
}

func openGrant(t *testing.T, admin, ben address, terms Terms, revocable bool, id int64) *Grant {
	t.Helper()
	g, err := New(hostPkg, hostAddr(), admin, ben, coinAsset(), terms, revocable, "grant")
	uassert.NoError(t, err)
	addr, err := g.Bind(id)
	uassert.NoError(t, err)
	uassert.Equal(t, chain.DerivePkgSubAddr(hostPkg, "g/"+strconv.FormatInt(id, 10)), addr)
	return g
}

func TestGrantLifecycle(t *testing.T) {
	admin := testutils.TestAddress("admin")
	ben := testutils.TestAddress("ben")
	next := testutils.TestAddress("next")
	const start = int64(5_000)
	g := openGrant(t, admin, ben, linear(start, 100, 1_000, 1_000), true, 1)

	v, err := g.Vested(start + 50)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), v)
	_, err = g.Claim(ben, start+50)
	uassert.ErrorIs(t, err, ErrNothing)
	back, err := g.Reclaimable(start + 50)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(1000), back)
	got, err := g.Revoke(admin, start+50)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(1000), got)
	uassert.True(t, g.Revoked())
	_, err = g.Claim(ben, start+10_000)
	uassert.ErrorIs(t, err, ErrNothing)
	_, err = g.Revoke(admin, start+10_000)
	uassert.ErrorIs(t, err, ErrRevoked)

	g = openGrant(t, admin, ben, linear(start, 100, 1_000, 1_000), true, 2)
	got, err = g.Revoke(admin, start+100)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(900), got)
	uassert.Equal(t, int64(100), g.Claimed()+mustRel(t, g, start+500))
	claimed, err := g.Claim(ben, start+500)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(100), claimed)
	_, err = g.Claim(ben, start+9_000)
	uassert.ErrorIs(t, err, ErrNothing)
	uassert.Equal(t, int64(100), g.Claimed())
	uassert.Equal(t, int64(900), g.Returned())
	out, err := g.Outstanding()
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), out)

	g = openGrant(t, admin, ben, linear(start, 0, 100, 100), true, 3)
	claimed, err = g.Claim(ben, start+100)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(100), claimed)
	_, err = g.Claim(ben, start+200)
	uassert.ErrorIs(t, err, ErrNothing)
	got, err = g.Revoke(admin, start+200)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), got)

	g = openGrant(t, admin, ben, linear(start, 0, 100, 100), false, 4)
	_, err = g.Revoke(admin, start+10)
	uassert.ErrorIs(t, err, ErrIrrevocable)
	err = g.TransferAdmin(admin, next)
	uassert.NoError(t, err)
	uassert.Equal(t, next, g.Admin())
	_, err = g.Revoke(next, start+10)
	uassert.ErrorIs(t, err, ErrIrrevocable)

	g = openGrant(t, admin, ben, linear(start, 0, 100, 100), true, 5)
	err = g.SetBeneficiary(ben, next)
	uassert.NoError(t, err)
	_, err = g.Claim(ben, start+100)
	uassert.ErrorIs(t, err, ErrNotBeneficiary)
	claimed, err = g.Claim(next, start+100)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(100), claimed)
	err = g.TransferAdmin(admin, ben)
	uassert.NoError(t, err)
	_, err = g.Revoke(admin, start+100)
	uassert.ErrorIs(t, err, ErrNotAdmin)
}

func mustRel(t *testing.T, g *Grant, now int64) int64 {
	t.Helper()
	n, err := g.Releasable(now)
	uassert.NoError(t, err)
	return n
}

func TestGrantErrors(t *testing.T) {
	admin := testutils.TestAddress("admin2")
	ben := testutils.TestAddress("ben2")
	other := testutils.TestAddress("other2")
	terms := linear(100, 0, 10, 10)

	_, err := New("gno.land/p/nt/vesting/v0", hostAddr(), admin, ben, coinAsset(), terms, true, "")
	uassert.ErrorIs(t, err, ErrHost)
	_, err = New(hostPkg+"#g/1", hostAddr(), admin, ben, coinAsset(), terms, true, "")
	uassert.ErrorIs(t, err, ErrHost)
	_, err = New(hostPkg, ben, admin, ben, coinAsset(), terms, true, "")
	uassert.ErrorIs(t, err, ErrHost)
	_, err = New(hostPkg, hostAddr(), address(""), ben, coinAsset(), terms, true, "")
	uassert.ErrorIs(t, err, ErrAddress)
	_, err = New(hostPkg, hostAddr(), admin, hostAddr(), coinAsset(), terms, true, "")
	uassert.ErrorIs(t, err, ErrSelf)
	_, err = New(hostPkg, hostAddr(), hostAddr(), ben, coinAsset(), terms, true, "")
	uassert.ErrorIs(t, err, ErrSelf)
	_, err = New(hostPkg, hostAddr(), admin, ben, coinAsset(), terms, true, strings.Repeat("a", MaxLabelLen+1))
	uassert.ErrorIs(t, err, ErrLabel)
	_, err = New(hostPkg, hostAddr(), admin, ben, coinAsset(), terms, true, "bad\nlabel")
	uassert.ErrorIs(t, err, ErrLabel)
	_, err = New(hostPkg, hostAddr(), admin, ben, Asset{Kind: AssetCoin, Denom: "NO"}, terms, true, "")
	uassert.ErrorIs(t, err, ErrDenom)
	_, err = New(hostPkg, hostAddr(), admin, ben, Asset{Kind: AssetCoin, Denom: "ugnot", Token: "x"}, terms, true, "")
	uassert.ErrorIs(t, err, ErrDenom)
	_, err = New(hostPkg, hostAddr(), admin, ben, Asset{Kind: AssetGRC20}, terms, true, "")
	uassert.ErrorIs(t, err, ErrToken)
	_, err = New(hostPkg, hostAddr(), admin, ben, Asset{Kind: AssetGRC20, Token: "has space"}, terms, true, "")
	uassert.ErrorIs(t, err, ErrToken)
	_, err = New(hostPkg, hostAddr(), admin, ben, Asset{Kind: AssetGRC20, Token: "bad#key"}, terms, true, "")
	uassert.ErrorIs(t, err, ErrToken)
	_, err = New(hostPkg, hostAddr(), admin, ben, Asset{Kind: 9}, terms, true, "")
	uassert.ErrorIs(t, err, ErrAsset)
	_, err = New(hostPkg, hostAddr(), admin, ben, coinAsset(), Terms{}, true, "")
	uassert.ErrorIs(t, err, ErrTotal)

	g, err := New(hostPkg, hostAddr(), admin, ben, coinAsset(), terms, true, "")
	uassert.NoError(t, err)
	_, err = g.Claim(ben, 100)
	uassert.ErrorIs(t, err, ErrUnbound)
	_, err = g.Bind(0)
	uassert.ErrorIs(t, err, ErrID)
	_, err = (*Grant)(nil).Bind(1)
	uassert.ErrorIs(t, err, ErrUnbound)
	_, err = g.Bind(7)
	uassert.NoError(t, err)
	_, err = g.Bind(8)
	uassert.ErrorIs(t, err, ErrBound)

	_, err = g.Claim(other, 110)
	uassert.ErrorIs(t, err, ErrNotBeneficiary)
	_, err = g.Revoke(other, 110)
	uassert.ErrorIs(t, err, ErrNotAdmin)
	_, err = g.Vested(-1)
	uassert.ErrorIs(t, err, ErrTime)
	_, err = g.Releasable(-1)
	uassert.ErrorIs(t, err, ErrTime)
	err = g.SetBeneficiary(other, admin)
	uassert.ErrorIs(t, err, ErrNotBeneficiary)
	err = g.SetBeneficiary(ben, address("nope"))
	uassert.ErrorIs(t, err, ErrAddress)
	err = g.SetBeneficiary(ben, ben)
	uassert.ErrorIs(t, err, ErrSame)
	err = g.SetBeneficiary(ben, g.Address())
	uassert.ErrorIs(t, err, ErrSelf)
	err = g.SetBeneficiary(ben, hostAddr())
	uassert.ErrorIs(t, err, ErrSelf)
	err = g.TransferAdmin(ben, other)
	uassert.ErrorIs(t, err, ErrNotAdmin)
	err = g.TransferAdmin(admin, address(""))
	uassert.ErrorIs(t, err, ErrAddress)
	err = g.TransferAdmin(admin, admin)
	uassert.ErrorIs(t, err, ErrSame)
	err = g.TransferAdmin(admin, g.Address())
	uassert.ErrorIs(t, err, ErrSelf)
	err = g.TransferAdmin(admin, hostAddr())
	uassert.ErrorIs(t, err, ErrSelf)

	view, err := g.View(110)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(7), view.ID)
	uassert.Equal(t, "ugnot", view.Asset.Denom)
	uassert.Equal(t, int64(10), view.Vested)
	_, err = g.View(-5)
	uassert.ErrorIs(t, err, ErrTime)

	long := strings.Repeat("k", MaxTokenLen+1)
	_, err = New(hostPkg, hostAddr(), admin, ben, Asset{Kind: AssetGRC20, Token: long}, terms, true, "")
	uassert.ErrorIs(t, err, ErrToken)
	_, err = New("gno.land/r/", chain.PackageAddress("gno.land/r/"), admin, ben, coinAsset(), terms, true, "")
	uassert.ErrorIs(t, err, ErrHost)
	_, err = New(strings.Repeat("gno.land/r/", 30), hostAddr(), admin, ben, coinAsset(), terms, true, "")
	uassert.ErrorIs(t, err, ErrHost)
}

func TestInvariantRejectsBrokenBooks(t *testing.T) {
	admin := testutils.TestAddress("ia")
	ben := testutils.TestAddress("ib")
	fresh := func(revocable bool) *Grant {
		return openGrant(t, admin, ben, linear(100, 0, 10, 50), revocable, 7)
	}

	g := fresh(true)
	uassert.NoError(t, g.Invariant(100))
	uassert.ErrorIs(t, g.Invariant(-1), ErrTime)

	// Negative or out-of-range counters.
	g = fresh(true)
	g.claimed = -1
	uassert.ErrorIs(t, g.Invariant(100), ErrMath)

	g = fresh(true)
	g.lastClaimAt = -1
	uassert.ErrorIs(t, g.Invariant(100), ErrMath)

	// Open grant with non-zero revoke fields or unsupported claimed amount.
	g = fresh(true)
	g.returned = 1
	uassert.ErrorIs(t, g.Invariant(100), ErrMath)

	g = fresh(true)
	g.frozen = 1
	uassert.ErrorIs(t, g.Invariant(100), ErrMath)

	g = fresh(true)
	g.revokedAt = 100
	uassert.ErrorIs(t, g.Invariant(100), ErrMath)

	g = fresh(true)
	g.lastClaimAt = 105
	uassert.ErrorIs(t, g.Invariant(110), ErrMath)

	g = fresh(true)
	g.claimed = 1
	uassert.ErrorIs(t, g.Invariant(110), ErrMath)

	g = fresh(true)
	g.claimed = 30
	g.lastClaimAt = 105 // VestedAmount at 105 is 25 < 30.
	uassert.ErrorIs(t, g.Invariant(110), ErrMath)

	// Revoked irrevocable grant.
	g = fresh(false)
	g.revoked = true
	g.revokedAt = 105
	g.frozen = 25
	g.returned = 25
	uassert.ErrorIs(t, g.Invariant(105), ErrIrrevocable)

	// Revoked grant with broken frozen / returned / claimed relations.
	g = fresh(true)
	_, err := g.Revoke(admin, 105) // frozen=25, returned=25
	uassert.NoError(t, err)
	uassert.NoError(t, g.Invariant(105))

	g.returned = 24
	uassert.ErrorIs(t, g.Invariant(105), ErrMath)
	g.returned = 25

	g.frozen = 30 // above VestedAmount(105) == 25
	g.returned = 20
	uassert.ErrorIs(t, g.Invariant(105), ErrMath)

	g.frozen = 20 // below VestedAmount(105) == 25
	g.returned = 30
	uassert.ErrorIs(t, g.Invariant(105), ErrMath)

	g.frozen = 25
	g.returned = 25
	g.claimed = 26 // above frozen
	uassert.ErrorIs(t, g.Invariant(105), ErrMath)

	// Corrupted address, roles, or label.
	g = fresh(true)
	g.addr = admin
	uassert.ErrorIs(t, g.Invariant(100), ErrMath)

	g = fresh(true)
	g.admin = hostAddr()
	uassert.ErrorIs(t, g.Invariant(100), ErrSelf)

	g = fresh(true)
	g.beneficiary = g.addr
	uassert.ErrorIs(t, g.Invariant(100), ErrSelf)

	g = fresh(true)
	g.label = "bad\nlabel"
	uassert.ErrorIs(t, g.Invariant(100), ErrLabel)

	unbound, err := New(hostPkg, hostAddr(), admin, ben, coinAsset(), linear(100, 0, 10, 50), false, "")
	uassert.NoError(t, err)
	uassert.ErrorIs(t, unbound.Invariant(100), ErrUnbound)
}

func TestNilGrant(t *testing.T) {
	var g *Grant
	uassert.Equal(t, int64(0), g.ID())
	uassert.Equal(t, address(""), g.Admin())
	uassert.Equal(t, address(""), g.Beneficiary())
	uassert.Equal(t, "", g.Label())
	uassert.True(t, g.Asset() == Asset{})
	uassert.True(t, g.Schedule() == Terms{})
	uassert.False(t, g.Revocable())
	uassert.False(t, g.Revoked())
	uassert.Equal(t, int64(0), g.RevokedAt())
	uassert.Equal(t, int64(0), g.Claimed())
	uassert.Equal(t, int64(0), g.Returned())
	uassert.Equal(t, "", g.Subpath())
	uassert.Equal(t, address(""), g.Address())
	uassert.Equal(t, "", g.HostPkg())
	_, err := g.Outstanding()
	uassert.ErrorIs(t, err, ErrUnbound)
	_, err = g.Reclaimable(1)
	uassert.ErrorIs(t, err, ErrUnbound)
}

func TestClockBackwardsDoesNotUnclaim(t *testing.T) {
	admin := testutils.TestAddress("admin3")
	ben := testutils.TestAddress("ben3")
	next := testutils.TestAddress("next3")
	g := openGrant(t, admin, ben, linear(0, 0, 100, 100), true, 9)
	got, err := g.Claim(ben, 40)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(40), got)
	uassert.NoError(t, g.Invariant(40))
	// Clock steps backwards: vested stays at claimed, Invariant holds,
	// and role updates and revoke do not fail.
	v, err := g.Vested(10)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(40), v)
	rel, err := g.Releasable(10)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(0), rel)
	uassert.NoError(t, g.Invariant(10))
	_, err = g.Claim(ben, 10)
	uassert.ErrorIs(t, err, ErrNothing)
	uassert.NoError(t, g.SetBeneficiary(ben, next))
	uassert.NoError(t, g.Invariant(10))
	uassert.NoError(t, g.TransferAdmin(admin, ben))
	uassert.NoError(t, g.Invariant(10))
	v, err = g.Vested(40)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(40), v)
	v, err = g.Vested(100)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(100), v)
	// Revoking while the clock is stepped back freezes at claimed (40),
	// not the lower raw curve (10), and returns 60.
	back, err := g.Revoke(ben, 10)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(60), back)
	uassert.NoError(t, g.Invariant(10))
	uassert.NoError(t, g.Invariant(100))
	_, err = g.Claim(next, 100)
	uassert.ErrorIs(t, err, ErrNothing)
}

func TestStepRevokeBeforeAndAfter(t *testing.T) {
	admin := testutils.TestAddress("admin4")
	ben := testutils.TestAddress("ben4")
	terms := Terms{Start: 0, Cliff: 0, Duration: 10, Period: 5, Total: 10, Curve: CurveStep}
	g := openGrant(t, admin, ben, terms, true, 11)
	got, err := g.Revoke(admin, 4)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(10), got)
	g = openGrant(t, admin, ben, terms, true, 12)
	got, err = g.Revoke(admin, 5)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(5), got)
	claimed, err := g.Claim(ben, 100)
	uassert.NoError(t, err)
	uassert.Equal(t, int64(5), claimed)
}
