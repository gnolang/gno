package predictionmarket

import (
	"testing"

	"chain"

	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var realmAddr = chain.PackageAddress("gno.land/r/demo/predictionmarket")

func TestCreateMarketAndPlaceBets(cur realm, t *testing.T) {
	admin := testutils.TestAddress("admin")
	owner = ownable.NewWithAddress(admin)
	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	id := CreateMarket(cross(cur), "Test Market", true)

	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))
	testing.SetOriginCaller(alice)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin(betDenom, 1000)))
	testing.IssueCoins(realmAddr, chain.NewCoins(chain.NewCoin(betDenom, 1000)))
	PlaceBet(cross(cur), id, OutcomeHome)

	bob := testutils.TestAddress("bob")
	testing.SetRealm(testing.NewUserRealm(bob))
	testing.SetOriginCaller(bob)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin(betDenom, 500)))
	testing.IssueCoins(realmAddr, chain.NewCoins(chain.NewCoin(betDenom, 1500)))
	PlaceBet(cross(cur), id, OutcomeAway)

	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	ResolveMarket(cross(cur), id, OutcomeHome)

	testing.SetRealm(testing.NewUserRealm(alice))
	testing.SetOriginCaller(alice)
	ClaimWinnings(cross(cur), id)
}

func TestAdminCannotPlaceBet(cur realm, t *testing.T) {
	admin := testutils.TestAddress("admin-nobet")
	owner = ownable.NewWithAddress(admin)
	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	id := CreateMarket(cross(cur), "Admin Cannot Bet", true)

	testing.SetOriginSend(chain.NewCoins(chain.NewCoin(betDenom, 1000)))
	testing.IssueCoins(realmAddr, chain.NewCoins(chain.NewCoin(betDenom, 1000)))

	uassert.AbortsWithMessage(t, cur, "admin cannot place bets", func() {
		PlaceBet(cross(cur), id, OutcomeHome)
	})
}

func TestNoWinnersTriggersRefund(cur realm, t *testing.T) {
	admin := testutils.TestAddress("admin-refund")
	owner = ownable.NewWithAddress(admin)
	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	id := CreateMarket(cross(cur), "No Winners Test", true)

	alice := testutils.TestAddress("alice-refund")
	testing.SetRealm(testing.NewUserRealm(alice))
	testing.SetOriginCaller(alice)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin(betDenom, 700)))
	testing.IssueCoins(realmAddr, chain.NewCoins(chain.NewCoin(betDenom, 700)))
	PlaceBet(cross(cur), id, OutcomeHome)

	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	ResolveMarket(cross(cur), id, OutcomeAway)

	testing.SetRealm(testing.NewUserRealm(alice))
	testing.SetOriginCaller(alice)
	uassert.NotPanics(t, cur, func() {
		ClaimWinnings(cross(cur), id)
	})
}

func TestResolveMarketTwiceFails(cur realm, t *testing.T) {
	admin := testutils.TestAddress("admin-double")
	owner = ownable.NewWithAddress(admin)
	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	id := CreateMarket(cross(cur), "Double Resolve Test", true)

	ResolveMarket(cross(cur), id, OutcomeAway)

	uassert.AbortsWithMessage(t, cur, "market already resolved", func() {
		ResolveMarket(cross(cur), id, OutcomeDraw)
	})
}

func TestClaimWinningsBeforeResolveFails(cur realm, t *testing.T) {
	admin := testutils.TestAddress("admin-early")
	owner = ownable.NewWithAddress(admin)
	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	id := CreateMarket(cross(cur), "Early Claim Test", true)

	uassert.AbortsWithMessage(t, cur, "market not resolved yet", func() {
		ClaimWinnings(cross(cur), id)
	})
}

func TestClaimWinningsOverflowingPoolPanics(cur realm, t *testing.T) {
	admin := testutils.TestAddress("admin-overflow")
	owner = ownable.NewWithAddress(admin)
	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	id := CreateMarket(cross(cur), "Overflow Test", true)

	alice := testutils.TestAddress("alice-overflow")
	testing.SetRealm(testing.NewUserRealm(alice))
	testing.SetOriginCaller(alice)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin(betDenom, 5_000_000_000)))
	testing.IssueCoins(realmAddr, chain.NewCoins(chain.NewCoin(betDenom, 5_000_000_000)))
	PlaceBet(cross(cur), id, OutcomeHome)

	bob := testutils.TestAddress("bob-overflow")
	testing.SetRealm(testing.NewUserRealm(bob))
	testing.SetOriginCaller(bob)
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin(betDenom, 5_000_000_000)))
	testing.IssueCoins(realmAddr, chain.NewCoins(chain.NewCoin(betDenom, 10_000_000_000)))
	PlaceBet(cross(cur), id, OutcomeAway)

	testing.SetRealm(testing.NewUserRealm(admin))
	testing.SetOriginCaller(admin)
	ResolveMarket(cross(cur), id, OutcomeHome)

	testing.SetRealm(testing.NewUserRealm(alice))
	testing.SetOriginCaller(alice)
	uassert.AbortsWithMessage(t, cur, "multiplication overflow", func() {
		ClaimWinnings(cross(cur), id)
	})
}
