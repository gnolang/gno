package emission

import (
	"std"
	"time"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gns"
)

var (
	// leftGNSAmount tracks undistributed GNS tokens from previous distributions
	leftGNSAmount uint64
	// lastExecutedHeight stores the last block height when distribution was executed
	lastExecutedHeight int64
)

// MintAndDistributeGns mints new GNS tokens and distributes them to targets
// Returns the total amount of GNS distributed
func MintAndDistributeGns() uint64 {
	assertOnlyNotHalted()

	currentHeight := std.GetHeight()
	lastMintedHeight := gns.GetLastMintedHeight()
	if lastMintedHeight >= currentHeight {
		// Skip if we've already minted tokens at this height
		return 0
	}
	// Mint new tokens and add any leftover amounts from previous distribution
	mintedEmissionRewardAmount := gns.MintGns(consts.EMISSION_ADDR)

	distributableAmount := mintedEmissionRewardAmount
	prevLeftAmount := GetLeftGNSAmount()
	if hasLeftGNSAmount() {
		distributableAmount += prevLeftAmount
		setLeftGNSAmount(0)
	}
	// Distribute tokens and track any undistributed amount
	distributedGNSAmount := distributeToTarget(distributableAmount)
	if distributableAmount != distributedGNSAmount {
		setLeftGNSAmount(distributableAmount - distributedGNSAmount)
	}

	// Emit event with distribution details
	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"MintAndDistributeGns",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"lastHeight", formatInt(lastExecutedHeight),
		"currentHeight", formatInt(currentHeight),
		"currentTimestamp", formatInt(time.Now().Unix()),
		"mintedAmount", formatUint(mintedEmissionRewardAmount),
		"prevLeftAmount", formatUint(prevLeftAmount),
		"distributedAmount", formatUint(distributedGNSAmount),
		"currentLeftAmount", formatUint(GetLeftGNSAmount()),
		"gnsTotalSupply", formatUint(gns.TotalSupply()),
	)

	setLastExecutedHeight(currentHeight)

	return distributedGNSAmount
}

// GetLeftGNSAmount returns the amount of undistributed GNS tokens
func GetLeftGNSAmount() uint64 {
	return leftGNSAmount
}

// setLeftGNSAmount updates the undistributed GNS token amount
func setLeftGNSAmount(amount uint64) {
	leftGNSAmount = amount
}

// GetLastExecutedHeight returns the last block height when distribution was executed
func GetLastExecutedHeight() int64 {
	return lastExecutedHeight
}

// setLastExecutedHeight updates the last executed block height
func setLastExecutedHeight(height int64) {
	lastExecutedHeight = height
}

// hasLeftGNSAmount checks if there are any undistributed GNS tokens
func hasLeftGNSAmount() bool {
	return leftGNSAmount > 0
}
