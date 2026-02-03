package test

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"sort"
	"strings"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/nt/fqname"
	"gno.land/p/nt/testutils"
	trs_pkg "gno.land/p/nt/treasury"
	"gno.land/p/nt/uassert"

	"gno.land/r/demo/defi/grc20reg"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/treasury"
)

var (
	user1Addr       = testutils.TestAddress("g1user1")
	user2Addr       = testutils.TestAddress("g1user2")
	treasuryAddr    = chain.PackageAddress("gno.land/r/gov/dao/v3/treasury")
	allowedRealm    = testing.NewCodeRealm("gno.land/r/test/allowed")
	notAllowedRealm = testing.NewCodeRealm("gno.land/r/test/notallowed")
	mintAmount      = int64(1000)
)

// Define a dummy trs_pkg.Payment type for testing purposes.
type dummyPayment struct {
	bankerID string
	str      string
}

var _ trs_pkg.Payment = (*dummyPayment)(nil)

func (dp *dummyPayment) BankerID() string { return dp.bankerID }
func (dp *dummyPayment) String() string   { return dp.str }

func init() {
	// Register allowed Realm path.
	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO:         impl.NewGovDAO(),
		AllowedDAOs: []string{allowedRealm.PkgPath()},
	})
}

func ugnotCoins(t *testing.T, amount int64) chain.Coins {
	t.Helper()

	// Create a new coin with the ugnot denomination.
	return chain.NewCoins(chain.NewCoin("ugnot", amount))
}

func ugnotBalance(t *testing.T, addr address) int64 {
	t.Helper()

	// Get the balance of ugnot coins for the given address.
	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	coins := banker_.GetCoins(addr)

	return coins.AmountOf("ugnot")
}

// Define a keyedToken type to hold the token and its key.
type keyedToken struct {
	key   string
	token *grc20.Token
}

func registerGRC20Tokens(t *testing.T, tokenNames []string, toMint address) []keyedToken {
	t.Helper()

	var (
		keyedTokens = make([]keyedToken, 0, len(tokenNames))
		keys        = make([]string, 0, len(tokenNames))
	)

	for _, name := range tokenNames {
		// Create the token.
		symbol := strings.ToUpper(name)
		token, ledger := grc20.NewToken(name, symbol, 0)

		// Register the token.
		grc20reg.Register(cross, token, symbol)

		// Mint tokens to the specified address.
		ledger.Mint(toMint, mintAmount)

		// Add the token and key to the lists.
		key := fqname.Construct(runtime.CurrentRealm().PkgPath(), symbol)
		keyedTokens = append(keyedTokens, keyedToken{key: key, token: token})
		keys = append(keys, key)
	}

	// Set the token keys in the treasury.
	treasury.SetTokenKeys(cross, keys)

	return keyedTokens
}

func TestAllowedDAOs(t *testing.T) {
	// Set the current Realm to the not allowed one.
	testing.SetRealm(notAllowedRealm)

	// Define a dummy payment to test sending.
	dummyP := &dummyPayment{bankerID: "Dummy"}

	// Try to send, it should abort because the Realm is not allowed.
	uassert.AbortsWithMessage(
		t,
		"this Realm is not allowed to send payment: "+notAllowedRealm.PkgPath(),
		func() { treasury.Send(cross, dummyP) },
	)

	// Set the current Realm to the allowed one.
	testing.SetRealm(allowedRealm)

	// Try to send, it should not abort because the Realm is allowed,
	// but because the dummy banker ID is not registered.
	uassert.AbortsWithMessage(
		t,
		"banker not found: "+dummyP.BankerID(),
		func() { treasury.Send(cross, dummyP) },
	)
}

func TestRegisteredBankers(t *testing.T) {
	// Set the current Realm to the allowed one.
	testing.SetRealm(allowedRealm)

	// Define the expected banker IDs.
	expectedBankerIDs := []string{
		trs_pkg.CoinsBanker{}.ID(),
		trs_pkg.GRC20Banker{}.ID(),
	}

	// Get the registered bankers from the treasury and compare their lengths.
	registeredBankerIDs := treasury.ListBankerIDs()
	uassert.Equal(t, len(registeredBankerIDs), len(expectedBankerIDs))

	// Sort both slices then compare them.
	sort.StringSlice(expectedBankerIDs).Sort()
	sort.StringSlice(registeredBankerIDs).Sort()

	for i := range expectedBankerIDs {
		uassert.Equal(t, expectedBankerIDs[i], registeredBankerIDs[i])
	}

	// Test HasBanker method.
	for _, bankerID := range expectedBankerIDs {
		uassert.True(t, treasury.HasBanker(bankerID))
	}
	uassert.False(t, treasury.HasBanker("UnknownBankerID"))

	// Test Address method.
	for _, bankerID := range expectedBankerIDs {
		// The two bankers used for now should have the treasury Realm address.
		uassert.Equal(t, treasury.Address(bankerID), treasuryAddr.String())
	}
}

func TestSendGRC20Payment(t *testing.T) {
	// Set the current Realm to the allowed one.
	testing.SetRealm(allowedRealm)

	// Try to send a GRC20 payment with a not registered token, it should abort.
	uassert.AbortsWithMessage(
		t,
		"failed to send payment: GRC20 token not found: UNKNOW",
		func() {
			treasury.Send(cross, trs_pkg.NewGRC20Payment("UNKNOW", 100, user1Addr))
		},
	)

	// Create 3 GRC20 tokens and register them.
	keyedTokens := registerGRC20Tokens(
		t,
		[]string{"TestToken0", "TestToken1", "TestToken2"},
		treasuryAddr,
	)

	const txAmount = 42

	// For each token-user pair.
	for i, userAddr := range []address{user1Addr, user2Addr} {
		for _, keyed := range keyedTokens {
			// Check that the treasury has the expected balance before sending.
			uassert.Equal(t, keyed.token.BalanceOf(treasuryAddr), mintAmount-int64(txAmount*i))

			// Check that the user has no balance before sending.
			uassert.Equal(t, keyed.token.BalanceOf(userAddr), int64(0))

			// Try to send a GRC20 payment with a registered token, it should not abort.
			uassert.NotAborts(t, func() {
				treasury.Send(
					cross,
					trs_pkg.NewGRC20Payment(
						keyed.key,
						txAmount,
						userAddr,
					),
				)
			})

			// Check that the user has the expected balance after sending.
			uassert.Equal(t, keyed.token.BalanceOf(userAddr), int64(txAmount))

			// Check that the treasury has the expected balance after sending.
			uassert.Equal(t, keyed.token.BalanceOf(treasuryAddr), mintAmount-int64(txAmount*(i+1)))
		}
	}

	// Get the GRC20Banker ID.
	grc20BankerID := trs_pkg.GRC20Banker{}.ID()

	// Test Balances method for the GRC20Banker.
	balances := treasury.Balances(grc20BankerID)
	uassert.Equal(t, len(balances), len(keyedTokens))

	compared := 0
	for _, balance := range balances {
		for _, keyed := range keyedTokens {
			if balance.Denom == keyed.key {
				uassert.Equal(t, balance.Amount, keyed.token.BalanceOf(treasuryAddr))
				compared++
			}
		}
	}
	uassert.Equal(t, compared, len(keyedTokens))

	// Check the history of the GRC20Banker.
	history := treasury.History(grc20BankerID, 1, 10)
	uassert.Equal(t, len(history), 6)

	// Try to send a dummy payment with the GRC20 banker ID, it should abort.
	uassert.AbortsWithMessage(
		t,
		"failed to send payment: invalid payment type",
		func() {
			treasury.Send(cross, &dummyPayment{bankerID: grc20BankerID})
		},
	)

	// Try to send a GRC20 payment without enough balance, it should abort.
	uassert.AbortsWithMessage(
		t,
		"failed to send payment: insufficient balance",
		func() {
			treasury.Send(
				cross,
				trs_pkg.NewGRC20Payment(
					keyedTokens[0].key,
					mintAmount*42, // Try to send more than the treasury has.
					user1Addr,
				),
			)
		},
	)

	// Check the history of the GRC20Banker.
	history = treasury.History(grc20BankerID, 1, 10)
	uassert.Equal(t, len(history), 6)
}

func TestSendCoinPayment(t *testing.T) {
	// Set the current Realm to the allowed one.
	testing.SetRealm(allowedRealm)

	// Issue initial ugnot coins to the treasury address.
	testing.IssueCoins(treasuryAddr, ugnotCoins(t, mintAmount))

	// Get the CoinsBanker ID.
	bankerID := trs_pkg.CoinsBanker{}.ID()

	// Define helper function to check balances and history.
	var (
		expectedTreasuryBalance = mintAmount
		expectedUser1Balance    = int64(0)
		expectedUser2Balance    = int64(0)
		expectedHistoryLen      = 0
		checkHistoryAndBalances = func() {
			t.Helper()

			uassert.Equal(t, ugnotBalance(t, treasuryAddr), expectedTreasuryBalance)
			uassert.Equal(t, ugnotBalance(t, user1Addr), expectedUser1Balance)
			uassert.Equal(t, ugnotBalance(t, user2Addr), expectedUser2Balance)

			// Check treasury.Balances returned value.
			balances := treasury.Balances(bankerID)
			uassert.Equal(t, len(balances), 1)
			uassert.Equal(t, balances[0].Denom, "ugnot")
			uassert.Equal(t, balances[0].Amount, expectedTreasuryBalance)

			// Check treasury.History returned value.
			history := treasury.History(bankerID, 1, expectedHistoryLen+1)
			uassert.Equal(t, len(history), expectedHistoryLen)
		}
	)

	// Check initial balances and history.
	checkHistoryAndBalances()

	const txAmount = int64(42)

	// Treasury send coins.
	for i := int64(0); i < 3; i++ {
		// Send ugnot coins to user1 and user2.
		uassert.NotAborts(t, func() {
			treasury.Send(
				cross,
				trs_pkg.NewCoinsPayment(ugnotCoins(t, txAmount), user1Addr),
			)
			treasury.Send(
				cross,
				trs_pkg.NewCoinsPayment(ugnotCoins(t, txAmount), user2Addr),
			)
		})

		// Update expected balances and history length.
		expectedTreasuryBalance = mintAmount - txAmount*2*(i+1)
		expectedUser1Balance = txAmount * (i + 1)
		expectedUser2Balance = expectedUser1Balance
		expectedHistoryLen = int(2 * (i + 1))

		// Check balances and history after sending.
		checkHistoryAndBalances()
	}
}
