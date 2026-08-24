package fomo3d

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"

	"gno.land/r/leon/hor"
	"gno.land/r/sys/users"
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
type GameState struct { // TODO: Separate GameState and RoundState and save round history tree in GameState
	StartBlock         int64   // Block when the game started
	EndBlock           int64   // Block when the game will end
	LastKeyBlock       int64   // Block of last key purchase
	LastBuyer          address // Address of last key buyer
	Jackpot            int64   // Current jackpot in ugnot
	KeyPrice           int64   // Current price of keys in ugnot
	TotalKeys          int64   // Total number of keys in circulation
	Ended              bool    // Whether the game has ended
	CurrentRound       int64   // Current round number
	NextPot            int64   // Next round's starting pot
	OwnerFee           int64   // Accumulated owner fees
	BuyKeysLink        string  // Link to BuyKeys function
	ClaimDividendsLink string  // Link to ClaimDividends function
	StartGameLink      string  // Link to StartGame function
}

var (
	gameState GameState
	players   *avl.Tree // maps address -> PlayerInfo
	Ownable   *ownable.Ownable
)

func init() {
	Ownable = ownable.NewWithOrigin()
	players = avl.NewTree()
	gameState.Ended = true
	hor.Register("FOMO3D Game!", "")
}

// StartGame starts a new game round
func StartGame() {
	if !gameState.Ended && gameState.StartBlock != 0 {
		panic(ErrGameInProgress.Error())
	}

	gameState.CurrentRound++
	gameState.StartBlock = std.ChainHeight()
	gameState.EndBlock = gameState.StartBlock + TIME_EXTENSION // Initial 24h window
	gameState.LastKeyBlock = gameState.StartBlock
	gameState.Jackpot = gameState.NextPot
	gameState.NextPot = 0
	gameState.Ended = false
	gameState.KeyPrice = MIN_KEY_PRICE
	gameState.TotalKeys = 0

	// Clear previous round's player data
	players = avl.NewTree()

	emitGameStarted(
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

	currentBlock := std.ChainHeight()
	if currentBlock > gameState.EndBlock {
		panic(ErrGameTimeExpired.Error())
	}

	// Get sent coins
	sent := std.OriginSend()
	if len(sent) != 1 || sent[0].Denom != "ugnot" {
		panic(ErrInvalidPayment.Error())
	}

	payment := sent.AmountOf("ugnot")
	if payment < gameState.KeyPrice {
		panic(ErrInsufficientPayment.Error())
	}

	// Calculate number of keys that can be bought and actual cost
	numKeys := payment / gameState.KeyPrice
	actualCost := numKeys * gameState.KeyPrice
	excess := payment - actualCost

	// Update buyer's info
	buyer := std.PreviousRealm().Address()
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
		banker := std.NewBanker(std.BankerTypeOriginSend)
		banker.SendCoins(
			std.CurrentRealm().Address(),
			buyer,
			std.NewCoins(std.NewCoin("ugnot", excess)),
		)
	}

	// Distribute dividends to all key holders
	if players.Size() > 0 && gameState.TotalKeys > 0 {
		dividendPerKey := dividendShare / gameState.TotalKeys
		players.Iterate("", "", func(key string, value any) bool {
			playerInfo := value.(PlayerInfo)
			playerInfo.Dividends += playerInfo.Keys * dividendPerKey
			players.Set(key, playerInfo)
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

	emitKeysPurchased(
		buyer,
		numKeys,
		gameState.KeyPrice,
		jackpotShare,
		dividendShare,
	)
}

// ClaimDividends allows players to withdraw their earned dividends
func ClaimDividends() {
	caller := std.PreviousRealm().Address()

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

	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Address(),
		caller,
		std.NewCoins(std.NewCoin("ugnot", amount)),
	)

	emitDividendsClaimed(caller, amount)
}

// ClaimOwnerFee allows the owner to withdraw accumulated fees
func ClaimOwnerFee() {
	Ownable.AssertOwned()

	if gameState.OwnerFee == 0 {
		panic(ErrNoFeesToClaim.Error())
	}

	amount := gameState.OwnerFee
	gameState.OwnerFee = 0

	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Address(),
		Ownable.Owner(),
		std.NewCoins(std.NewCoin("ugnot", amount)),
	)

	emitOwnerFeeClaimed(Ownable.Owner(), amount)
}

// EndGame ends the current round and distributes the jackpot
func EndGame() {
	if gameState.Ended {
		panic(ErrGameEnded.Error())
	}

	currentBlock := std.ChainHeight()
	if currentBlock <= gameState.EndBlock {
		panic(ErrGameNotInProgress.Error())
	}

	if gameState.LastBuyer == "" {
		panic(ErrNoKeysPurchased.Error())
	}

	gameState.Ended = true

	// Send jackpot to winner
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Address(),
		gameState.LastBuyer,
		std.NewCoins(std.NewCoin("ugnot", gameState.Jackpot)),
	)

	emitGameEnded(
		gameState.CurrentRound,
		gameState.LastBuyer,
		gameState.Jackpot,
	)

	// Mint NFT for the winner
	if err := mintRoundWinnerNFT(gameState.LastBuyer, gameState.CurrentRound); err != nil {
		panic(err.Error())
	}
}

// GetGameState returns current game state
func GetGameState() (int64, int64, int64, address, int64, int64, int64, bool, int64, int64) {
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

// GetOwnerInfo returns the owner address and unclaimed fees
func GetOwnerInfo() (address, int64) {
	return Ownable.Owner(), gameState.OwnerFee
}

// Helper to convert string (address or username) to address
func stringToAddress(input string) address {
	// Check if input is valid address
	addr := address(input)
	if addr.IsValid() {
		return addr
	}

	// Not an address, try to find namespace
	if user, _ := users.ResolveName(input); user != nil {
		return user.Addr()
	}

	return ""
}

func isPlayerInGame(addr address) bool {
	_, exists := players.Get(addr.String())
	return exists
}

// GetPlayerInfo returns a player's keys and dividends
func GetPlayerInfo(addrOrName string) (int64, int64) {
	addr := stringToAddress(addrOrName)

	if addr == "" {
		panic(ErrInvalidAddressOrName.Error())
	}

	if !isPlayerInGame(addr) {
		panic(ErrPlayerNotInGame.Error())
	}

	info, _ := players.Get(addr.String())
	playerInfo := info.(PlayerInfo)
	return playerInfo.Keys, playerInfo.Dividends
}

// Render handles the rendering of game state
func Render(path string) string {
	parts := strings.Split(path, "/")
	c := len(parts)

	switch {
	case path == "":
		return RenderHome()
	case c == 2 && parts[0] == "player":
		if gameState.Ended {
			return ufmt.Sprintf("🔴 Game has not started yet.\n\n Call [`StartGame()`](%s) to start a new round.\n\n", gameState.StartGameLink)
		}
		addr := stringToAddress(parts[1])
		if addr == "" || !isPlayerInGame(addr) {
			return "Address not found in game. You need to buy keys first to view your stats.\n\n"
		}
		keys, dividends := GetPlayerInfo(parts[1])
		return RenderPlayer(addr, keys, dividends)
	default:
		return "404: Invalid path\n\n"
	}
}
