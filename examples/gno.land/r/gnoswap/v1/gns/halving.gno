package gns

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
)

const (
	DAY_PER_YEAR    = 365
	SECONDS_IN_YEAR = 31536000

	HALVING_START_YEAR = int64(1)
	HALVING_END_YEAR   = int64(12)
)

// Annual halving amount - maximum issuance per year
var halvingAmountsPerYear = [HALVING_END_YEAR]int64{
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

var (
	blockPerYear = SECONDS_IN_YEAR / milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockPerDay  = consts.SECONDS_PER_DAY / milliToSec(consts.BLOCK_GENERATION_INTERVAL)

	avgBlockTimeMs int64 = consts.BLOCK_GENERATION_INTERVAL
	perBlockMint         = avl.NewTree() // height => uint64
)

// HalvingData is a struct that stores data for each halving period
type HalvingData struct {
	startBlockHeight []int64 // starting block height
	endBlockHeight   []int64 // ending block height
	startTimestamp   []int64 // starting timestamp
	maxAmount        []int64 // maximum issuance amount
	mintedAmount     []int64 // amount already minted
	leftAmount       []int64 // remaining amount to be minted
	accumAmount      []int64 // accumulated amount
	amountPerBlock   []int64 // amount per block
}

// Getter/setter methods for each field
// Each slice is 0-indexed but year is 1-indexed, so index adjustment is needed
func (h *HalvingData) getStartBlockHeight(year int64) int64 {
	if year <= 0 {
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

func (h *HalvingData) getMaxAmount(year int64) int64 {
	if year == 0 {
		return 0
	}
	return h.maxAmount[year-1]
}

func (h *HalvingData) setMaxAmount(year, amount int64) {
	assertValidYear(year)
	h.maxAmount[year-1] = amount
}

func (h *HalvingData) getMintedAmount(year int64) int64 {
	if year == 0 {
		return 0
	}
	return h.mintedAmount[year-1]
}

func (h *HalvingData) setMintedAmount(year, amount int64) {
	assertValidYear(year)
	h.mintedAmount[year-1] = amount
}

func (h *HalvingData) getLeftAmount(year int64) int64 {
	if year == 0 {
		return 0
	}
	return h.leftAmount[year-1]
}

func (h *HalvingData) setLeftAmount(year, amount int64) {
	assertValidYear(year)
	h.leftAmount[year-1] = amount
}

func (h *HalvingData) getAccumAmount(year int64) int64 {
	if year == 0 {
		return 0
	}
	return h.accumAmount[year-1]
}

func (h *HalvingData) setAccumAmount(year, amount int64) {
	assertValidYear(year)
	h.accumAmount[year-1] = amount
}

func (h *HalvingData) addAccumAmount(year, amount int64) {
	assertValidYear(year)
	h.accumAmount[year-1] += amount
}

func (h *HalvingData) getAmountPerBlock(year int64) int64 {
	if year == 0 {
		return 0
	}
	return h.amountPerBlock[year-1]
}

func (h *HalvingData) setAmountPerBlock(year, amount int64) {
	assertValidYear(year)
	h.amountPerBlock[year-1] = amount
}

// initialize halving data according to emission schedule
// Reference: https://docs.gnoswap.io/gnoswap-token/release-schedule
func (h *HalvingData) initialize(startHeight, startTimestamp int64) {
	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		// Set maximum issuance amount for the year
		currentYearMaxAmount := GetHalvingAmountsPerYear(year)
		h.setMaxAmount(year, currentYearMaxAmount)

		// For the first year, directly set start height and timestamp
		if year == HALVING_START_YEAR {
			h.setAccumAmount(year, currentYearMaxAmount)
			h.setStartBlockHeight(year, startHeight)
			h.setEndBlockHeight(year, startHeight+(blockPerYear)-1)
		} else {
			// Add current year's max amount to previous year's accumulated amount
			h.setAccumAmount(year, currentYearMaxAmount+h.getAccumAmount(year-1))

			// Current year's start block is previous year's end block + 1
			h.setStartBlockHeight(year, h.getEndBlockHeight(year-1)+1)

			// Current year's end block is start block + blocks per year - 1
			h.setEndBlockHeight(year, h.getStartBlockHeight(year)+blockPerYear-1)
		}

		// Set start timestamp
		h.setStartTimestamp(year, startTimestamp+(SECONDS_IN_YEAR*(year-1)))

		// Calculate and set amount per block
		amountPerDay := currentYearMaxAmount / DAY_PER_YEAR
		amountPerBlock := amountPerDay / blockPerDay
		h.setAmountPerBlock(year, amountPerBlock)

		// Set initial minted amount and remaining amount
		h.setMintedAmount(year, 0)
		h.setLeftAmount(year, currentYearMaxAmount)
	}
}

// Emission change callback function
var callbackEmissionChange func(amount int64)

func SetCallbackEmissionChange(cur realm, callback func(amount int64)) {
	assertCallerIsEmission()

	callbackEmissionChange = callback
}

// EmissionState is a struct for managing emission state
type EmissionState struct {
	startHeight    int64       // emission start block height
	startTimestamp int64       // emission start timestamp
	endTimestamp   int64       // emission end timestamp
	halvingData    HalvingData // halving data
}

// Global variable for singleton pattern
var emissionState *EmissionState

// GetEmissionState returns emission state (singleton pattern)
func GetEmissionState() *EmissionState {
	if emissionState == nil {
		emissionState = NewEmissionState()
		emissionState.initializeHalvingData()
	}
	return emissionState
}

// NewEmissionState creates a new emission state
func NewEmissionState() *EmissionState {
	now := time.Now().Unix() + milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	emissionEndTime := now + SECONDS_IN_YEAR*HALVING_END_YEAR

	return &EmissionState{
		startHeight:    std.ChainHeight() + 1,
		startTimestamp: now,
		endTimestamp:   emissionEndTime,
		halvingData: HalvingData{
			startBlockHeight: make([]int64, HALVING_END_YEAR),
			endBlockHeight:   make([]int64, HALVING_END_YEAR),
			startTimestamp:   make([]int64, HALVING_END_YEAR),
			maxAmount:        make([]int64, HALVING_END_YEAR),
			mintedAmount:     make([]int64, HALVING_END_YEAR),
			leftAmount:       make([]int64, HALVING_END_YEAR),
			accumAmount:      make([]int64, HALVING_END_YEAR),
			amountPerBlock:   make([]int64, HALVING_END_YEAR),
		},
	}
}

func (e *EmissionState) getStartHeight() int64             { return e.startHeight }
func (e *EmissionState) setStartHeight(height int64)       { e.startHeight = height }
func (e *EmissionState) getStartTimestamp() int64          { return e.startTimestamp }
func (e *EmissionState) setStartTimestamp(timestamp int64) { e.startTimestamp = timestamp }
func (e *EmissionState) getEndTimestamp() int64            { return e.endTimestamp }
func (e *EmissionState) setEndTimestamp(timestamp int64)   { e.endTimestamp = timestamp }
func (e *EmissionState) getHalvingData() HalvingData       { return e.halvingData }
func (e *EmissionState) setHalvingData(data HalvingData)   { e.halvingData = data }

// initializeHalvingData initializes halving data
func (e *EmissionState) initializeHalvingData() {
	halvingData := e.getHalvingData()
	halvingData.initialize(e.getStartHeight(), e.getStartTimestamp())
	e.setHalvingData(halvingData)
}

// init package initialization function
func init() {
	emissionState = GetEmissionState()
}

// GetAvgBlockTimeInMs returns average block time in milliseconds
func GetAvgBlockTimeInMs() int64 {
	return avgBlockTimeMs
}

// SetAvgBlockTimeInMsByAdmin admin sets average block time
func SetAvgBlockTimeInMsByAdmin(cur realm, ms int64) {
	if ms <= 0 {
		panic(errInvalidAvgBlockTimeInMs)
	}
	assertCallerIsAdmin()

	prevAvgBlockTimeInMs := GetAvgBlockTimeInMs()
	setAvgBlockTimeInMs(ms)

	// Log event generation
	halvingData := GetEmissionState().getHalvingData()
	prevAddr, prevPkgPath := getPrev()
	emitBlockTimeChangeEvent("SetAvgBlockTimeInMsByAdmin", prevAddr, prevPkgPath,
		prevAvgBlockTimeInMs, ms, halvingData)
}

// SetAvgBlockTimeInMs sets average block time through governance proposal
func SetAvgBlockTimeInMs(cur realm, ms int64) {
	assertCallerIsGovernance()
	prevAvgBlockTimeInMs := GetAvgBlockTimeInMs()
	setAvgBlockTimeInMs(ms)

	// Log event generation
	halvingData := GetEmissionState().getHalvingData()
	prevAddr, prevPkgPath := getPrev()
	emitBlockTimeChangeEvent("SetAvgBlockTimeInMs", prevAddr, prevPkgPath,
		prevAvgBlockTimeInMs, ms, halvingData)
}

// Helper function to emit event to remove duplicate code
func emitBlockTimeChangeEvent(eventName, prevAddr, prevPkgPath string, prevTime, newTime int64, halvingData HalvingData) {
	eventParams := []string{
		eventName,
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevAvgBlockTimeMs", formatInt(prevTime),
		"newAvgBlockTimeMs", formatInt(newTime),
	}

	// add end block heights for all years
	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		eventParams = append(eventParams,
			ufmt.Sprintf("endBlock%dYear", year),
			formatInt(halvingData.getEndBlockHeight(year)))
	}
}

// setAvgBlockTimeInMs sets the new average block time (ms) and updates all halving schedules.
func setAvgBlockTimeInMs(ms int64) {
	assertShouldNotBeHalted()

	// prevent division errors due to very large values or zero.
	// block time values outside of practical range are blocked
	if ms <= 0 || ms >= 1e9 {
		panic("invalid block time interval")
	}

	now := time.Now().Unix()

	// sync blockPerYear, blockPerDay
	// SECONDS_IN_YEAR, consts.SECONDS_PER_DAY must be in seconds
	blockPerYear = secToMs(SECONDS_IN_YEAR) / ms
	blockPerDay = secToMs(consts.SECONDS_PER_DAY) / ms

	// recalculate all remaining years (amountPerBlock)
	es := GetEmissionState()

	currentHeight := std.ChainHeight()
	currentYear := GetHalvingYearByHeight(currentHeight)
	prevBlockTimeMs := GetAvgBlockTimeInMs()

	hd := &es.halvingData
	for year := currentYear; year <= HALVING_END_YEAR; year++ {
		maxMintedAmount := hd.getMaxAmount(year)
		halvingEndHeight := hd.getEndBlockHeight(year)

		// Calculate remaining blocks with new block time
		remainBlockNum := halvingEndHeight - currentHeight
		changedRemainBlockNum := remainBlockNum * prevBlockTimeMs / ms

		// For current year, use remaining amount to avoid over/under emission
		if year == currentYear {
			startHeight := hd.getStartBlockHeight(year)
			lastGNSMintedHeight := GetLastMintedHeight() + 1
			if lastGNSMintedHeight < startHeight {
				lastGNSMintedHeight = startHeight
			}

			mintedAmount := hd.getMintedAmount(year)
			leftAmount := maxMintedAmount - mintedAmount

			if changedRemainBlockNum > 0 {
				changedAmountPerBlock := leftAmount / changedRemainBlockNum
				hd.setAmountPerBlock(year, changedAmountPerBlock)
			}
		} else {
			// For future years, use total year amount
			dailyAmount := maxMintedAmount / DAY_PER_YEAR
			changedAmountPerBlock := dailyAmount / blockPerDay
			hd.setAmountPerBlock(year, changedAmountPerBlock)
		}

		changedEndHeight := currentHeight + changedRemainBlockNum
		hd.setEndBlockHeight(year, changedEndHeight)

		// next year's start block height is updated by calculated end block heightAdd commentMore actions
		if year < HALVING_END_YEAR {
			hd.setStartBlockHeight(year+1, changedEndHeight+1)
		}
	}

	// adjust the current year's amount per block
	currentYear, _ = getHalvingYearAndEndTimestamp(now)
	adjusted := hd.getAmountPerBlock(currentYear)

	if callbackEmissionChange == nil {
		panic(errCallbackEmissionChangeIsNil.Error())
	}

	callbackEmissionChange(adjusted)

	// final update of global variables
	avgBlockTimeMs = ms
}

// GetAmountByHeight returns emission amount based on block height
func GetAmountByHeight(height int64) int64 {
	if isEmissionEnded(height) {
		return 0
	}

	halvingYear := GetHalvingYearByHeight(height)
	return GetAmountPerBlockPerHalvingYear(halvingYear)
}

// GetHalvingYearByHeight returns halving year based on block height
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
