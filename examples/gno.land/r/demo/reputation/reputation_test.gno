package reputation

import (
	"testing"

	"chain/runtime"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestAddPointsFromRealmSucceeds(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/somecaller"))

	target := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	AddPoints(cross(cur), target, "test_category", 100)

	score := GetScore(target, "test_category", "gno.land/r/demo/somecaller")
	uassert.Equal(t, int64(100), score)

	total := GetTotalScore(target)
	uassert.Equal(t, int64(100), total)
}

func TestAddPointsAccumulates(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/anothercaller"))

	target := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	AddPoints(cross(cur), target, "another_category", 50)
	AddPoints(cross(cur), target, "another_category", 25)

	score := GetScore(target, "another_category", "gno.land/r/demo/anothercaller")
	uassert.Equal(t, int64(75), score)
}

func TestAddPointsIsolatesDifferentIssuersSameCategory(cur realm, t *testing.T) {
	target := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/issuera"))
	AddPoints(cross(cur), target, "shared_category", 40)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/issuerb"))
	AddPoints(cross(cur), target, "shared_category", 60)

	scoreA := GetScore(target, "shared_category", "gno.land/r/demo/issuera")
	scoreB := GetScore(target, "shared_category", "gno.land/r/demo/issuerb")
	uassert.Equal(t, int64(40), scoreA)
	uassert.Equal(t, int64(60), scoreB)
}

func TestAddPointsDirectUserCallPanics(cur realm, t *testing.T) {
	addr := testutils.TestAddress("direct-user-caller")
	testing.SetRealm(testing.NewUserRealm(addr))

	uassert.AbortsWithMessage(t, cur,
		"reputation: AddPoints can only be called by another realm, not directly by a user",
		func() {
			AddPoints(cross(cur), addr, "test_category", 100)
		})
}

// A `gnokey maketx run` script executes in an ephemeral realm at
// domain/e/<addr>/run, which has a non-empty pkgPath -- so IsUserCall()
// alone would let it through. IsUser() catches this case too.
// NewCodeRealm refuses ephemeral paths (isRealm requires a real
// deployed realm path), so this is built directly with
// runtime.NewRealm, which does no such check.
func TestAddPointsEphemeralRunRealmPanics(cur realm, t *testing.T) {
	addr := testutils.TestAddress("ephemeral-run-caller")
	ephemeralPath := "gno.land/e/" + addr.String() + "/run"
	testing.SetRealm(runtime.NewRealm(addr, ephemeralPath))

	uassert.AbortsWithMessage(t, cur,
		"reputation: AddPoints can only be called by another realm, not directly by a user",
		func() {
			AddPoints(cross(cur), addr, "test_category", 100)
		})
}

func TestAddPointsZeroOrNegativePanics(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/somecaller"))

	target := testutils.TestAddress("zero-points-target")
	uassert.AbortsWithMessage(t, cur,
		"reputation: points must be positive",
		func() {
			AddPoints(cross(cur), target, "test_category", 0)
		})
}

func TestAddPointsOverflowPanics(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/somecaller"))

	target := testutils.TestAddress("overflow-target")
	AddPoints(cross(cur), target, "overflow_category", maxInt64)

	uassert.AbortsWithMessage(t, cur,
		"reputation: score overflow",
		func() {
			AddPoints(cross(cur), target, "overflow_category", maxInt64)
		})
}

func TestGetScoreUnknownReturnsZero(t *testing.T) {
	target := testutils.TestAddress("unknown-score-target")
	score := GetScore(target, "nonexistent_category", "gno.land/r/demo/somerandomrealm")
	uassert.Equal(t, int64(0), score)
}
