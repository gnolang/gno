package gns

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
)

var (
	govRealm   = std.NewCodeRealm(consts.GOV_GOVERNANCE_PATH)
	adminRealm = std.NewUserRealm(consts.ADMIN)

	startHeight    int64 = std.GetHeight()
	startTimestamp int64 = time.Now().Unix() + milliToSec(consts.BLOCK_GENERATION_INTERVAL)
)

var FIRST_BLOCK_OF_YEAR = []int64{
	startHeight + (blockPerYear * 0),
	startHeight + (blockPerYear * 1) + 1,
	startHeight + (blockPerYear * 2) + 2,
	startHeight + (blockPerYear * 3) + 3,
	startHeight + (blockPerYear * 4) + 4,
	startHeight + (blockPerYear * 5) + 5,
	startHeight + (blockPerYear * 6) + 6,
	startHeight + (blockPerYear * 7) + 7,
	startHeight + (blockPerYear * 8) + 8,
	startHeight + (blockPerYear * 9) + 9,
	startHeight + (blockPerYear * 10) + 10,
	startHeight + (blockPerYear * 11) + 11,
}

var FIRST_TIMESTAMP_OF_YEAR = []int64{
	startTimestamp + SECONDS_IN_YEAR*0,
	startTimestamp + SECONDS_IN_YEAR*1,
	startTimestamp + SECONDS_IN_YEAR*2,
	startTimestamp + SECONDS_IN_YEAR*3,
	startTimestamp + SECONDS_IN_YEAR*4,
	startTimestamp + SECONDS_IN_YEAR*5,
	startTimestamp + SECONDS_IN_YEAR*6,
	startTimestamp + SECONDS_IN_YEAR*7,
	startTimestamp + SECONDS_IN_YEAR*8,
	startTimestamp + SECONDS_IN_YEAR*9,
	startTimestamp + SECONDS_IN_YEAR*10,
	startTimestamp + SECONDS_IN_YEAR*11,
}

var END_TIMESTAMP_OF_YEAR = []int64{
	startTimestamp + SECONDS_IN_YEAR*1,
	startTimestamp + SECONDS_IN_YEAR*2,
	startTimestamp + SECONDS_IN_YEAR*3,
	startTimestamp + SECONDS_IN_YEAR*4,
	startTimestamp + SECONDS_IN_YEAR*5,
	startTimestamp + SECONDS_IN_YEAR*6,
	startTimestamp + SECONDS_IN_YEAR*7,
	startTimestamp + SECONDS_IN_YEAR*8,
	startTimestamp + SECONDS_IN_YEAR*9,
	startTimestamp + SECONDS_IN_YEAR*10,
	startTimestamp + SECONDS_IN_YEAR*11,
	startTimestamp + SECONDS_IN_YEAR*12,
}

func TestGetAmountByHeight(t *testing.T) {
	for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
		firstBlockOfYear := FIRST_BLOCK_OF_YEAR[year-1]
		uassert.Equal(t, GetAmountPerBlockPerHalvingYear(year), GetAmountByHeight(firstBlockOfYear))
		if year == HALVING_START_YEAR {
			uassert.Equal(t, GetHalvingYearAccuAmount(year), GetHalvingYearMaxAmount(year))
		} else {
			uassert.Equal(t, GetHalvingYearAccuAmount(year), GetHalvingYearAccuAmount(year-1)+GetHalvingYearMaxAmount(year))
			uassert.Equal(t, int64(1), GetHalvingYearStartBlock(year)-GetHalvingYearEndBlock(year-1))
		}
		uassert.Equal(t, blockPerYear, (GetHalvingYearEndBlock(year) - GetHalvingYearStartBlock(year) + 1))

		if year == HALVING_START_YEAR {
			uassert.Equal(t, GetHalvingYearMaxAmount(year)+GetAmountPerBlockPerHalvingYear(year+1), calculateAmountToMint(GetHalvingYearStartBlock(year), GetHalvingYearStartBlock(year+1)))
		} else if year == HALVING_END_YEAR {
			uassert.Equal(t, GetHalvingYearMaxAmount(year)-GetAmountPerBlockPerHalvingYear(year), calculateAmountToMint(GetHalvingYearStartBlock(year)+1, GetHalvingYearEndBlock(year)+1))
		} else {
			uassert.Equal(t, GetHalvingYearMaxAmount(year)-GetAmountPerBlockPerHalvingYear(year)+GetAmountPerBlockPerHalvingYear(year+1), calculateAmountToMint(GetHalvingYearStartBlock(year)+1, GetHalvingYearStartBlock(year+1)))
		}
	}
}

func TestGetHalvingYearByHeight(t *testing.T) {
	t.Run("during halving years", func(t *testing.T) {
		for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
			firstBlockOfYear := FIRST_BLOCK_OF_YEAR[year-1]
			uassert.Equal(t, year, GetHalvingYearByHeight(firstBlockOfYear))
		}
	})

	t.Run("no year after 12 years", func(t *testing.T) {
		uassert.Equal(t, int64(0), GetHalvingYearByHeight(GetEndHeight()+1))
	})
}

func TestGetHalvingYearAndEndTimestamp(t *testing.T) {
	t.Run("bit of extra timestamp for each year", func(t *testing.T) {
		for year := HALVING_START_YEAR; year <= HALVING_END_YEAR; year++ {
			firstTimestampOfYear := FIRST_TIMESTAMP_OF_YEAR[year-1]
			gotYear, gotEndTimestamp := getHalvingYearAndEndTimestamp(firstTimestampOfYear + 5) // after 5s
			uassert.Equal(t, year, gotYear)
			uassert.Equal(t, gotEndTimestamp, END_TIMESTAMP_OF_YEAR[year-1])
		}
	})

	t.Run("after 12 years", func(t *testing.T) {
		year, endTimestamp := getHalvingYearAndEndTimestamp(GetEndTimestamp() + 1)
		uassert.Equal(t, int64(0), year)
		uassert.Equal(t, int64(0), endTimestamp)
	})
}

func TestHalvingYearStartBlock(t *testing.T) {
	setHalvingYearStartBlock(1, 100)
	uassert.Equal(t, GetHalvingYearStartBlock(1), int64(100))
}

func TestHalvingYearTimestamp(t *testing.T) {
	setHalvingYearTimestamp(2, 200)
	uassert.Equal(t, GetHalvingYearTimestamp(2), int64(200))
}

func TestHalvingYearMaxAmount(t *testing.T) {
	setHalvingYearMaxAmount(3, 300)
	uassert.Equal(t, GetHalvingYearMaxAmount(3), uint64(300))
}

func TestHalvingYearMintAmount(t *testing.T) {
	setHalvingYearMintAmount(4, 400)
	uassert.Equal(t, GetHalvingYearMintAmount(4), uint64(400))
}

func TestHalvingYearAccuAmount(t *testing.T) {
	setHalvingYearAccuAmount(5, 500)
	uassert.Equal(t, GetHalvingYearAccuAmount(5), uint64(500))
}

func TestAmountPerBlockPerHalvingYear(t *testing.T) {
	setAmountPerBlockPerHalvingYear(6, 600)
	uassert.Equal(t, GetAmountPerBlockPerHalvingYear(6), uint64(600))
}

func TestGetHalvingInfo(t *testing.T) {
	jsonStr, err := json.Unmarshal([]byte(GetHalvingInfo()))
	uassert.NoError(t, err)

	halving := jsonStr.MustKey("halvings").MustArray()
	uassert.Equal(t, len(halving), 12)
}

func TestSetAvgBlockTimeInMsByAdmin(t *testing.T) {
	t.Run("panic if caller is not admin", func(t *testing.T) {
		oldCallback := callbackEmissionChange
		callbackEmissionChange = func(amount uint64) {}
		uassert.PanicsWithMessage(t,
			"caller(g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) has no permission",
			func() {
				SetAvgBlockTimeInMsByAdmin(1)
			},
		)
		callbackEmissionChange = oldCallback
	})

	t.Run("success if caller is admin", func(t *testing.T) {
		oldCallback := callbackEmissionChange
		callbackEmissionChange = func(amount uint64) {}
		std.TestSkipHeights(1)
		std.TestSetRealm(std.NewUserRealm(consts.ADMIN))
		SetAvgBlockTimeInMsByAdmin(2)
		uassert.Equal(t, GetAvgBlockTimeInMs(), int64(2))
		callbackEmissionChange = oldCallback
	})
}

func TestSetAvgBlockTimeInMs(t *testing.T) {
	t.Run("panic if caller is not governance contract", func(t *testing.T) {
		oldCallback := callbackEmissionChange
		callbackEmissionChange = func(amount uint64) {}
		uassert.PanicsWithMessage(t,
			"caller(g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) has no permission",
			func() {
				SetAvgBlockTimeInMs(3)
			},
		)
		callbackEmissionChange = oldCallback
	})

	t.Run("success if caller is governance contract", func(t *testing.T) {
		oldCallback := callbackEmissionChange
		callbackEmissionChange = func(amount uint64) {}
		std.TestSkipHeights(3)
		std.TestSetRealm(govRealm)
		SetAvgBlockTimeInMs(4)
		uassert.Equal(t, GetAvgBlockTimeInMs(), int64(4))
		callbackEmissionChange = oldCallback
	})
}
