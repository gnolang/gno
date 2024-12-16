package fomo3d

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
)

// Test helpers
func shouldEqual(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()
	if got != expected {
		t.Errorf("expected %v(%T), got %v(%T)", expected, expected, got, got)
	}
}

func shouldPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	f()
}

func shouldNoPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	f()
}

// Setup helper
func setupTestGame(t *testing.T) {
	// Reset game state
	gameStartBlock = int64(0)
	gameEndBlock = int64(0)
	lastKeyBlock = int64(0)
	currentRound = int64(1)
	jackpot = int64(0)
	nextPot = int64(0)
	keyPrice = MIN_KEY_PRICE
	lastBuyer = ""
	players = avl.NewTree()
	totalKeys = int64(0)
	ended = true // So we can start a new game
}

// Test full game flow
func TestFullGameFlow(t *testing.T) {
	setupTestGame(t)

	player1 := testutils.TestAddress("player1")
	player2 := testutils.TestAddress("player2")
	player3 := testutils.TestAddress("player3")

	// Test initial state
	shouldEqual(t, currentRound, int64(1))
	shouldEqual(t, keyPrice, MIN_KEY_PRICE)
	shouldEqual(t, ended, true)

	// Start game
	shouldNoPanic(t, StartGame)
	shouldEqual(t, ended, false)
	shouldEqual(t, gameStartBlock, std.GetHeight())

	// Test buying keys
	{
		// Test insufficient payment
		std.TestSetOrigCaller(player1)
		std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE - 1}}, nil)
		shouldPanic(t, BuyKeys)

		// Test successful key purchase
		payment := MIN_KEY_PRICE * 3 // Buy 3 keys
		std.TestSetOrigSend(std.Coins{{"ugnot", payment}}, nil)
		shouldNoPanic(t, BuyKeys)

		// Verify player state
		keys, dividends := GetPlayerInfo(player1)
		shouldEqual(t, keys, int64(3))
		shouldEqual(t, dividends, int64(0))
		shouldEqual(t, lastBuyer, player1)

		// Verify game state
		start, end, last, buyer, pot, price, keys, isEnded := GetGameState()
		shouldEqual(t, buyer, player1)
		shouldEqual(t, keys, int64(3))
		shouldEqual(t, isEnded, false)
		shouldEqual(t, pot, payment*JACKPOT_PERCENT/100)
	}

	// Test dividend distribution and claiming
	{
		// Player 2 buys keys
		std.TestSetOrigCaller(player2)
		payment := MIN_KEY_PRICE * 2 // Buy 2 keys
		std.TestSetOrigSend(std.Coins{{"ugnot", payment}}, nil)
		shouldNoPanic(t, BuyKeys)

		// Check player1 received dividends
		keys1, dividends1 := GetPlayerInfo(player1)
		shouldEqual(t, keys1, int64(3))
		expectedDividends := payment * DIVIDENDS_PERCENT / 100 * 3 / totalKeys
		shouldEqual(t, dividends1, expectedDividends)

		// Test claiming dividends
		{
			// Player1 claims dividends
			std.TestSetOrigCaller(player1)
			initialBalance := std.GetBanker(std.BankerTypeRealmSend).GetCoins(player1)
			shouldNoPanic(t, ClaimDividends)

			// Verify dividends were claimed
			_, dividendsAfter := GetPlayerInfo(player1)
			shouldEqual(t, dividendsAfter, int64(0))

			finalBalance := std.GetBanker(std.BankerTypeRealmSend).GetCoins(player1)
			shouldEqual(t, finalBalance.AmountOf("ugnot"), initialBalance.AmountOf("ugnot")+expectedDividends)
		}
	}

	// Test game ending
	{
		// Try ending too early
		shouldPanic(t, EndGame)

		// Skip to end of game
		std.TestSkipHeights(MAX_TIME_EXTEND + 1)

		// End game successfully
		shouldNoPanic(t, EndGame)
		shouldEqual(t, ended, true)
		shouldEqual(t, currentRound, int64(2))

		// Verify winner received jackpot
		finalBalance := std.GetBanker(std.BankerTypeRealmSend).GetCoins(lastBuyer)
		shouldEqual(t, finalBalance.AmountOf("ugnot"), jackpot)
	}

	// Test new round
	{
		// Start new round
		shouldNoPanic(t, StartGame)
		shouldEqual(t, ended, false)
		shouldEqual(t, currentRound, int64(2))
		shouldEqual(t, jackpot, int64(1000000)) // Previous round's nextPot was 20% of 5000000 (total payments)
		shouldEqual(t, nextPot, int64(0))
	}
}

// Test individual components
func TestStartGame(t *testing.T) {
	setupTestGame(t)

	// Test starting first game
	shouldNoPanic(t, StartGame)
	shouldEqual(t, ended, false)
	shouldEqual(t, gameStartBlock, std.GetHeight())

	// Test cannot start while game in progress
	shouldPanic(t, StartGame)
}

func TestBuyKeys(t *testing.T) {
	setupTestGame(t)
	StartGame()

	player := testutils.TestAddress("player")
	std.TestSetOrigCaller(player)

	// Test invalid coin denomination
	std.TestSetOrigSend(std.Coins{{"invalid", MIN_KEY_PRICE}}, nil)
	shouldPanic(t, BuyKeys)

	// Test multiple coin types
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE}, {"other", 100}}, nil)
	shouldPanic(t, BuyKeys)

	// Test insufficient payment
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE - 1}}, nil)
	shouldPanic(t, BuyKeys)

	// Test successful purchase
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE * 2}}, nil)
	shouldNoPanic(t, BuyKeys)
}

func TestClaimDividends(t *testing.T) {
	setupTestGame(t)
	StartGame()

	player := testutils.TestAddress("player")
	std.TestSetOrigCaller(player)

	// Test claiming with no dividends
	shouldPanic(t, ClaimDividends)

	// Setup player with dividends
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE}}, nil)
	BuyKeys()

	// Have another player buy to generate dividends
	player2 := testutils.TestAddress("player2")
	std.TestSetOrigCaller(player2)
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE * 2}}, nil)
	BuyKeys()

	// Test successful claim
	std.TestSetOrigCaller(player)
	shouldNoPanic(t, ClaimDividends)
}

func TestEndGame(t *testing.T) {
	setupTestGame(t)
	StartGame()
	
	// Skip to end + 1 to ensure we're past the end block
	std.TestSkipHeights(MAX_TIME_EXTEND + 1)
	
	shouldPanic(t, EndGame) // Should panic because no keys purchased

	// Reset and start new game
	setupTestGame(t)
	StartGame()

	// Buy keys
	player := testutils.TestAddress("player")
	std.TestSetOrigCaller(player)
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE}}, nil)
	shouldNoPanic(t, BuyKeys)

	// Test ending too early
	shouldPanic(t, EndGame)

	// Skip to well past the end
	std.TestSkipHeights(MAX_TIME_EXTEND + TIME_EXTENSION + 1)
	
	shouldNoPanic(t, EndGame)
}
