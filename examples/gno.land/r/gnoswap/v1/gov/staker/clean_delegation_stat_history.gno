package staker

import (
	"std"

	"gno.land/p/gnoswap/consts"
)

var (
	// default one day
	thresholdVotingWeightBlockHeight        = consts.SECONDS_PER_DAY / milliToSec(consts.BLOCK_GENERATION_INTERVAL)
	running                          bool   = true
	lastCleanedHeight                uint64 = 0
)

// CleanDelegationStatHistoryByAdmin cleans delegation stat history
func CleanDelegationStatHistoryByAdmin() {
	assertCallerIsAdmin()
	cleanDelegationStatHistory()
}

// GetThresholdVotingWeightBlockHeight returns the threshold voting weight block height
func GetThresholdVotingWeightBlockHeight() uint64 {
	return uint64(thresholdVotingWeightBlockHeight)
}

// SetThresholdVotingWeightBlockHeightByAdmin sets the threshold voting weight block height
func SetThresholdVotingWeightBlockHeightByAdmin(height uint64) {
	assertCallerIsAdmin()

	thresholdVotingWeightBlockHeight = int64(height)
}

// GetRunning returns the running status
func GetRunning() bool {
	return running
}

// SetRunning sets the running status
func SetRunning(run bool) {
	assertCallerIsAdmin()

	running = run
}

// cleanDelegationStatHistory cleans delegation stat history
func cleanDelegationStatHistory() {
	height := uint64(std.ChainHeight())
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
