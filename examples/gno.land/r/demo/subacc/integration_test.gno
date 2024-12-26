package subacc_test

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	"gno.land/r/demo/authbanker"
	"gno.land/r/demo/subacc"
)

func TestEntityID(t *testing.T) {
	alice := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // test1
	expected := "/g1u5hdzuqfln65xyx5dvz9ldl9e45pmew5exg302/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/savings"
	urequire.Equal(t, expected, subacc.EntityID(alice, "savings"))
}

func TestSubacc(t *testing.T) {
	alice := testutils.TestAddress("alice")
	aliceAccountID := subacc.EntityID(alice, "savings")

	bob := testutils.TestAddress("bob")
	bobAccountID := subacc.EntityID(bob, "savings")

	std.TestSetOrigSend(std.Coins{{"ugnot", 42}}, nil)
	authbanker.FundVault(aliceAccountID)
	urequire.Equal(t, int64(42), authbanker.GetCoins(aliceAccountID))

	std.TestSetRealm(std.NewUserRealm(alice))
	aliceToken := subacc.AuthToken("savings")
	authbanker.SendCoins(aliceToken, bobAccountID, 14)
	urequire.Equal(t, int64(28), authbanker.GetCoins(aliceAccountID))
	urequire.Equal(t, int64(14), authbanker.GetCoins(bobAccountID))

	std.TestSetRealm(std.NewUserRealm(bob))
	bobToken := subacc.AuthToken("savings")
	urequire.PanicsWithMessage(t, `not enough "ugnot" in account, wanted 15, got 14`, func() { authbanker.SendCoins(bobToken, alice.String(), 15) })
	/*
		authbanker.SendCoins(bobToken, alice.String(), 7)
		urequire.Equal(t, int64(28), authbanker.GetCoins(aliceAccountID))
		urequire.Equal(t, int64(7), authbanker.GetCoins(bobAccountID))
		urequire.Equal(t, int64(7), std.GetBanker(std.BankerTypeReadonly).GetCoins(alice)[0].Amount)
	*/
}
