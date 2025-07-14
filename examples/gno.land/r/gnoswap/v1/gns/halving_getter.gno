package gns

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/json"
)

// GetHalvingBlocksInRange returns halving blocks and emissions per block within the range
func GetHalvingBlocksInRange(startHeight, endHeight int64) ([]int64, []int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingBlocks := make([]int64, 0)
	halvingEmissions := make([]int64, 0)

	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		startBlock := halvingData.getStartBlockHeight(year)
		if startBlock < startHeight {
			continue
		}
		if endHeight < startBlock {
			break
		}
		halvingBlocks = append(halvingBlocks, startBlock)
		halvingEmissions = append(halvingEmissions, halvingData.getAmountPerBlock(year))
	}
	return halvingBlocks, halvingEmissions
}

// GetEmission returns the emission amount for the current block,
// **and** updates that year's mintedAmount and leftAmount.
func GetEmission() int64 {
	es := GetEmissionState()
	hd := &es.halvingData
	height := std.ChainHeight()

	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		startBlock := hd.getStartBlockHeight(year)
		endBlock := hd.getEndBlockHeight(year)

		// include both startBlock and endBlock
		if startBlock <= height && height <= endBlock {
			amount := hd.getAmountPerBlock(year)

			// update mintedAmount and leftAmount
			prevMinted := hd.getMintedAmount(year)
			hd.setMintedAmount(year, prevMinted+amount)

			prevLeft := hd.getLeftAmount(year)
			hd.setLeftAmount(year, prevLeft-amount)

			return amount
		}
	}

	return 0
}

// GetHalvingYearStartBlock returns the starting block height for a specific halving year
func GetHalvingYearStartBlock(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getStartBlockHeight(year)
}

// setHalvingYearStartBlock sets the starting block height for a specific halving year
func setHalvingYearStartBlock(year int64, block int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setStartBlockHeight(year, block)
	GetEmissionState().setHalvingData(halvingData)
}

// GetHalvingYearEndBlock returns the ending block height for a specific halving year
func GetHalvingYearEndBlock(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getEndBlockHeight(year)
}

// setHalvingYearEndBlock sets the ending block height for a specific halving year
func setHalvingYearEndBlock(year int64, block int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setEndBlockHeight(year, block)
	GetEmissionState().setHalvingData(halvingData)
}

// GetHalvingYearTimestamp returns the starting timestamp for a specific halving year
func GetHalvingYearTimestamp(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getStartTimestamp(year)
}

// setHalvingYearTimestamp sets the starting timestamp for a specific halving year
func setHalvingYearTimestamp(year int64, timestamp int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setStartTimestamp(year, timestamp)
	GetEmissionState().setHalvingData(halvingData)
}

// GetHalvingYearMaxAmount returns the maximum issuance amount for a specific halving year
func GetHalvingYearMaxAmount(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getMaxAmount(year)
}

// setHalvingYearMaxAmount sets the maximum issuance amount for a specific halving year
func setHalvingYearMaxAmount(year, amount int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setMaxAmount(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

// GetHalvingYearMintAmount returns the minted amount for a specific halving year
func GetHalvingYearMintAmount(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getMintedAmount(year)
}

// setHalvingYearMintAmount sets the minted amount for a specific halving year
func setHalvingYearMintAmount(year, amount int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setMintedAmount(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

// GetHalvingYearLeftAmount returns the remaining issuance amount for a specific halving year
func GetHalvingYearLeftAmount(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getLeftAmount(year)
}

// setHalvingYearLeftAmount sets the remaining issuance amount for a specific halving year
func setHalvingYearLeftAmount(year, amount int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setLeftAmount(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

// GetHalvingYearAccuAmount returns the accumulated issuance amount for a specific halving year
func GetHalvingYearAccuAmount(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getAccumAmount(year)
}

// setHalvingYearAccuAmount sets the accumulated issuance amount for a specific halving year
func setHalvingYearAccuAmount(year, amount int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setAccumAmount(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

// GetAmountPerBlockPerHalvingYear returns the issuance amount per block for a specific halving year
func GetAmountPerBlockPerHalvingYear(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getAmountPerBlock(year)
}

// setAmountPerBlockPerHalvingYear sets the issuance amount per block for a specific halving year
func setAmountPerBlockPerHalvingYear(year, amount int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setAmountPerBlock(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

// GetHalvingAmountsPerYear returns the halving amount for a specific year
func GetHalvingAmountsPerYear(year int64) int64 {
	return halvingAmountsPerYear[year-1]
}

// GetEndHeight returns the end block height of the emission period
func GetEndHeight() int64 {
	return GetHalvingYearEndBlock(HALVING_END_YEAR)
}

// GetEndTimestamp returns the end timestamp of the emission period
func GetEndTimestamp() int64 {
	return GetEmissionState().getEndTimestamp()
}

// setEndTimestamp sets the end timestamp of the emission period
func setEndTimestamp(timestamp int64) {
	GetEmissionState().setEndTimestamp(timestamp)
}

// GetHalvingInfo returns the halving information in JSON format
func GetHalvingInfo() string {
	height := std.ChainHeight()
	now := time.Now().Unix()

	halvings := make([]*json.Node, 0)

	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		halvings = append(halvings, json.ObjectNode("", map[string]*json.Node{
			"year":   json.StringNode("year", strconv.FormatInt(year, 10)),
			"block":  json.NumberNode("block", float64(GetHalvingYearStartBlock(year))),
			"amount": json.NumberNode("amount", float64(GetAmountPerBlockPerHalvingYear(year))),
		}))
	}

	node := json.ObjectNode("", map[string]*json.Node{
		"height":         json.NumberNode("height", float64(height)),
		"timestamp":      json.NumberNode("timestamp", float64(now)),
		"avgBlockTimeMs": json.NumberNode("avgBlockTimeMs", float64(avgBlockTimeMs)),
		"halvings":       json.ArrayNode("", halvings),
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

// isEmissionEnded checks if the emission period has ended based on the given block height
func isEmissionEnded(height int64) bool {
	return height > GetEndHeight()
}

// getHalvingYearAndEndTimestamp returns the halving year and end timestamp corresponding to the given timestamp
func getHalvingYearAndEndTimestamp(timestamp int64) (int64, int64) {
	state := GetEmissionState()

	endTimestamp := state.getEndTimestamp()
	startTimestamp := state.getStartTimestamp()

	if timestamp > endTimestamp { // After 12 years
		return 0, 0
	}

	// Calculate year based on elapsed time from start timestamp
	timestamp -= startTimestamp
	year := timestamp / SECONDS_IN_YEAR
	year += 1 // Convert to 1-indexed year

	return year, startTimestamp + (SECONDS_IN_YEAR * year)
}
