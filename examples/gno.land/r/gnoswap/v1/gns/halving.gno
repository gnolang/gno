package gns

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
	"gno.land/p/gnoswap/consts"
)

// init 12 years halving tier block
/*
	NOTE: assume block will be created every 1 second by default
	1 second = 1 block
	1 minute = 60 block
	1 hour = 3600 block
	1 day = 86400 block
	(365 days) 1 year = 31536000 block
*/

const (
	DAY_PER_YEAR    = 365
	SECONDS_IN_YEAR = 31536000

	HALVING_START_YEAR = int64(1)
	HALVING_END_YEAR   = int64(12)
)

var (
	HALVING_AMOUNTS_PER_YEAR = [HALVING_END_YEAR]uint64{
		18_750_000_000_000 * 12, // Year 1:  225000000000000
		18_750_000_000_000 * 12, // Year 2:  225000000000000
		9_375_000_000_000 * 12,  // Year 3:  112500000000000
		9_375_000_000_000 * 12,  // Year 4:  112500000000000
		4_687_500_000_000 * 12,  // Year 5:  56250000000000
		4_687_500_000_000 * 12,  // Year 6:  56250000000000
		2_343_750_000_000 * 12,  // Year 7:  28125000000000
		2_343_750_000_000 * 12,  // Year 8:  28125000000000
		1_171_875_000_000 * 12,  // Year 9:  14062500000000
		1_171_875_000_000 * 12,  // Year 10: 14062500000000
		1_171_875_000_000 * 12,  // Year 11: 14062500000000
		1_171_875_000_000 * 12,  // Year 12: 14062500000000
	}
)

var (
	blockPerYear = SECONDS_IN_YEAR / milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockPerDay  = consts.SECONDS_PER_DAY / milliToSec(consts.BLOCK_GENERATION_INTERVAL)

	avgBlockTimeMs int64 = consts.BLOCK_GENERATION_INTERVAL
	perBlockMint         = avl.NewTree() // height => uint64
)

type HalvingData struct {
	startBlockHeight []int64
	endBlockHeight   []int64
	startTimestamp   []int64
	maxAmount        []uint64
	mintedAmount     []uint64
	leftAmount       []uint64
	accumAmount      []uint64
	amountPerBlock   []uint64
}

func (h *HalvingData) getStartBlockHeight(year int64) int64 {
	if year == 0 {
		return 0
	}
	return h.startBlockHeight[year-1]
}

func (h *HalvingData) setStartBlockHeight(year int64, height int64) {
	assertValidYear(year)
	h.startBlockHeight[year-1] = height
}

func (h *HalvingData) getEndBlockHeight(year int64) int64 {
	if year == 0 {
		return 0
	}
	return h.endBlockHeight[year-1]
}

func (h *HalvingData) setEndBlockHeight(year int64, height int64) {
	assertValidYear(year)
	h.endBlockHeight[year-1] = height
}

func (h *HalvingData) getStartTimestamp(year int64) int64 {
	if year == 0 {
		return 0
	}
	return h.startTimestamp[year-1]
}

func (h *HalvingData) setStartTimestamp(year int64, timestamp int64) {
	assertValidYear(year)
	h.startTimestamp[year-1] = timestamp
}

func (h *HalvingData) getMaxAmount(year int64) uint64 {
	if year == 0 {
		return 0
	}
	return h.maxAmount[year-1]
}

func (h *HalvingData) setMaxAmount(year int64, amount uint64) {
	assertValidYear(year)
	h.maxAmount[year-1] = amount
}

func (h *HalvingData) getMintedAmount(year int64) uint64 {
	if year == 0 {
		return 0
	}
	return h.mintedAmount[year-1]
}

func (h *HalvingData) setMintedAmount(year int64, amount uint64) {
	assertValidYear(year)
	h.mintedAmount[year-1] = amount
}

func (h *HalvingData) getLeftAmount(year int64) uint64 {
	if year == 0 {
		return 0
	}
	return h.leftAmount[year-1]
}

func (h *HalvingData) setLeftAmount(year int64, amount uint64) {
	assertValidYear(year)
	h.leftAmount[year-1] = amount
}

func (h *HalvingData) getAccumAmount(year int64) uint64 {
	if year == 0 {
		return 0
	}
	return h.accumAmount[year-1]
}

func (h *HalvingData) setAccumAmount(year int64, amount uint64) {
	assertValidYear(year)
	h.accumAmount[year-1] = amount
}

func (h *HalvingData) addAccumAmount(year int64, amount uint64) {
	assertValidYear(year)
	h.accumAmount[year-1] += amount
}

func (h *HalvingData) getAmountPerBlock(year int64) uint64 {
	if year == 0 {
		return 0
	}
	return h.amountPerBlock[year-1]
}

func (h *HalvingData) setAmountPerBlock(year int64, amount uint64) {
	assertValidYear(year)
	h.amountPerBlock[year-1] = amount
}

// initialize initializes the halving data based on release schedule
// ref: https://docs.gnoswap.io/gnoswap-token/release-schedule
func (h *HalvingData) initialize(startHeight, startTimestamp int64) {
	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		// set max emission amount per year
		// each year can not mint more than this amount
		currentYearMaxAmount := GetHalvingAmountsPerYear(year)
		h.setMaxAmount(year, currentYearMaxAmount)

		if year == HALVING_START_YEAR {
			h.setAccumAmount(year, currentYearMaxAmount)
			h.setStartBlockHeight(year, startHeight)
			h.setEndBlockHeight(year, startHeight+(blockPerYear*year)-1)
		} else {
			// accumulate amount until current year, is the sum of current year max amount and accumulated amount until previous year
			h.setAccumAmount(year, currentYearMaxAmount+h.getAccumAmount(year-1))

			// start block of current year, is the next block of previous year of end block
			h.setStartBlockHeight(year, h.getEndBlockHeight(year-1)+1)

			// end block of current year, is sum of start block and block per year
			h.setEndBlockHeight(year, h.getStartBlockHeight(year)+blockPerYear-1)
		}

		h.setStartTimestamp(year, startTimestamp+(SECONDS_IN_YEAR*(year-1)))

		amountPerDay := currentYearMaxAmount / DAY_PER_YEAR
		amountPerBlock := amountPerDay / uint64(blockPerDay)
		h.setAmountPerBlock(year, uint64(amountPerBlock))
		h.setMintedAmount(year, uint64(0))
		h.setLeftAmount(year, currentYearMaxAmount)
	}
}

var callbackEmissionChange func(amount uint64)

func SetCallbackEmissionChange(callback func(amount uint64)) {
	if callbackEmissionChange != nil {
		panic("callbackEmissionChange already set")
	}
	callbackEmissionChange = callback
}

// endHeight MUST be smaller than the current height(we don't read future halving blocks with this function)
func GetHalvingBlocksInRange(startHeight, endHeight int64) ([]int64, []uint64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingBlocks := []int64{}
	halvingEmissions := []uint64{}
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

// height MUST be smaller than the current height(we don't read future emissions)
func GetEmission() uint64 {
	halvingData := GetEmissionState().getHalvingData()
	height := std.GetHeight()

	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		startBlock := halvingData.getStartBlockHeight(year)
		endBlock := halvingData.getEndBlockHeight(year)
		if startBlock <= height && height < endBlock {
			return halvingData.getAmountPerBlock(year)
		}
	}

	return 0
}

type EmissionState struct {
	startHeight    int64
	startTimestamp int64
	endTimestamp   int64
	halvingData    HalvingData
}

func GetEmissionState() *EmissionState {
	if emissionState == nil {
		emissionState = NewEmissionState()
		emissionState.initializeHalvingData()
	}
	return emissionState
}

func NewEmissionState() *EmissionState {
	now := time.Now().Unix() + milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	emissionEndTime := now + SECONDS_IN_YEAR*HALVING_END_YEAR

	return &EmissionState{
		startHeight:    std.GetHeight() + 1,
		startTimestamp: now,
		endTimestamp:   emissionEndTime,
		halvingData: HalvingData{
			startBlockHeight: make([]int64, HALVING_END_YEAR),
			endBlockHeight:   make([]int64, HALVING_END_YEAR),
			startTimestamp:   make([]int64, HALVING_END_YEAR),
			maxAmount:        make([]uint64, HALVING_END_YEAR),
			mintedAmount:     make([]uint64, HALVING_END_YEAR),
			leftAmount:       make([]uint64, HALVING_END_YEAR),
			accumAmount:      make([]uint64, HALVING_END_YEAR),
			amountPerBlock:   make([]uint64, HALVING_END_YEAR),
		},
	}
}

func (e *EmissionState) getStartHeight() int64 {
	return e.startHeight
}

func (e *EmissionState) setStartHeight(height int64) {
	e.startHeight = height
}

func (e *EmissionState) getStartTimestamp() int64 {
	return e.startTimestamp
}

func (e *EmissionState) setStartTimestamp(timestamp int64) {
	e.startTimestamp = timestamp
}

func (e *EmissionState) getEndTimestamp() int64 {
	return e.endTimestamp
}

func (e *EmissionState) setEndTimestamp(timestamp int64) {
	e.endTimestamp = timestamp
}

func (e *EmissionState) getHalvingData() HalvingData {
	return e.halvingData
}

func (e *EmissionState) setHalvingData(data HalvingData) {
	e.halvingData = data
}

// initializeHalvingData initializes the halving data
// it should be called only once, so we call this in init()
func (e *EmissionState) initializeHalvingData() {
	halvingData := e.getHalvingData()
	halvingData.initialize(e.getStartHeight(), e.getStartTimestamp())
	e.setHalvingData(halvingData)
}

var emissionState *EmissionState

func init() {
	emissionState = GetEmissionState()
}

func GetAvgBlockTimeInMs() int64 {
	return avgBlockTimeMs
}

// SetAvgBlockTimeInMsByAdmin sets the average block time in millisecond.
func SetAvgBlockTimeInMsByAdmin(ms int64) {
	assertCallerIsAdmin()

	prevAvgBlockTimeInMs := GetAvgBlockTimeInMs()
	setAvgBlockTimeInMs(ms)

	halvingData := GetEmissionState().getHalvingData()
	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"SetAvgBlockTimeInMsByAdmin",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevAvgBlockTimeMs", formatInt(prevAvgBlockTimeInMs),
		"newAvgBlockTimeMs", formatInt(ms),
		"endBlock1Year", formatInt(halvingData.getEndBlockHeight(1)),
		"endBlock2Year", formatInt(halvingData.getEndBlockHeight(2)),
		"endBlock3Year", formatInt(halvingData.getEndBlockHeight(3)),
		"endBlock4Year", formatInt(halvingData.getEndBlockHeight(4)),
		"endBlock5Year", formatInt(halvingData.getEndBlockHeight(4)),
		"endBlock6Year", formatInt(halvingData.getEndBlockHeight(5)),
		"endBlock7Year", formatInt(halvingData.getEndBlockHeight(6)),
		"endBlock8Year", formatInt(halvingData.getEndBlockHeight(7)),
		"endBlock9Year", formatInt(halvingData.getEndBlockHeight(8)),
		"endBlock10Year", formatInt(halvingData.getEndBlockHeight(9)),
		"endBlock11Year", formatInt(halvingData.getEndBlockHeight(10)),
		"endBlock12Year", formatInt(halvingData.getEndBlockHeight(11)),
	)
}

// SetAvgBlockTimeInMs sets the average block time in millisecond.
// Only governance contract can execute this function via proposal
func SetAvgBlockTimeInMs(ms int64) {
	assertCallerIsGovernance()

	prevAvgBlockTimeInMs := GetAvgBlockTimeInMs()
	setAvgBlockTimeInMs(ms)

	halvingData := GetEmissionState().getHalvingData()
	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"SetAvgBlockTimeInMs",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevAvgBlockTimeMs", formatInt(prevAvgBlockTimeInMs),
		"newAvgBlockTimeMs", formatInt(ms),
		"endBlock1Year", formatInt(halvingData.getEndBlockHeight(1)),
		"endBlock2Year", formatInt(halvingData.getEndBlockHeight(2)),
		"endBlock3Year", formatInt(halvingData.getEndBlockHeight(3)),
		"endBlock4Year", formatInt(halvingData.getEndBlockHeight(4)),
		"endBlock5Year", formatInt(halvingData.getEndBlockHeight(4)),
		"endBlock6Year", formatInt(halvingData.getEndBlockHeight(5)),
		"endBlock7Year", formatInt(halvingData.getEndBlockHeight(6)),
		"endBlock8Year", formatInt(halvingData.getEndBlockHeight(7)),
		"endBlock9Year", formatInt(halvingData.getEndBlockHeight(8)),
		"endBlock10Year", formatInt(halvingData.getEndBlockHeight(9)),
		"endBlock11Year", formatInt(halvingData.getEndBlockHeight(10)),
		"endBlock12Year", formatInt(halvingData.getEndBlockHeight(11)),
	)
}

func setAvgBlockTimeInMs(ms int64) {
	assertShouldNotBeHalted()

	now := time.Now().Unix()
	height := std.GetHeight()

	// update block per year
	yearByMilliSec := secToMs(SECONDS_IN_YEAR)
	blockPerYear = yearByMilliSec / ms

	// get the halving year and end timestamp of current time
	currentYear, currentYearEndTimestamp := getHalvingYearAndEndTimestamp(now)

	// how much time left for current halving year
	timeLeft := currentYearEndTimestamp - now
	timeLeftMs := secToMs(timeLeft)

	// how many block left for current halving year
	blockLeft := timeLeftMs / ms
	// how many reward left for current halving year
	minted := GetMintedEmissionAmount()
	amountLeft := GetHalvingYearAccuAmount(currentYear) - minted

	// how much reward should be minted per block for current halving year
	adjustedAmountPerBlock := amountLeft / uint64(blockLeft)
	// update it
	setAmountPerBlockPerHalvingYear(currentYear, adjustedAmountPerBlock)
	if callbackEmissionChange == nil {
		panic(errCallbackEmissionChangeIsNil.Error())
	}
	callbackEmissionChange(adjustedAmountPerBlock)

	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		if year < currentYear {
			// pass past halving years
			continue
		}

		yearEndTimestamp := GetHalvingYearTimestamp(year) + SECONDS_IN_YEAR
		timeLeftForYear := yearEndTimestamp - now
		numBlock := (timeLeftForYear * consts.MILLISECONDS_PER_SECOND) / ms
		yearEndHeight := height + numBlock

		if year == currentYear {
			// for current year, update only end block
			setHalvingYearEndBlock(year, yearEndHeight)
		} else {
			// update start block
			prevYearEnd := GetHalvingYearEndBlock(year - 1)
			nextYearStart := prevYearEnd + 1
			nextYearEnd := nextYearStart + blockPerYear

			setHalvingYearStartBlock(year, nextYearStart)
			setHalvingYearEndBlock(year, nextYearEnd)
		}
	}

	avgBlockTimeMs = ms

}

// GetAmountByHeight returns the amount of gns to mint by height
func GetAmountByHeight(height int64) uint64 {
	if isEmissionEnded(height) {
		return 0
	}

	halvingYear := GetHalvingYearByHeight(height)
	return GetAmountPerBlockPerHalvingYear(halvingYear)
}

// GetHalvingYearByHeight returns the halving year by height
func GetHalvingYearByHeight(height int64) int64 {
	if isEmissionEnded(height) {
		return 0
	}

	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		endBlock := GetHalvingYearEndBlock(year)
		if height <= endBlock {
			return year
		}
	}

	return 0
}

func GetHalvingYearStartBlock(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getStartBlockHeight(year)
}

func setHalvingYearStartBlock(year int64, block int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setStartBlockHeight(year, block)
	GetEmissionState().setHalvingData(halvingData)
}

func GetHalvingYearEndBlock(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getEndBlockHeight(year)
}

func setHalvingYearEndBlock(year int64, block int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setEndBlockHeight(year, block)
	GetEmissionState().setHalvingData(halvingData)
}

func GetHalvingYearTimestamp(year int64) int64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getStartTimestamp(year)
}

func setHalvingYearTimestamp(year int64, timestamp int64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setStartTimestamp(year, timestamp)
	GetEmissionState().setHalvingData(halvingData)
}

func GetHalvingYearMaxAmount(year int64) uint64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getMaxAmount(year)
}

func setHalvingYearMaxAmount(year int64, amount uint64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setMaxAmount(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

func GetHalvingYearMintAmount(year int64) uint64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getMintedAmount(year)
}

func setHalvingYearMintAmount(year int64, amount uint64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setMintedAmount(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

func GetHalvingYearLeftAmount(year int64) uint64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getLeftAmount(year)
}

func setHalvingYearLeftAmount(year int64, amount uint64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setLeftAmount(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

func GetHalvingYearAccuAmount(year int64) uint64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getAccumAmount(year)
}

func setHalvingYearAccuAmount(year int64, amount uint64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setAccumAmount(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

func GetAmountPerBlockPerHalvingYear(year int64) uint64 {
	halvingData := GetEmissionState().getHalvingData()
	return halvingData.getAmountPerBlock(year)
}

func setAmountPerBlockPerHalvingYear(year int64, amount uint64) {
	halvingData := GetEmissionState().getHalvingData()
	halvingData.setAmountPerBlock(year, amount)
	GetEmissionState().setHalvingData(halvingData)
}

func GetHalvingAmountsPerYear(year int64) uint64 {
	return HALVING_AMOUNTS_PER_YEAR[year-1]
}

func GetEndHeight() int64 {
	// last block of last halving year(12) is last block of emission
	// later than this block, no more gns will be minted
	return GetHalvingYearEndBlock(HALVING_END_YEAR)
}

func GetEndTimestamp() int64 {
	return GetEmissionState().getEndTimestamp()
}

func setEndTimestamp(timestamp int64) {
	GetEmissionState().setEndTimestamp(timestamp)
}

func GetHalvingInfo() string {
	height := std.GetHeight()
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

func isEmissionEnded(height int64) bool {
	return height > GetEndHeight()
}

// getHalvingYearAndEndTimestamp returns the halving year and end timestamp of the given timestamp
// if the timestamp is not in any halving year, it returns 0, 0
func getHalvingYearAndEndTimestamp(timestamp int64) (int64, int64) {
	state := GetEmissionState()

	endTimestamp := state.getEndTimestamp()
	startTimestamp := state.getStartTimestamp()

	if timestamp > endTimestamp { // after 12 years
		return 0, 0
	}

	timestamp -= startTimestamp

	year := timestamp / SECONDS_IN_YEAR
	year += 1 // since we subtract startTimestamp at line 215, we need to add 1 to get the correct year

	return year, startTimestamp + (SECONDS_IN_YEAR * year)
}
