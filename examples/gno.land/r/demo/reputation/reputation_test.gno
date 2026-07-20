package reputation

import (
	"testing"

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

func TestAddPointsDirectUserCallPanics(cur realm, t *testing.T) {
	addr := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	testing.SetRealm(testing.NewUserRealm(addr))

	uassert.AbortsWithMessage(t, cur,
		"reputation: AddPoints can only be called by another realm, not directly by a user",
		func() {
			AddPoints(cross(cur), addr, "test_category", 100)
		})
}

func TestAddPointsZeroOrNegativePanics(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/somecaller"))

	target := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	uassert.AbortsWithMessage(t, cur,
		"reputation: points must be positive",
		func() {
			AddPoints(cross(cur), target, "test_category", 0)
		})
}

func TestGetScoreUnknownReturnsZero(t *testing.T) {
	target := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	score := GetScore(target, "nonexistent_category", "gno.land/r/demo/somerandomrealm")
	uassert.Equal(t, int64(0), score)
}
