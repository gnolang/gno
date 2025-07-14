package launchpad

import (
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"
)

const (
	projectTier30  = int64(30)
	projectTier90  = int64(90)
	projectTier180 = int64(180)

	dayTime = int64(24 * 60 * 60) // 86400

	minimumDepositAmount = int64(1_000_000)

	stringSplitterPad = "*PAD*"

	projectMinimumStartDelayTime = int64(60 * 60) // 1 hour
)

var projectTierDurations = []int64{
	projectTier30,
	projectTier90,
	projectTier180,
}

var projectTierDurationTimes = map[int64]int64{
	projectTier30:  dayTime * projectTier30,  // 30 days
	projectTier90:  dayTime * projectTier90,  // 90 days
	projectTier180: dayTime * projectTier180, // 180 days
}

var projectTierRewardCollectableDuration = map[int64]int64{
	projectTier30:  dayTime * 3,  // 3 days
	projectTier90:  dayTime * 7,  // 7 days
	projectTier180: dayTime * 14, // 14 days
}

var q128 = u256.MustFromDecimal(consts.Q128)
