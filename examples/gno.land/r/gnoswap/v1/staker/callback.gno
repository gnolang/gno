package staker

import (
	"std"
)

// callbackStakerEmissionChange is called by emission when
// - msPerBlock is changed
// - staker emission % is changed
// it does NOT get called in regards of halving(manually handled in cacheReward).
//
// Initially, it is passed to emission in staker.gno#init().
//
// For the parameter `emission`, which is a per-block emission for staker contract,
//   - It first caches the reward until the current block height.
//   - Then, it updates the `currentEmission` of the poolTier,
//     which will be applied for future blocks thereafter.
func callbackStakerEmissionChange(emission int64) {
	poolTier.currentEmission = int64(emission)
	poolTier.cacheReward(std.ChainHeight(), pools)
}
