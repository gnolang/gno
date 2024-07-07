package users

import (
	"std"
	"testing"

	"gno.land/p/demo/entropy"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

/*
 	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice)
	std.TestIssueCoins(alice, std.Coins{{"ugnot", 199000000}})
	std.TestSetOrigSend(std.Coins{{"ugnot", 199000000}}, nil)
*/

func TestRegister_successWithPayment(t *testing.T) {
	std.TestSetOrigSend(std.Coins{std.NewCoin("ugnot", 200000000)}, nil)
	Register("", "gnouser", "myprofile")
}

func TestRegister_notEnoughPayment(t *testing.T) {
	std.TestSetOrigSend(std.Coins{std.NewCoin("ugnot", 199000000)}, nil)
	uassert.PanicsWithMessage(t, "payment must not be less than 200000000", func() {
		register("", "gnouser", "myprofile")
	})
}

func TestRegister_invalidDenom(t *testing.T) {
	std.TestSetOrigSend(std.Coins{std.NewCoin("dontcare", 1)}, nil)
	uassert.PanicsWithMessage(t, "incompatible coin denominations: dontcare, ugnot", func() {
		register("", "gnouser", "myprofile")
	})
}

func TestClaimDefaultName(t *testing.T) {
	ent = entropy.FromSeed(42) // fixed entropy
	profile := "hello world"
	name := ClaimDefaultName(profile)
	expected := "nifty_teal_engineer_51" // this can easily break if p/demo/wordlist is modified
	uassert.Equal(t, expected, name)
}
