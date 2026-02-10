package fomo3d

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// Event names
const (
	// Game events
	GameStartedEvent   = "GameStarted"
	GameEndedEvent     = "GameEnded"
	KeysPurchasedEvent = "KeysPurchased"

	// Player events
	DividendsClaimedEvent = "DividendsClaimed"

	// Admin events
	OwnerFeeClaimedEvent = "OwnerFeeClaimed"
)

// Event keys
const (
	// Common keys
	EventRoundKey  = "round"
	EventAmountKey = "amount"

	// Game keys
	EventStartBlockKey  = "startBlock"
	EventEndBlockKey    = "endBlock"
	EventStartingPotKey = "startingPot"
	EventWinnerKey      = "winner"
	EventJackpotKey     = "jackpot"

	// Player keys
	EventBuyerKey         = "buyer"
	EventNumKeysKey       = "numKeys"
	EventPriceKey         = "price"
	EventJackpotShareKey  = "jackpotShare"
	EventDividendShareKey = "dividendShare"
	EventClaimerKey       = "claimer"

	// Admin keys
	EventOwnerKey         = "owner"
	EventPreviousOwnerKey = "previousOwner"
	EventNewOwnerKey      = "newOwner"
)

func emitGameStarted(round, startBlock, endBlock, startingPot int64) {
	std.Emit(
		GameStartedEvent,
		EventRoundKey, ufmt.Sprintf("%d", round),
		EventStartBlockKey, ufmt.Sprintf("%d", startBlock),
		EventEndBlockKey, ufmt.Sprintf("%d", endBlock),
		EventStartingPotKey, ufmt.Sprintf("%d", startingPot),
	)
}

func emitGameEnded(round int64, winner address, jackpot int64) {
	std.Emit(
		GameEndedEvent,
		EventRoundKey, ufmt.Sprintf("%d", round),
		EventWinnerKey, winner.String(),
		EventJackpotKey, ufmt.Sprintf("%d", jackpot),
	)
}

func emitKeysPurchased(buyer address, numKeys, price, jackpotShare, dividendShare int64) {
	std.Emit(
		KeysPurchasedEvent,
		EventBuyerKey, buyer.String(),
		EventNumKeysKey, ufmt.Sprintf("%d", numKeys),
		EventPriceKey, ufmt.Sprintf("%d", price),
		EventJackpotShareKey, ufmt.Sprintf("%d", jackpotShare),
		EventDividendShareKey, ufmt.Sprintf("%d", dividendShare),
	)
}

func emitDividendsClaimed(claimer address, amount int64) {
	std.Emit(
		DividendsClaimedEvent,
		EventClaimerKey, claimer.String(),
		EventAmountKey, ufmt.Sprintf("%d", amount),
	)
}

func emitOwnerFeeClaimed(owner address, amount int64) {
	std.Emit(
		OwnerFeeClaimedEvent,
		EventOwnerKey, owner.String(),
		EventAmountKey, ufmt.Sprintf("%d", amount),
	)
}
