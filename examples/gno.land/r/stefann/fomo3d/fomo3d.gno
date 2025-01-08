package fomo3d

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
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
// - The last person to buy a key before the timer expires wins the jackpot (47% of all purchases)
// - Key holders earn dividends from each purchase (28% of all purchases)
// - 20% of purchases go to the next round's starting pot
// - 5% goes to development fee
// - Game ends when the timer expires
//
// Inspired by the original Ethereum FOMO3D game but implemented in Gno.

const (
	MIN_KEY_PRICE  int64 = 100000 // minimum key price in ugnot
	TIME_EXTENSION int64 = 86400  // time extension in blocks when new key is bought (~24 hours @ 1s blocks)

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

// GameState represents the current state of the FOMO3D game
type GameState struct {
	StartBlock         int64       // Block when the game started
	EndBlock           int64       // Block when the game will end
	LastKeyBlock       int64       // Block of last key purchase
	LastBuyer          std.Address // Address of last key buyer
	Jackpot            int64       // Current jackpot in ugnot
	KeyPrice           int64       // Current price of keys in ugnot
	TotalKeys          int64       // Total number of keys in circulation
	Ended              bool        // Whether the game has ended
	CurrentRound       int64       // Current round number
	NextPot            int64       // Next round's starting pot
	OwnerFee           int64       // Accumulated owner fees
	BuyKeysLink        string      // Link to BuyKeys function
	ClaimDividendsLink string      // Link to ClaimDividends function
}

var (
	gameState GameState
	players   *avl.Tree // maps address -> PlayerInfo
	Ownable   *ownable.Ownable
)

func init() {
	Ownable = ownable.New()
	hof.Register()
}

// StartGame starts a new game round
func StartGame() {
	if !gameState.Ended && gameState.StartBlock != 0 {
		panic(ErrGameInProgress.Error())
	}

	gameState.StartBlock = std.GetHeight()
	gameState.EndBlock = gameState.StartBlock + TIME_EXTENSION // Initial 24h window
	gameState.LastKeyBlock = gameState.StartBlock
	gameState.Jackpot = gameState.NextPot
	gameState.NextPot = 0
	gameState.Ended = false
	gameState.KeyPrice = MIN_KEY_PRICE
	gameState.TotalKeys = 0

	// Clear previous round's player data
	players = avl.NewTree()

	EmitGameStarted(
		gameState.CurrentRound,
		gameState.StartBlock,
		gameState.EndBlock,
		gameState.Jackpot,
	)
}

// BuyKeys allows players to purchase keys
func BuyKeys() {
	if gameState.Ended {
		panic(ErrGameEnded.Error())
	}

	currentBlock := std.GetHeight()
	if currentBlock > gameState.EndBlock {
		panic(ErrGameTimeExpired.Error())
	}

	// Get sent coins
	sent := std.GetOrigSend()
	if len(sent) != 1 || sent[0].Denom != "ugnot" {
		panic(ErrInvalidPayment.Error())
	}

	payment := sent[0].Amount
	if payment < gameState.KeyPrice {
		panic(ErrInsufficientPayment.Error())
	}

	// Calculate number of keys that can be bought and actual cost
	numKeys := payment / gameState.KeyPrice
	actualCost := numKeys * gameState.KeyPrice
	excess := payment - actualCost

	// Update buyer's info
	buyer := std.GetOrigCaller()
	var buyerInfo PlayerInfo
	if info, exists := players.Get(buyer.String()); exists {
		buyerInfo = info.(PlayerInfo)
	}
	buyerInfo.Keys += numKeys
	gameState.TotalKeys += numKeys

	// Distribute actual cost
	jackpotShare := actualCost * JACKPOT_PERCENT / 100
	dividendShare := actualCost * DIVIDENDS_PERCENT / 100
	nextPotShare := actualCost * NEXT_ROUND_POT / 100
	ownerShare := actualCost * OWNER_FEE_PERCENT / 100

	// Update pools
	gameState.Jackpot += jackpotShare
	gameState.NextPot += nextPotShare
	gameState.OwnerFee += ownerShare

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
	if players.Size() > 0 && gameState.TotalKeys > 0 {
		dividendPerKey := dividendShare / gameState.TotalKeys
		players.Iterate("", "", func(key string, value interface{}) bool {
			addr := std.Address(key)
			playerInfo := value.(PlayerInfo)
			playerInfo.Dividends += playerInfo.Keys * dividendPerKey
			players.Set(addr.String(), playerInfo)
			return false
		})
	}

	// Update game state
	gameState.LastBuyer = buyer
	gameState.LastKeyBlock = currentBlock
	gameState.EndBlock = currentBlock + TIME_EXTENSION // Always extend 24h from current block
	gameState.KeyPrice += (gameState.KeyPrice * numKeys) / 100

	// Save buyer's updated info
	players.Set(buyer.String(), buyerInfo)

	EmitKeysPurchased(
		buyer,
		numKeys,
		gameState.KeyPrice,
		jackpotShare,
		dividendShare,
	)
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

	EmitDividendsClaimed(caller, amount)
}

// ClaimOwnerFee allows the owner to withdraw accumulated fees
func ClaimOwnerFee() {
	Ownable.AssertCallerIsOwner()

	if gameState.OwnerFee == 0 {
		panic(ErrNoFeesToClaim.Error())
	}

	amount := gameState.OwnerFee
	gameState.OwnerFee = 0

	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Addr(),
		Ownable.Owner(),
		std.Coins{{"ugnot", amount}},
	)

	EmitOwnerFeeClaimed(Ownable.Owner(), amount)
}

// TransferOwnership allows the current owner to transfer ownership to a new address
func TransferOwnership(newOwner std.Address) {
	Ownable.AssertCallerIsOwner()
	Ownable.TransferOwnership(newOwner)
	EmitOwnershipTransferred(Ownable.Owner(), newOwner)
}

// EndGame ends the current round and distributes the jackpot
func EndGame() {
	if gameState.Ended {
		panic(ErrGameEnded.Error())
	}

	currentBlock := std.GetHeight()
	if currentBlock <= gameState.EndBlock {
		panic(ErrGameNotInProgress.Error())
	}

	if gameState.LastBuyer == "" {
		panic(ErrNoKeysPurchased.Error())
	}

	gameState.Ended = true
	gameState.CurrentRound++

	// Send jackpot to winner
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Addr(),
		gameState.LastBuyer,
		std.Coins{{"ugnot", gameState.Jackpot}},
	)

	EmitGameEnded(
		gameState.CurrentRound,
		gameState.LastBuyer,
		gameState.Jackpot,
	)
}

// GetGameState returns current game state
func GetGameState() (int64, int64, int64, std.Address, int64, int64, int64, bool, int64, int64) {
	return gameState.StartBlock,
		gameState.EndBlock,
		gameState.LastKeyBlock,
		gameState.LastBuyer,
		gameState.Jackpot,
		gameState.KeyPrice,
		gameState.TotalKeys,
		gameState.Ended,
		gameState.NextPot,
		gameState.CurrentRound
}

// Helper to convert string (address or username) to address
func stringToAddress(input string) std.Address {
	// Check if input is valid address
	addr := std.Address(input)
	if addr.IsValid() {
		return addr
	}

	// Not an address, try to find namespace
	if user := users.GetUserByName(input); user != nil {
		return user.Address
	}

	return ""
}

// GetPlayerInfo returns a player's keys and dividends
func GetPlayerInfo(addrOrName interface{}) (int64, int64) {
	var addr std.Address

	switch v := addrOrName.(type) {
	case std.Address:
		addr = v
	case string:
		addr = stringToAddress(v)
		if addr == "" {
			panic(ErrInvalidAddressOrName.Error())
		}
	default:
		return 0, 0
	}

	if info, exists := players.Get(addr.String()); exists {
		playerInfo := info.(PlayerInfo)
		return playerInfo.Keys, playerInfo.Dividends
	}
	return 0, 0
}

// GetOwnerInfo returns the owner address and unclaimed fees
func GetOwnerInfo() (std.Address, int64) {
	return Ownable.Owner(), gameState.OwnerFee
}

// Render handles the rendering of game state
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return RenderHome()
	case c == 2 && parts[0] == "player":
		addr := stringToAddress(parts[1])
		if addr == "" {
			return "Invalid address or unregistered username. Please use a valid Gno address or registered username.\n\n"
		}
		keys, dividends := GetPlayerInfo(addr)
		return RenderPlayer(addr, keys, dividends)
	default:
		return "404: Invalid path\n\n"
	}
}

// RenderHome renders the main game state
func RenderHome() string {
	var builder strings.Builder
	builder.WriteString("# FOMO3D - The Ultimate Game of Greed\n\n")

	// About section
	builder.WriteString("## About the Game\n\n")
	builder.WriteString("FOMO3D is a game that combines elements of lottery and investment mechanics. ")
	builder.WriteString("Players purchase keys using GNOT tokens, where each key purchase:\n\n")
	builder.WriteString("* Extends the game timer\n")
	builder.WriteString("* Increases the key price by 1%\n")
	builder.WriteString("* Makes you the potential winner of the jackpot\n")
	builder.WriteString("* Distributes dividends to all key holders\n\n")
	builder.WriteString("## How to Win\n\n")
	builder.WriteString("* Be the last person to buy a key before the timer expires!\n\n")
	builder.WriteString("**Rewards Distribution:**\n")
	builder.WriteString("* 47% goes to the jackpot (for the winner)\n")
	builder.WriteString("* 28% distributed as dividends to all key holders\n")
	builder.WriteString("* 20% goes to next round's starting pot\n")
	builder.WriteString("* 5% development fee for continuous improvement\n\n")

	// Game Status section
	builder.WriteString("## CurrentGame Status\n\n")
	if gameState.StartBlock == 0 {
		builder.WriteString("🔴 Game has not started yet.\n\n")
	} else {
		if gameState.Ended {
			builder.WriteString("🔴 **Game Status:** Ended\n")
			builder.WriteString(ufmt.Sprintf("🏆 **Winner:** %s\n\n", gameState.LastBuyer))
		} else {
			builder.WriteString("🟢 **Game Status:** Active\n\n")
			builder.WriteString(ufmt.Sprintf("⏱️ **Time Remaining:** %d blocks\n\n", gameState.EndBlock-std.GetHeight()))
		}
		builder.WriteString(ufmt.Sprintf("💰 **Current Jackpot:** %d ugnot\n\n", gameState.Jackpot))
		builder.WriteString(ufmt.Sprintf("🔑 **Current Key Price:** %d ugnot\n\n", gameState.KeyPrice))
		builder.WriteString(ufmt.Sprintf("📊 **Total Keys:** %d\n\n", gameState.TotalKeys))
		builder.WriteString(ufmt.Sprintf("👤 **Last Buyer:** %s\n\n", getDisplayName(gameState.LastBuyer)))
		builder.WriteString(ufmt.Sprintf("🎮 **Next Round Pot:** %d ugnot\n\n", gameState.NextPot))
		builder.WriteString(ufmt.Sprintf("🔄 **Current Round:** %d\n\n", gameState.CurrentRound))
	}

	// Play Game section
	builder.WriteString("## How to Play\n\n")
	builder.WriteString(ufmt.Sprintf("1. **Buy Keys** - Send GNOT to this realm with function [`BuyKeys()`](%s)\n", gameState.BuyKeysLink))
	builder.WriteString(ufmt.Sprintf("2. **Collect Dividends** - Call [`ClaimDividends()`](%s) to collect your earnings\n", gameState.ClaimDividendsLink))
	builder.WriteString("3. **Check Your Stats** - Append `:player/` followed by your address or namespace to the current URL to view your keys and dividends\n")
	if gameState.Ended {
		builder.WriteString("4. **Start New Round** - Call `StartGame()` to begin a new round\n")
	}
	builder.WriteString("\n")

	// Report Bug section
	builder.WriteString("## Report a Bug\n\n")
	builder.WriteString("Something unusual happened? Help us improve the game by reporting bugs!\n")
	builder.WriteString("[Visit our GitHub repository](https://github.com/gnolang/gno/issues)\n\n")
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
	displayName := getDisplayName(addr)
	builder.WriteString(ufmt.Sprintf("# Player Stats: %s\n\n", displayName))
	builder.WriteString("## Your Holdings\n\n")
	builder.WriteString(ufmt.Sprintf("🔑 **Keys Owned:** %d\n\n", keys))
	builder.WriteString(ufmt.Sprintf("💰 **Unclaimed Dividends:** %d ugnot\n\n", dividends))

	builder.WriteString("## Actions\n\n")
	builder.WriteString("* To buy more keys, send GNOT to this realm with `BuyKeys()`\n")
	if dividends > 0 {
		builder.WriteString("* You have unclaimed dividends! Call `ClaimDividends()` to collect them\n")
	}

	return builder.String()
}

// Helper to get display name - just returns namespace if exists, otherwise address
func getDisplayName(addr std.Address) string {
	if user := users.GetUserByAddress(addr); user != nil {
		return user.Name
	}
	return addr.String()
}

// updates the links for game functions
func UpdateFunctionLinks(buyKeysLink string, claimDividendsLink string) {
	Ownable.AssertCallerIsOwner()
	gameState.BuyKeysLink = buyKeysLink
	gameState.ClaimDividendsLink = claimDividendsLink
}
