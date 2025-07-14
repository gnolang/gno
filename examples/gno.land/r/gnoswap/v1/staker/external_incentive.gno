package staker

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"
	"gno.land/r/gnoswap/v1/gns"
	pl "gno.land/r/gnoswap/v1/pool"
)

// CreateExternalIncentive creates an incentive program for a pool.
// ref: https://docs.gnoswap.io/contracts/staker/staker.gno#createexternalincentive
func CreateExternalIncentive(cur realm,
	targetPoolPath string,
	rewardToken string, // token path should be registered
	rewardAmount int64,
	startTimestamp int64,
	endTimestamp int64,
) {
	assertOnlyNotHalted()
	if common.GetLimitCaller() {
		prev := std.PreviousRealm()
		if err := common.UserOnly(prev); err != nil {
			panic(ufmt.Sprintf("%v: %v", errNoPermission, err))
		}
	}

	// panic if pool does not exist
	if !pl.DoesPoolPathExist(targetPoolPath) {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("targetPoolPath(%s) does not exist", targetPoolPath),
		))
	}

	// Check if reward amount meets the minimum requirement for the specific token
	minReward := GetMinimumRewardAmountForToken(rewardToken)
	if rewardAmount < int64(minReward) {
		panic(addDetailToError(
			errInvalidInput,
			ufmt.Sprintf("rewardAmount(%d) is less than minimum required amount(%d)", rewardAmount, minReward),
		))
	}

	en.MintAndDistributeGns(cross)

	// check token can be used as reward
	if err := isAllowedForExternalReward(targetPoolPath, rewardToken); err != nil {
		panic(err.Error())
	}

	// native ugnot check
	if rewardToken == consts.GNOT {
		sent := std.OriginSend()
		ugnotSent := sent.AmountOf("ugnot")

		if ugnotSent != rewardAmount {
			panic(addDetailToError(
				errInvalidInput,
				ufmt.Sprintf("user(%s) sent ugnot(%d) amount not equal to rewardAmount(%d)", getPrevAddr(), ugnotSent, rewardAmount),
			))
		}

		wrap(ugnotSent)

		rewardToken = consts.WUGNOT_PATH
	}

	// must be in seconds format, not milliseconds
	// must be at least +1 day midnight
	// must be midnight of the day
	if err := checkStartTime(startTimestamp); err != nil {
		panic(err.Error())
	}

	// endTimestamp cannot be later than 253402300799 (9999-12-31 23:59:59)
	if endTimestamp >= MAX_UNIX_EPOCH_TIME {
		panic(addDetailToError(
			errInvalidIncentiveEndTime,
			ufmt.Sprintf("endTimestamp(%d) cannot be later than 253402300799 (9999-12-31 23:59:59)", endTimestamp),
		))
	}

	caller := getPrevAddr()

	// incentiveId := incentiveIdCompute(std.PrevRealm().Address(), targetPoolPath, rewardToken, startTimestamp, endTimestamp, std.GetHeight())
	incentiveId := incentiveIdByTime(startTimestamp, endTimestamp, caller, rewardToken, targetPoolPath)

	externalDuration := uint64(endTimestamp - startTimestamp)
	if err := isValidIncentiveDuration(externalDuration); err != nil {
		panic(err)
	}

	pool := pools.GetOrCreate(targetPoolPath)

	incentive := NewExternalIncentive(
		incentiveId,
		targetPoolPath,
		rewardToken,
		rewardAmount,
		int64(startTimestamp),
		int64(endTimestamp),
		caller,
		std.ChainHeight(),
		depositGnsAmount,
		time.Now().Unix(),
		gns.GetAvgBlockTimeInMs(),
	)

	if externalIncentives.Has(incentiveId) {
		panic(addDetailToError(
			errIncentiveAlreadyExists,
			ufmt.Sprintf("incentiveId(%s)", incentiveId),
		))
	}
	// store external incentive information for each incentiveId
	externalIncentives.set(incentiveId, incentive)

	// deposit gns amount
	gns.TransferFrom(cross, caller, stakerAddr, depositGnsAmount)

	// transfer reward token from user to staker
	common.TransferFrom(cross, rewardToken, caller, stakerAddr, rewardAmount)

	pool.incentives.create(caller, incentive)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"CreateExternalIncentive",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"incentiveId", incentiveId,
		"targetPoolPath", targetPoolPath,
		"rewardToken", rewardToken,
		"rewardAmount", formatInt(rewardAmount),
		"startTimestamp", formatInt(startTimestamp),
		"endTimestamp", formatInt(endTimestamp),
		"depositGnsAmount", formatInt(depositGnsAmount),
		"currentHeight", formatInt(std.ChainHeight()),
		"currentTime", formatInt(time.Now().Unix()),
		"avgBlockTimeInMs", formatInt(gns.GetAvgBlockTimeInMs()),
	)
}

// EndExternalIncentive ends the external incentive and refunds the remaining reward
// ref: https://docs.gnoswap.io/contracts/staker/staker.gno#endexternalincentive
func EndExternalIncentive(cur realm, refundee std.Address, targetPoolPath, rewardToken string, startTimestamp, endTimestamp, height int64) {
	assertOnlyNotHalted()
	pool, exists := pools.Get(targetPoolPath)
	if !exists {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("targetPoolPath(%s) does not exist", targetPoolPath),
		))
	}

	incentiveId := incentiveIdByTime(startTimestamp, endTimestamp, refundee, rewardToken, targetPoolPath)
	readOnlyIncentive, exists := pool.incentives.Get(startTimestamp, endTimestamp, refundee, rewardToken)
	incentive := readOnlyIncentive.Clone()
	if !exists {
		panic(addDetailToError(
			errCannotEndIncentive,
			ufmt.Sprintf("cannot end non existent incentive(%d:%d:%s:%s)", startTimestamp, endTimestamp, refundee, rewardToken),
		))
	}

	currentHeight := std.ChainHeight()
	if currentHeight < incentive.endHeight {
		panic(addDetailToError(
			errCannotEndIncentive,
			ufmt.Sprintf("cannot end incentive before endHeight(%d), current(%d)", incentive.endHeight, currentHeight),
		))
	}

	// when incentive end time is over
	// admin or refundee can end incentive ( left amount will be refunded )
	caller := getPrevAddr()
	if caller != consts.ADMIN && caller != refundee {
		panic(addDetailToError(
			errNoPermission,
			ufmt.Sprintf("only refundee(%s) or admin(%s) can end incentive, but called from %s", refundee, consts.ADMIN, caller),
		))
	}

	// when incentive ended, refund remaining reward
	refund := int64(incentive.rewardLeft)

	if !incentive.unclaimableRefunded {
		refund += int64(pool.incentives.calculateUnclaimableReward(incentive.incentiveId))
		incentive.unclaimableRefunded = true
	}

	poolLeftExternalRewardAmount := common.BalanceOf(incentive.rewardToken, stakerAddr)
	if poolLeftExternalRewardAmount < refund {
		refund = poolLeftExternalRewardAmount
	}

	common.Transfer(cross, incentive.rewardToken, incentive.refundee, refund)

	// unwrap if wugnot
	if incentive.rewardToken == consts.WUGNOT_PATH {
		unwrap(refund)
	}

	// also refund deposit gns amount
	gns.Transfer(cross, incentive.refundee, int64(incentive.depositGnsAmount))

	pool.incentives.update(incentive.refundee, incentive)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"EndExternalIncentive",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"incentiveId", incentiveId,
		"targetPoolPath", targetPoolPath,
		"refundee", refundee.String(),
		"refundToken", rewardToken,
		"refundAmount", formatInt(refund),
		"refundGnsAmount", formatInt(incentive.depositGnsAmount),
		"isRequestUnwrap", formatBool(incentive.rewardToken == consts.WUGNOT_PATH),
		"externalIncentiveEndBy", prevAddr,
	)
}

func isValidIncentiveDuration(dur uint64) error {
	switch dur {
	case TIMESTAMP_90DAYS, TIMESTAMP_180DAYS, TIMESTAMP_365DAYS:
		return nil
	}

	return ufmt.Errorf(
		"%v: externalDuration(%d) must be 90, 180, 365 days",
		errInvalidIncentiveDuration, dur,
	)
}

// checkStartTime checks whether the current time meets the conditions for generating an
// external rewards. Since the earliest time this reward can be generated is from
// midnight of the next day, it uses a timestamp to verify this timing.
func checkStartTime(startTimestamp int64) error {
	// must be in seconds format, not milliseconds
	// REF: https://stackoverflow.com/a/23982005
	numStr := strconv.Itoa(int(startTimestamp))
	if len(numStr) >= 13 {
		return ufmt.Errorf(
			"%v: startTimestamp(%d) must be in seconds format, not milliseconds",
			errInvalidIncentiveStartTime, startTimestamp,
		)
	}

	// must be at least +1 day midnight
	tomorrowMidnight := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour).Unix()
	if startTimestamp < tomorrowMidnight {
		return ufmt.Errorf(
			"%v: startTimestamp(%d) must be at least +1 day midnight(%d)",
			errInvalidIncentiveStartTime, startTimestamp, tomorrowMidnight,
		)
	}

	// must be midnight of the day
	startTime := time.Unix(startTimestamp, 0)
	if !isMidnight(startTime) {
		return ufmt.Errorf(
			"%v: startTime(%d = %s) must be midnight of the day",
			errInvalidIncentiveStartTime, startTimestamp, startTime.String(),
		)
	}

	return nil
}

func isMidnight(startTime time.Time) bool {
	hour := startTime.Hour()
	minute := startTime.Minute()
	second := startTime.Second()

	return hour == 0 && minute == 0 && second == 0
}

func isAllowedForExternalReward(poolPath, tokenPath string) error {
	token0, token1, _ := poolPathDivide(poolPath)

	if tokenPath == token0 || tokenPath == token1 {
		return nil
	}

	allowed := contains(allowedTokens, tokenPath)
	if allowed {
		return nil
	}

	return ufmt.Errorf(
		"%v: tokenPath(%s) is not allowed for external reward for poolPath(%s)",
		errNotAllowedForExternalReward, tokenPath, poolPath,
	)
}
