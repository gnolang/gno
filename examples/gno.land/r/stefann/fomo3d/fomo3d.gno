package fomo3d

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// FOMO3D (Fear Of Missing Out 3D) is a blockchain-based game that combines elements
// of a lottery and investment mechanics. Players purchase keys using GNOT tokens,
// where each key purchase:
// - Extends the game timer
// - Increases the key price by 1%
// - Makes the buyer the potential winner of the jackpot
// - Distributes dividends to all key holders
//
// Game Mechanics:
// - The last person to buy a key before the timer expires wins the jackpot (50% of all purchases)
// - Key holders earn dividends from each purchase (30% of all purchases)
// - 20% of purchases go to the next round's starting pot
// - Game ends when the timer expires (max ~24 hours)
//
// Inspired by the original Ethereum FOMO3D game but implemented in Gno.

const (
	MIN_KEY_PRICE   int64 = 100000 // minimum key price in ugnot
	TIME_EXTENSION  int64 = 30      // time extension in blocks when new key is bought
	MAX_TIME_EXTEND int64 = 86400   // max time that can be added (~24 hours @ 1s blocks)

	// Distribution percentages (total 100%)
	JACKPOT_PERCENT   int64 = 50 // 50% goes to jackpot
	DIVIDENDS_PERCENT int64 = 30 // 30% distributed to key holders
	NEXT_ROUND_POT    int64 = 20 // 20% goes to next round's starting pot
)

type PlayerInfo struct {
	Keys      int64 // number of keys owned
	Dividends int64 // unclaimed dividends in ugnot
}

var (
	gameStartBlock int64
	gameEndBlock   int64
	lastKeyBlock   int64
	currentRound   int64

	// Monetary pools
	jackpot  int64 // current jackpot in ugnot
	nextPot  int64 // next round's starting pot
	keyPrice int64 // current price of keys in ugnot

	lastBuyer std.Address // last key buyer (potential winner)
	players   *avl.Tree   // maps address -> PlayerInfo
	totalKeys int64      // total number of keys in circulation
	ended     bool
)

// init initializes the first round
func init() {
	keyPrice = MIN_KEY_PRICE
	currentRound = 1
}

// StartGame starts a new game round
func StartGame() {
	if !ended && gameStartBlock != 0 {
		panic("game already in progress")
	}

	gameStartBlock = std.GetHeight()
	gameEndBlock = gameStartBlock + MAX_TIME_EXTEND
	lastKeyBlock = gameStartBlock

	// Initialize round with previous nextPot
	jackpot = nextPot
	nextPot = 0
	keyPrice = MIN_KEY_PRICE
	ended = false

	// Clear previous round's player data
	players = avl.NewTree()
	totalKeys = 0
}

// BuyKeys allows players to purchase keys
func BuyKeys() {
	if ended {
		panic("game has ended")
	}

	currentBlock := std.GetHeight()
	if currentBlock > gameEndBlock {
		panic("game time expired")
	}

	// Get sent coins
	sent := std.GetOrigSend()
	if len(sent) != 1 || sent[0].Denom != "ugnot" {
		panic("must send ugnot only")
	}

	payment := sent[0].Amount
	if payment < keyPrice {
		panic("insufficient payment for key")
	}

	// Calculate number of keys that can be bought
	numKeys := payment / keyPrice

	// Update buyer's info
	buyer := std.GetOrigCaller()
	var buyerInfo PlayerInfo
	if info, exists := players.Get(buyer.String()); exists {
		buyerInfo = info.(PlayerInfo)
	}
	buyerInfo.Keys += numKeys
	totalKeys += numKeys

	// Distribute payment
	jackpotShare := payment * JACKPOT_PERCENT / 100
	dividendShare := payment * DIVIDENDS_PERCENT / 100
	nextPotShare := payment * NEXT_ROUND_POT / 100

	// Update pools
	jackpot += jackpotShare
	nextPot += nextPotShare

	// Distribute dividends to all key holders
	if players.Size() > 0 && totalKeys > 0 {
		dividendPerKey := dividendShare / totalKeys
		players.Iterate("", "", func(key string, value interface{}) bool {
			addr := std.Address(key)
			playerInfo := value.(PlayerInfo)
			playerInfo.Dividends += playerInfo.Keys * dividendPerKey
			players.Set(addr.String(), playerInfo)
			return false
		})
	}

	// Update game state
	lastBuyer = buyer
	lastKeyBlock = currentBlock

	// Extend deadline within max time limit
	newEndBlock := currentBlock + TIME_EXTENSION
	if newEndBlock > gameStartBlock+MAX_TIME_EXTEND {
		newEndBlock = gameStartBlock + MAX_TIME_EXTEND
	}
	if newEndBlock > gameEndBlock {
		gameEndBlock = newEndBlock
	}

	// Increase key price by 1%
	keyPrice += keyPrice / 100

	// Save buyer's updated info
	players.Set(buyer.String(), buyerInfo)
}

// ClaimDividends allows players to withdraw their earned dividends
func ClaimDividends() {
	caller := std.GetOrigCaller()

	info, exists := players.Get(caller.String())
	if !exists {
		panic("no dividends to claim")
	}

	playerInfo := info.(PlayerInfo)
	if playerInfo.Dividends == 0 {
		panic("no dividends to claim")
	}

	// Reset dividends and send coins
	amount := playerInfo.Dividends
	playerInfo.Dividends = 0
	players.Set(caller.String(), playerInfo)

	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Addr(),
		caller,
		std.Coins{{"ugnot", amount}},
	)
}

// EndGame ends the current round and distributes the jackpot
func EndGame() {
	if ended {
		panic("game already ended")
	}

	currentBlock := std.GetHeight()
	if currentBlock <= gameEndBlock {
		panic("game still in progress")
	}

	if lastBuyer == "" {
		panic("no keys purchased")
	}

	ended = true
	currentRound++

	// Send jackpot to winner
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Addr(),
		lastBuyer,
		std.Coins{{"ugnot", jackpot}},
	)
}

// GetGameState returns current game state
func GetGameState() (int64, int64, int64, std.Address, int64, int64, int64, bool) {
	return gameStartBlock, gameEndBlock, lastKeyBlock, lastBuyer, jackpot, keyPrice, totalKeys, ended
}

// GetPlayerInfo returns a player's keys and dividends
func GetPlayerInfo(addr std.Address) (int64, int64) {
	if info, exists := players.Get(addr.String()); exists {
		playerInfo := info.(PlayerInfo)
		return playerInfo.Keys, playerInfo.Dividends
	}
	return 0, 0
}

// Render handles the rendering of game state
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		// Default render showing game overview
		return RenderHome()
	case c == 2 && parts[0] == "player":
		// Render specific player info
		addr := std.Address(parts[1])
		keys, dividends := GetPlayerInfo(addr)
		return RenderPlayer(addr, keys, dividends)
	default:
		return "404: Invalid path\n"
	}
}

// RenderHome renders the main game state
func RenderHome() string {
	gameStart, gameEnd, _, lastBuyer, pot, price, totalKeys, isEnded := GetGameState()
	
	var builder strings.Builder
	builder.WriteString("# FOMO3D Game Status\n\n")
	
	if gameStart == 0 {
		builder.WriteString("Game has not started yet.\n")
		return builder.String()
	}

	builder.WriteString(ufmt.Sprintf("## Round %d\n\n", currentRound))
	
	if isEnded {
		builder.WriteString("**Game Status:** Ended\n")
		builder.WriteString(ufmt.Sprintf("**Winner:** %s\n", lastBuyer))
	} else {
		builder.WriteString("**Game Status:** Active\n")
		builder.WriteString(ufmt.Sprintf("**Time Remaining:** %d blocks\n", gameEnd - std.GetHeight()))
	}

	builder.WriteString("\n## Game Statistics\n\n")
	builder.WriteString(ufmt.Sprintf("* **Current Jackpot:** %d ugnot\n", pot))
	builder.WriteString(ufmt.Sprintf("* **Current Key Price:** %d ugnot\n", price))
	builder.WriteString(ufmt.Sprintf("* **Total Keys:** %d\n", totalKeys))
	builder.WriteString(ufmt.Sprintf("* **Last Buyer:** %s\n", lastBuyer))
	builder.WriteString(ufmt.Sprintf("* **Next Round Pot:** %d ugnot\n", nextPot))

	return builder.String()
}

// RenderPlayer renders specific player information
func RenderPlayer(addr std.Address, keys int64, dividends int64) string {
	var builder strings.Builder
	builder.WriteString(ufmt.Sprintf("# Player Info: %s\n\n", addr))
	builder.WriteString(ufmt.Sprintf("* **Keys Owned:** %d\n", keys))
	builder.WriteString(ufmt.Sprintf("* **Unclaimed Dividends:** %d ugnot\n", dividends))
	return builder.String()
}
