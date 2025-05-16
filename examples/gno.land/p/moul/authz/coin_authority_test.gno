package authz_test

/*
import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/moul/authz"
)

func TestCoinAuthority_Basic(t *testing.T) {
	t.Skip("banker is hard to test in unit tests")
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	// Setup test coins and actions
	minCoins := std.Coins{
		{Denom: "ugnot", Amount: 1000},
		{Denom: "utoken", Amount: 500},
	}
	actions := []string{"propose", "vote"}
	auth := authz.NewCoinAuthority(actions, minCoins)

	// Create test addresses
	richUser := std.Address("g1rich")
	poorUser := std.Address("g1poor")

	// Setup mock banker
	testing.IssueCoins(richUser, std.Coins{{"ugnot", 2000}, {"utoken", 1000}})
	testing.IssueCoins(poorUser, std.Coins{{"ugnot", 500}, {"utoken", 200}})

	// Test rich user permissions
	uassert.True(t, auth.Can(richUser, "propose"), "rich user should be able to propose")
	uassert.True(t, auth.Can(richUser, "vote"), "rich user should be able to vote")
	uassert.False(t, auth.Can(richUser, "invalid_action"), "rich user should not be able to perform invalid action")

	// Test poor user permissions
	uassert.False(t, auth.Can(poorUser, "propose"), "poor user should not be able to propose")
	uassert.False(t, auth.Can(poorUser, "vote"), "poor user should not be able to vote")
	uassert.False(t, auth.Can(poorUser, "invalid_action"), "poor user should not be able to perform invalid action")
}

func TestCoinAuthority_SingleCoin(t *testing.T) {
	t.Skip("banker is hard to test in unit tests")
	testing.SetRealm(std.NewUserRealm("test1"))
	// Setup test with single coin
	minCoins := std.Coins{
		{Denom: "ugnot", Amount: 1000},
	}
	actions := []string{"action"}
	auth := authz.NewCoinAuthority(actions, minCoins)

	// Create test addresses
	user := std.Address("g1user")

	// Setup mock banker
	testing.IssueCoins(user, std.Coins{{"ugnot", 1000}})

	// Test exact balance
	uassert.True(t, auth.Can(user, "action"), "user should be able to perform action with exact balance")

	// Test insufficient balance
	testing.IssueCoins(user, std.Coins{{"ugnot", 999}})
	uassert.False(t, auth.Can(user, "action"), "user should not be able to perform action with insufficient balance")
}

func TestCoinAuthority_MultipleCoins(t *testing.T) {
	t.Skip("banker is hard to test in unit tests")
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test"))

	// Setup test with multiple coins
	minCoins := std.Coins{
		{Denom: "ugnot", Amount: 1000},
		{Denom: "utoken", Amount: 500},
		{Denom: "uother", Amount: 100},
	}
	actions := []string{"action"}
	auth := authz.NewCoinAuthority(actions, minCoins)

	// Create test address
	user := std.Address("g1user")

	// Test all coins sufficient
	testing.IssueCoins(user, std.Coins{
		{"ugnot", 1000},
		{"utoken", 500},
		{"uother", 100},
	})
	uassert.True(t, auth.Can(user, "action"), "user should be able to perform action with all sufficient balances")

	// Test one coin insufficient
	testing.IssueCoins(user, std.Coins{{"utoken", 499}})
	uassert.False(t, auth.Can(user, "action"), "user should not be able to perform action with one insufficient balance")

	// Test missing coin
	testing.IssueCoins(user, std.Coins{{"ugnot", 1000}})
	testing.IssueCoins(user, std.Coins{{"utoken", 500}})
	// uother not set
	uassert.False(t, auth.Can(user, "action"), "user should not be able to perform action with missing coin")
}

func TestCoinAuthority_Empty(t *testing.T) {
	t.Skip("banker is hard to test in unit tests")

	// Test with empty coins
	auth := authz.NewCoinAuthority([]string{"action"}, std.Coins{})
	user := std.Address("g1user")
	uassert.True(t, auth.Can(user, "action"), "user should be able to perform action with no coin requirements")

	// Test with empty actions
	auth = authz.NewCoinAuthority([]string{}, std.Coins{{Denom: "ugnot", Amount: 1000}})
	uassert.False(t, auth.Can(user, "any_action"), "no actions should be allowed with empty whitelist")
}

func TestCoinAuthority_ZeroAmount(t *testing.T) {
	t.Skip("banker is hard to test in unit tests")

	// Test with zero amount coins
	minCoins := std.Coins{
		{Denom: "ugnot", Amount: 0},
	}
	actions := []string{"action"}
	auth := authz.NewCoinAuthority(actions, minCoins)

	// Create test address
	user := std.Address("g1user")

	// Setup mock banker
	testing.IssueCoins(user, std.Coins{{"ugnot", 0}})

	// Test zero balance
	uassert.True(t, auth.Can(user, "action"), "user should be able to perform action with zero balance requirement")
}

*/
