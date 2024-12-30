package fomo3d

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
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

// Reset game state
func setupTestGame(t *testing.T) {
	gameState = GameState{
		StartBlock:   0,
		EndBlock:     0,
		LastKeyBlock: 0,
		LastBuyer:    "",
		Jackpot:      0,
		KeyPrice:     MIN_KEY_PRICE,
		TotalKeys:    0,
		Ended:        true,
		CurrentRound: 1,
		NextPot:      0,
		OwnerFee:     0,
	}
	players = avl.NewTree()
	Ownable = ownable.New()
}

// Test ownership functionality
func TestOwnership(t *testing.T) {
	owner := testutils.TestAddress("owner")
	nonOwner := testutils.TestAddress("nonOwner")

	// Set up initial owner
	std.TestSetOrigCaller(owner)
	std.TestSetOrigPkgAddr(owner)
	setupTestGame(t)

	// Transfer ownership to nonOwner first to test ownership functions
	std.TestSetOrigCaller(owner)
	shouldNoPanic(t, func() {
		TransferOwnership(nonOwner)
	})

	// Test fee accumulation
	StartGame()
	payment := MIN_KEY_PRICE * 10
	std.TestSetOrigCaller(owner)
	std.TestSetOrigSend(std.Coins{{"ugnot", payment}}, nil)
	std.TestIssueCoins(owner, std.Coins{{"ugnot", payment}})
	std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", payment}})
	BuyKeys()

	// Verify fee accumulation
	_, fees := GetOwnerInfo()
	expectedFees := payment * OWNER_FEE_PERCENT / 100
	shouldEqual(t, fees, expectedFees)

	// Test unauthorized fee claim (using old owner)
	std.TestSetOrigCaller(owner)
	shouldPanic(t, ClaimOwnerFee)

	// Test authorized fee claim (using new owner)
	std.TestSetOrigCaller(nonOwner)
	initialBalance := std.GetBanker(std.BankerTypeRealmSend).GetCoins(nonOwner)
	std.TestIssueCoins(std.CurrentRealm().Addr(), std.Coins{{"ugnot", expectedFees}})
	shouldNoPanic(t, ClaimOwnerFee)

	// Verify fees were claimed
	_, feesAfter := GetOwnerInfo()
	shouldEqual(t, feesAfter, int64(0))

	finalBalance := std.GetBanker(std.BankerTypeRealmSend).GetCoins(nonOwner)
	shouldEqual(t, finalBalance.AmountOf("ugnot"), initialBalance.AmountOf("ugnot")+expectedFees)
}

// Test full game flow
func TestFullGameFlow(t *testing.T) {
	setupTestGame(t)

	player1 := testutils.TestAddress("player1")
	player2 := testutils.TestAddress("player2")
	player3 := testutils.TestAddress("player3")

	// Test initial state
	shouldEqual(t, gameState.CurrentRound, int64(1))
	shouldEqual(t, gameState.KeyPrice, MIN_KEY_PRICE)
	shouldEqual(t, gameState.Ended, true)

	// Start game
	shouldNoPanic(t, StartGame)
	shouldEqual(t, gameState.Ended, false)
	shouldEqual(t, gameState.StartBlock, std.GetHeight())

	t.Run("buying keys", func(t *testing.T) {
		// Test insufficient payment
		std.TestSetOrigCaller(player1)
		std.TestIssueCoins(player1, std.Coins{{"ugnot", MIN_KEY_PRICE - 1}})
		std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE - 1}}, nil)
		std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", MIN_KEY_PRICE - 1}})
		shouldPanic(t, BuyKeys)

		// Test successful key purchase
		payment := MIN_KEY_PRICE * 3
		std.TestSetOrigSend(std.Coins{{"ugnot", payment}}, nil)
		std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", payment}})
		shouldNoPanic(t, BuyKeys)

		// Verify player state
		keys, dividends := GetPlayerInfo(player1)

		shouldEqual(t, keys, int64(3))
		shouldEqual(t, dividends, int64(0))
		shouldEqual(t, gameState.LastBuyer, player1)

		// Verify game state
		start, end, last, buyer, pot, price, keys, isEnded := GetGameState()
		shouldEqual(t, buyer, player1)
		shouldEqual(t, keys, int64(3))
		shouldEqual(t, isEnded, false)
		shouldEqual(t, pot, payment*JACKPOT_PERCENT/100)

		// Verify owner fee
		_, ownerFees := GetOwnerInfo()
		shouldEqual(t, ownerFees, payment*OWNER_FEE_PERCENT/100)
	})

	t.Run("dividend distribution and claiming", func(t *testing.T) {
		// Player 2 buys keys
		std.TestSetOrigCaller(player2)
		payment := gameState.KeyPrice * 2 // Buy 2 keys using current keyPrice
		std.TestSetOrigSend(std.Coins{{"ugnot", payment}}, nil)
		std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", payment}})
		shouldNoPanic(t, BuyKeys)

		// Check player1 received dividends
		keys1, dividends1 := GetPlayerInfo(player1)

		shouldEqual(t, keys1, int64(3))
		expectedDividends := payment * DIVIDENDS_PERCENT / 100 * 3 / gameState.TotalKeys
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

			lastBuyerBalance := std.GetBanker(std.BankerTypeRealmSend).GetCoins(player1)
			shouldEqual(t, lastBuyerBalance.AmountOf("ugnot"), initialBalance.AmountOf("ugnot")+expectedDividends)
		}
	})

	t.Run("game ending", func(t *testing.T) {
		// Try ending too early
		shouldPanic(t, EndGame)

		// Skip to end of game
		std.TestSkipHeights(MAX_TIME_EXTEND + 1)

		// End game successfully
		shouldNoPanic(t, EndGame)
		shouldEqual(t, gameState.Ended, true)
		shouldEqual(t, gameState.CurrentRound, int64(2))

		// Verify winner received jackpot
		lastBuyerBalance := std.GetBanker(std.BankerTypeRealmSend).GetCoins(gameState.LastBuyer)
		shouldEqual(t, lastBuyerBalance.AmountOf("ugnot"), gameState.Jackpot)
	})

	// Test new round
	t.Run("new round", func(t *testing.T) {
		// Calculate expected next pot from previous round
		payment1 := MIN_KEY_PRICE * 3
		// After buying 3 keys, price increased by 3% (1% per key)
		secondKeyPrice := MIN_KEY_PRICE + (MIN_KEY_PRICE * 3 / 100)
		payment2 := secondKeyPrice * 2
		expectedNextPot := (payment1 * NEXT_ROUND_POT / 100) + (payment2 * NEXT_ROUND_POT / 100)

		// Start new round
		shouldNoPanic(t, StartGame)
		shouldEqual(t, gameState.Ended, false)
		shouldEqual(t, gameState.CurrentRound, int64(2))

		shouldEqual(t, gameState.Jackpot, expectedNextPot) // New round starts with previous round's nextPot
		shouldEqual(t, gameState.NextPot, int64(0))
	})
}

// Test individual components
func TestStartGame(t *testing.T) {
	setupTestGame(t)

	// Test starting first game
	shouldNoPanic(t, StartGame)
	shouldEqual(t, gameState.Ended, false)
	shouldEqual(t, gameState.StartBlock, std.GetHeight())

	// Test cannot start while game in progress
	shouldPanic(t, StartGame)
}

func TestBuyKeys(t *testing.T) {
	setupTestGame(t)
	StartGame()

	player := testutils.TestAddress("player")
	std.TestSetOrigCaller(player)

	// Test invalid coin denomination
	std.TestIssueCoins(player, std.Coins{{"invalid", MIN_KEY_PRICE}})
	std.TestSetOrigSend(std.Coins{{"invalid", MIN_KEY_PRICE}}, nil)
	std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"invalid", MIN_KEY_PRICE}})
	shouldPanic(t, BuyKeys)

	// Test multiple coin types
	std.TestIssueCoins(player, std.Coins{{"ugnot", MIN_KEY_PRICE}, {"other", 100}})
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE}, {"other", 100}}, nil)
	std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", MIN_KEY_PRICE}, {"other", 100}})
	shouldPanic(t, BuyKeys)

	// Test insufficient payment
	std.TestIssueCoins(player, std.Coins{{"ugnot", MIN_KEY_PRICE - 1}})
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE - 1}}, nil)
	std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", MIN_KEY_PRICE - 1}})
	shouldPanic(t, BuyKeys)

	// Test successful purchase
	std.TestIssueCoins(player, std.Coins{{"ugnot", MIN_KEY_PRICE * 2}})
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE * 2}}, nil)
	std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", MIN_KEY_PRICE * 2}})
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
	std.TestIssueCoins(player, std.Coins{{"ugnot", MIN_KEY_PRICE}})
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE}}, nil)
	std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", MIN_KEY_PRICE}})
	BuyKeys()

	// Have another player buy to generate dividends
	player2 := testutils.TestAddress("player2")
	std.TestSetOrigCaller(player2)
	std.TestIssueCoins(player2, std.Coins{{"ugnot", MIN_KEY_PRICE * 2}})
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE * 2}}, nil)
	std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", MIN_KEY_PRICE * 2}})
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
	std.TestIssueCoins(player, std.Coins{{"ugnot", MIN_KEY_PRICE}})
	std.TestSetOrigSend(std.Coins{{"ugnot", MIN_KEY_PRICE}}, nil)
	std.TestIssueCoins(std.GetOrigPkgAddr(), std.Coins{{"ugnot", MIN_KEY_PRICE}})
	shouldNoPanic(t, BuyKeys)

	// Test ending too early
	shouldPanic(t, EndGame)

	// Skip to well past the end
	std.TestSkipHeights(MAX_TIME_EXTEND + TIME_EXTENSION + 1)

	shouldNoPanic(t, EndGame)
}
