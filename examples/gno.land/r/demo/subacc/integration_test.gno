package subacc_test

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	"gno.land/r/demo/authbanker"
	"gno.land/r/demo/subacc"
)

func TestAuthIntegration(t *testing.T) {
	authbankerAddr := std.DerivePkgAddr("gno.land/r/demo/authbanker")

	alice := testutils.TestAddress("alice")
	aliceAccountID := subacc.EntityID(alice, "savings")

	bob := testutils.TestAddress("bob")
	bobAccountID := subacc.EntityID(bob, "savings")

	std.TestSetOrigSend(std.Coins{{"ugnot", 42}}, nil)
	authbanker.FundVault(aliceAccountID)
	urequire.Equal(t, int64(42), authbanker.GetCoins(aliceAccountID))
	// XXX: uncoment next line when tx send in tests is fixed
	// urequire.Equal(t, int64(42), std.GetBanker(std.BankerTypeReadonly).GetCoins(authbankerAddr)[0].Amount)

	std.TestSetRealm(std.NewUserRealm(alice))
	aliceToken := subacc.AuthToken("savings")
	authbanker.SendCoins(aliceToken, bobAccountID, 14)
	urequire.Equal(t, int64(28), authbanker.GetCoins(aliceAccountID))
	urequire.Equal(t, int64(14), authbanker.GetCoins(bobAccountID))

	std.TestSetRealm(std.NewUserRealm(bob))
	bobToken := subacc.AuthToken("savings")
	urequire.PanicsWithMessage(t, "not enough in account", func() { authbanker.SendCoins(bobToken, alice.String(), 15) })
	std.TestIssueCoins(authbankerAddr, std.Coins{{"ugnot", 7}}) // XXX: we need this line because in tests, tx send does not work, remove when tx send in test is fixed
	authbanker.SendCoins(bobToken, alice.String(), 7)
	urequire.Equal(t, int64(28), authbanker.GetCoins(aliceAccountID))
	urequire.Equal(t, int64(7), authbanker.GetCoins(bobAccountID))
	urequire.Equal(t, int64(7), std.GetBanker(std.BankerTypeReadonly).GetCoins(alice)[0].Amount)
	urequire.Equal(t, 0, len(std.GetBanker(std.BankerTypeReadonly).GetCoins(authbankerAddr))) // XXX: replace with next line when tx send in tests is fixed
	// urequire.Equal(t, int64(35), std.GetBanker(std.BankerTypeReadonly).GetCoins(authbankerAddr)[0].Amount)
}
