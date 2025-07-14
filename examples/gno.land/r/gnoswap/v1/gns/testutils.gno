package gns

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/gnoswap/consts"
)

func InitGnsTest(t *testing.T) {
	t.Helper()

	setLeftEmissionAmount(MAX_EMISSION_AMOUNT)
	setMintedEmissionAmount(0)
	setLastMintedHeight(std.ChainHeight())
	burnAmount = 0

	blockPerYear = SECONDS_IN_YEAR / milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	blockPerDay = consts.SECONDS_PER_DAY / milliToSec(consts.BLOCK_GENERATION_INTERVAL)

	avgBlockTimeMs = consts.BLOCK_GENERATION_INTERVAL
	perBlockMint = avl.NewTree() // height => uint64

	emissionState = NewEmissionState()
	emissionState.initializeHalvingData()
}
