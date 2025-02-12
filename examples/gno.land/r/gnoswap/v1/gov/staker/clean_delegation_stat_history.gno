package staker

import (
	"std"

	"gno.land/p/gnoswap/consts"
)

// default one day
var thresholdVotingWeightBlockHeight = consts.SECONDS_PER_DAY / milliToSec(consts.BLOCK_GENERATION_INTERVAL)

var (
	lastCleanedHeight uint64 = 0
	running           bool   = true
)

func CleanDelegationStatHistoryByAdmin() {
	assertCallerIsAdmin()
	cleanDelegationStatHistory()
}

func GetRunning() bool {
	return running
}

func SetRunning(run bool) {
	assertCallerIsAdmin()

	running = run
}

func GetThresholdVotingWeightBlockHeight() uint64 {
	return uint64(thresholdVotingWeightBlockHeight)
}

func SetThresholdVotingWeightBlockHeightByAdmin(height uint64) {
	assertCallerIsAdmin()

	thresholdVotingWeightBlockHeight = int64(height)
}

func cleanDelegationStatHistory() {
	height := uint64(std.GetHeight())
	sinceLast := height - lastCleanedHeight

	if sinceLast < uint64(thresholdVotingWeightBlockHeight) {
		return
	}

	lastCleanedHeight = height

	// delete history older than 1 day, but keep the latest one
	keepFrom := height - uint64(thresholdVotingWeightBlockHeight)

	delegationSnapShotHistory.Iterate("", "", func(key string, value interface{}) bool {
		history := value.([]DelegationSnapShotHistory)

		// reverse history
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].updatedBlock > keepFrom {
				continue
			}

			// save truncated history
			newHistory := history[i:]
			delegationSnapShotHistory.Set(key, newHistory)
			break
		}

		return false
	})
}
