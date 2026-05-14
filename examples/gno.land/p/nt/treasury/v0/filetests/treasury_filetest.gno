// PKGPATH: gno.land/r/treasury/main

package main

import (
	"chain"
	"chain/banker"
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/treasury/v0"
)

func checkBalanceAndHistory(t *treasury.Treasury, bankerIDs []string) {
	for _, bankerID := range bankerIDs {
		balances, err := t.Balances(bankerID)
		if err != nil {
			panic("failed to get banker balances: " + err.Error())
		}
		println("Banker", bankerID, "Balance:", balances[0].Amount)

		history, err := t.History(bankerID, 1, 10)
		if err != nil {
			panic("failed to get banker history: " + err.Error())
		}
		println("Banker", bankerID, "History count:", len(history))
	}
}

func main() {
	// Define addresses for the sender (owner) and destination.
	ownerAddr := chain.PackageAddress("gno.land/r/treasury/main")
	destAddr := chain.PackageAddress("gno.land/r/dest/main")

	// Try to create a Treasury instance with no bankers.
	_, err := treasury.New(nil)
	if err != treasury.ErrNoBankerProvided {
		panic("expected error when creating Treasury with no bankers")
	}

	// Define a token and the associated lister.
	const amount = int64(1000)
	token, ledger := grc20.NewToken("TestToken", "TEST", 0)
	ledger.Mint(ownerAddr, amount)

	grc20Lister := func() map[string]*grc20.Token {
		return map[string]*grc20.Token{
			"TEST": token,
		}
	}

	// Try to create a Treasury instance with a duplicate banker.
	var (
		banker_           = banker.NewBanker(banker.BankerTypeRealmSend)
		coinsBanker, _    = treasury.NewCoinsBanker(banker_)
		grc20Banker, _    = treasury.NewGRC20Banker(grc20Lister)
		grc20BankerDup, _ = treasury.NewGRC20Banker(grc20Lister)
	)

	_, err = treasury.New([]treasury.Banker{coinsBanker, grc20Banker, grc20BankerDup})
	if !strings.Contains(err.Error(), treasury.ErrDuplicateBanker.Error()) {
		panic("expected error when creating Treasury with duplicate banker")
	}

	// Create a Treasury instance with valid bankers.
	bankers := []treasury.Banker{coinsBanker, grc20Banker}
	t, err := treasury.New(bankers)
	if err != nil {
		panic("failed to create Treasury: " + err.Error())
	}

	// Test if the Treasury instance has the expected bankers.
	println("Treasury banker IDs:", t.ListBankerIDs())

	const unknownBankerID = "unknown-banker-id"

	if t.HasBanker(unknownBankerID) {
		panic("expected banker not to be found")
	}

	// Check if the addresses of the bankers matches the owner address.
	for _, banker_ := range bankers {
		addr, err := t.Address(banker_.ID())
		if err != nil {
			panic("failed to get banker address: " + err.Error())
		}
		println("Banker", banker_.ID(), "Address:", addr)
	}

	// Check if the balances and history of the bankers match the expected values.
	testing.IssueCoins(ownerAddr, chain.NewCoins(chain.NewCoin("ugnot", amount)))
	bankerIDs := []string{coinsBanker.ID(), grc20Banker.ID()}
	checkBalanceAndHistory(t, bankerIDs)

	// Send 3 valid payments using the CoinsBanker.
	validCoinsPayment := treasury.NewCoinsPayment(
		chain.NewCoins(chain.NewCoin("ugnot", 100)),
		destAddr,
	)
	for i := 0; i < 3; i++ {
		err = t.Send(validCoinsPayment)
		if err != nil {
			panic("failed to send valid Coins payment: " + err.Error())
		}
	}

	// Send 3 valid payments using the GRC20Banker.
	validGRC20Payment := treasury.NewGRC20Payment(
		token.GetSymbol(),
		100,
		destAddr,
	)
	for i := 0; i < 3; i++ {
		err = t.Send(validGRC20Payment)
		if err != nil {
			panic("failed to send valid GRC20 payment: " + err.Error())
		}
	}

	// Check if the balances and history of the bankers match the expected values.
	checkBalanceAndHistory(t, bankerIDs)
}

// Output:
// Treasury banker IDs: slice[("Coins" string),("GRC20" string)]
// Banker Coins Address: g1ynsdz5zaxhn9gnqtr6t40m5k4fueeutq7xy224
// Banker GRC20 Address: g1ynsdz5zaxhn9gnqtr6t40m5k4fueeutq7xy224
// Banker Coins Balance: 1000
// Banker Coins History count: 0
// Banker GRC20 Balance: 1000
// Banker GRC20 History count: 0
// Banker Coins Balance: 700
// Banker Coins History count: 3
// Banker GRC20 Balance: 700
// Banker GRC20 History count: 3
