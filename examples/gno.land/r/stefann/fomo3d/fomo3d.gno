package fomo3d

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/r/leon/hof"
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
	JACKPOT_PERCENT   int64 = 47 // 47% goes to jackpot
	DIVIDENDS_PERCENT int64 = 28 // 28% distributed to key holders
	NEXT_ROUND_POT    int64 = 20 // 20% goes to next round's starting pot
	OWNER_FEE_PERCENT int64 = 5  // 5% goes to contract owner
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
	ownerFee int64 // accumulated owner fees

	lastBuyer std.Address // last key buyer (potential winner)
	players   *avl.Tree   // maps address -> PlayerInfo
	totalKeys int64      // total number of keys in circulation
	ended     bool

	// Ownable instance
	Ownable *ownable.Ownable
)

// init initializes the first round
func init() {
	keyPrice = MIN_KEY_PRICE
	currentRound = 1
	Ownable = ownable.New() // Initialize with caller as owner
	hof.Register()
}

// StartGame starts a new game round
func StartGame() {
	if !ended && gameStartBlock != 0 {
		panic(ErrGameInProgress.Error())
	}

	gameStartBlock = std.GetHeight()
	gameEndBlock = gameStartBlock + MAX_TIME_EXTEND
	lastKeyBlock = gameStartBlock

	// Initialize round with previous nextPot
	jackpot = nextPot
	nextPot = 0
	ended = false

	// Clear previous round's player data
	players = avl.NewTree()
	totalKeys = 0
}

// BuyKeys allows players to purchase keys
func BuyKeys() {
	if ended {
		panic(ErrGameEnded.Error())
	}

	currentBlock := std.GetHeight()
	if currentBlock > gameEndBlock {
		panic(ErrGameTimeExpired.Error())
	}

	// Get sent coins
	sent := std.GetOrigSend()
	if len(sent) != 1 || sent[0].Denom != "ugnot" {
		panic(ErrInvalidPayment.Error())
	}

	payment := sent[0].Amount
	if payment < keyPrice {
		panic(ErrInsufficientPayment.Error())
	}

	// Calculate number of keys that can be bought and actual cost
	numKeys := payment / keyPrice
	actualCost := numKeys * keyPrice
	excess := payment - actualCost

	// Update buyer's info
	buyer := std.GetOrigCaller()
	var buyerInfo PlayerInfo
	if info, exists := players.Get(buyer.String()); exists {
		buyerInfo = info.(PlayerInfo)
	}
	buyerInfo.Keys += numKeys
	totalKeys += numKeys

	// Distribute actual cost
	jackpotShare := actualCost * JACKPOT_PERCENT / 100
	dividendShare := actualCost * DIVIDENDS_PERCENT / 100
	nextPotShare := actualCost * NEXT_ROUND_POT / 100
	ownerShare := actualCost * OWNER_FEE_PERCENT / 100

	// Update pools
	jackpot += jackpotShare
	nextPot += nextPotShare
	ownerFee += ownerShare

	// Return excess payment to buyer if any
	if excess > 0 {
		banker := std.GetBanker(std.BankerTypeOrigSend)
		banker.SendCoins(
			std.CurrentRealm().Addr(),
			buyer,
			std.Coins{{"ugnot", excess}},
		)
	}

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

	// Increase key price by 1% for each key bought
	keyPrice += (keyPrice * numKeys) / 100

	// Save buyer's updated info
	players.Set(buyer.String(), buyerInfo)
}

// ClaimDividends allows players to withdraw their earned dividends
func ClaimDividends() {
	caller := std.GetOrigCaller()

	info, exists := players.Get(caller.String())
	if !exists {
		panic(ErrNoDividendsToClaim.Error())
	}

	playerInfo := info.(PlayerInfo)
	if playerInfo.Dividends == 0 {
		panic(ErrNoDividendsToClaim.Error())
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

// ClaimOwnerFee allows the owner to withdraw accumulated fees
func ClaimOwnerFee() {
	if !Ownable.CallerIsOwner() {
		panic(ownable.ErrUnauthorized.Error())
	}

	if ownerFee == 0 {
		panic(ErrNoFeesToClaim.Error())
	}

	amount := ownerFee
	ownerFee = 0

	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Addr(),
		Ownable.Owner(),
		std.Coins{{"ugnot", amount}},
	)
}

// TransferOwnership allows the current owner to transfer ownership to a new address
func TransferOwnership(newOwner std.Address) {
	err := Ownable.TransferOwnership(newOwner)
	if err != nil {
		panic(err.Error())
	}
}

// EndGame ends the current round and distributes the jackpot
func EndGame() {
	if ended {
		panic(ErrGameEnded.Error())
	}

	currentBlock := std.GetHeight()
	if currentBlock <= gameEndBlock {
		panic(ErrGameNotInProgress.Error())
	}

	if lastBuyer == "" {
		panic(ErrNoKeysPurchased.Error())
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

// GetOwnerInfo returns the owner address and unclaimed fees
func GetOwnerInfo() (std.Address, int64) {
	return Ownable.Owner(), ownerFee
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
	// _, ownerFees := GetOwnerInfo()
	
	var builder strings.Builder
	builder.WriteString("# FOMO3D - The Ultimate Game of Greed\n\n")
	
	// About section
	builder.WriteString("## About the Game\n\n")
	builder.WriteString("FOMO3D is a blockchain-based game that combines elements of lottery and investment mechanics. ")
	builder.WriteString("Players purchase keys using GNOT tokens, where each key purchase:\n\n")
	builder.WriteString("* Extends the game timer\n")
	builder.WriteString("* Increases the key price by 1%\n")
	builder.WriteString("* Makes you the potential winner of the jackpot\n")
	builder.WriteString("* Distributes dividends to all key holders\n\n")
	builder.WriteString("**How to Win:** Be the last person to buy a key before the timer expires!\n\n")
	builder.WriteString("**Rewards Distribution:**\n")
	builder.WriteString("* 47% goes to the jackpot (for the winner)\n")
	builder.WriteString("* 28% distributed as dividends to all key holders\n")
	builder.WriteString("* 20% goes to next round's starting pot\n")
	builder.WriteString("* 5% development fee for continuous improvement\n\n")

	// Game Status section
	builder.WriteString("## Game Status\n\n")
	if gameStart == 0 {
		builder.WriteString("🔴 Game has not started yet.\n\n")
	} else {
		if isEnded {
			builder.WriteString("🔴 **Game Status:** Ended\n")
			builder.WriteString(ufmt.Sprintf("🏆 **Winner:** %s\n", lastBuyer))
		} else {
			builder.WriteString("🟢 **Game Status:** Active\n")
			builder.WriteString(ufmt.Sprintf("⏱️ **Time Remaining:** %d blocks\n", gameEnd - std.GetHeight()))
		}
		builder.WriteString(ufmt.Sprintf("💰 **Current Jackpot:** %d ugnot\n", pot))
		builder.WriteString(ufmt.Sprintf("🔑 **Current Key Price:** %d ugnot\n", price))
		builder.WriteString(ufmt.Sprintf("📊 **Total Keys:** %d\n", totalKeys))
		builder.WriteString(ufmt.Sprintf("👤 **Last Buyer:** %s\n", lastBuyer))
		builder.WriteString(ufmt.Sprintf("🎮 **Next Round Pot:** %d ugnot\n", nextPot))
		builder.WriteString(ufmt.Sprintf("🔄 **Current Round:** %d\n\n", currentRound))
	}

	// Play Game section
	builder.WriteString("## Play Game\n\n")
	builder.WriteString("### How to Play:\n")
	builder.WriteString("1. **Buy Keys** - Send GNOT to this realm with function `BuyKeys()`\n")
	builder.WriteString("2. **Collect Dividends** - Call `ClaimDividends()` to collect your earnings\n")
	builder.WriteString("3. **Check Your Stats** - Visit `/player/YOUR_ADDRESS` to see your keys and dividends\n")
	if isEnded {
		builder.WriteString("4. **Start New Round** - Call `StartGame()` to begin a new round\n")
	}
	builder.WriteString("\n")

	// Report Bug section
	builder.WriteString("## Report a Bug\n\n")
	builder.WriteString("Something unusual happened? Help us improve the game by reporting bugs!\n")
	builder.WriteString("Visit our GitHub repository: https://github.com/gnolang/gno/issues\n")
	builder.WriteString("Please include:\n")
	builder.WriteString("* Detailed description of what happened\n")
	builder.WriteString("* Transaction hash (if applicable)\n")
	builder.WriteString("* Your address\n")
	builder.WriteString("* Current round number\n")

	return builder.String()
}

// RenderPlayer renders specific player information
func RenderPlayer(addr std.Address, keys int64, dividends int64) string {
	var builder strings.Builder
	builder.WriteString(ufmt.Sprintf("# Player Stats: %s\n\n", addr))
	builder.WriteString("## Your Holdings\n\n")
	builder.WriteString(ufmt.Sprintf("🔑 **Keys Owned:** %d\n", keys))
	builder.WriteString(ufmt.Sprintf("💰 **Unclaimed Dividends:** %d ugnot\n\n", dividends))
	
	builder.WriteString("## Actions\n\n")
	builder.WriteString("* To buy more keys, send GNOT to this realm with `BuyKeys()`\n")
	if dividends > 0 {
		builder.WriteString("* You have unclaimed dividends! Call `ClaimDividends()` to collect them\n")
	}
	
	return builder.String()
}
