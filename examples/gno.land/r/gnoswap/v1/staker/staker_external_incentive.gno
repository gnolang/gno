package staker

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/r/gnoswap/v1/gns"

	en "gno.land/r/gnoswap/v1/emission"
	pl "gno.land/r/gnoswap/v1/pool"
)

// CreateExternalIncentive creates an incentive program for a pool.
// ref: https://docs.gnoswap.io/contracts/staker/staker.gno#createexternalincentive
func CreateExternalIncentive(
	targetPoolPath string,
	rewardToken string, // token path should be registered
	rewardAmount uint64,
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

	en.MintAndDistributeGns()

	// check token can be used as reward
	if err := isAllowedForExternalReward(targetPoolPath, rewardToken); err != nil {
		panic(err.Error())
	}

	// native ugnot check
	if rewardToken == consts.GNOT {
		sent := std.OriginSend()
		ugnotSent := uint64(sent.AmountOf("ugnot"))

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
			ufmt.Sprintf("staker.gno__CreateExternalIncentive() || endTimestamp(%d) cannot be later than 253402300799 (9999-12-31 23:59:59)", endTimestamp),
		))
	}

	caller := getPrevAddr()

	// incentiveId := incentiveIdCompute(std.PrevRealm().Address(), targetPoolPath, rewardToken, startTimestamp, endTimestamp, std.GetHeight())
	incentiveId := incentiveIdByTime(startTimestamp, endTimestamp, caller, rewardToken)

	externalDuration := uint64(endTimestamp - startTimestamp)
	if err := isValidIncentiveDuration(externalDuration); err != nil {
		panic(err.Error())
	}

	pool := pools.GetOrCreate(targetPoolPath)

	incentive := NewExternalIncentive(
		incentiveId,
		targetPoolPath,
		rewardToken,
		rewardAmount,
		startTimestamp,
		endTimestamp,
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
	gns.TransferFrom(caller, consts.STAKER_ADDR, depositGnsAmount)

	// transfer reward token from user to staker
	teller := common.GetTokenTeller(rewardToken)
	teller.TransferFrom(caller, consts.STAKER_ADDR, rewardAmount)

	pool.incentives.create(caller, incentive)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"CreateExternalIncentive",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"incentiveId", incentiveId,
		"targetPoolPath", targetPoolPath,
		"rewardToken", rewardToken,
		"rewardAmount", formatUint(rewardAmount),
		"startTimestamp", formatInt(startTimestamp),
		"endTimestamp", formatInt(endTimestamp),
		"depositGnsAmount", formatUint(depositGnsAmount),
		"currentHeight", formatInt(std.ChainHeight()),
		"currentTime", formatInt(time.Now().Unix()),
		"avgBlockTimeInMs", formatInt(gns.GetAvgBlockTimeInMs()),
	)
}

// EndExternalIncentive ends the external incentive and refunds the remaining reward
// ref: https://docs.gnoswap.io/contracts/staker/staker.gno#endexternalincentive
func EndExternalIncentive(refundee std.Address, targetPoolPath, rewardToken string, startTimestamp, endTimestamp, height int64) {
	assertOnlyNotHalted()

	pool, exists := pools.Get(targetPoolPath)
	if !exists {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("staker.gno__EndExternalIncentive() || targetPoolPath(%s) does not exist", targetPoolPath),
		))
	}

	ictv, exists := pool.incentives.Get(startTimestamp, endTimestamp, refundee, rewardToken)
	if !exists {
		panic(addDetailToError(
			errCannotEndIncentive,
			ufmt.Sprintf("cannot end non existent incentive(%d:%d:%s:%s)", startTimestamp, endTimestamp, refundee, rewardToken),
		))
	}

	currentHeight := std.ChainHeight()
	if currentHeight < ictv.endHeight {
		panic(addDetailToError(
			errCannotEndIncentive,
			ufmt.Sprintf("cannot end incentive before endHeight(%d), current(%d)", ictv.endHeight, currentHeight),
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
	refund := ictv.rewardLeft

	if !ictv.unclaimableRefunded {
		refund += pool.incentives.calculateUnclaimableReward(ictv.incentiveId)
		ictv.unclaimableRefunded = true
	}

	poolLeftExternalRewardAmount := common.BalanceOf(ictv.rewardToken, consts.STAKER_ADDR)
	if poolLeftExternalRewardAmount < refund {
		refund = poolLeftExternalRewardAmount
	}

	teller := common.GetTokenTeller(ictv.rewardToken)
	teller.Transfer(ictv.refundee, refund)

	// unwrap if wugnot
	if ictv.rewardToken == consts.WUGNOT_PATH {
		unwrap(refund)
	}

	// also refund deposit gns amount
	gns.Transfer(ictv.refundee, ictv.depositGnsAmount)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"EndExternalIncentive",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"incentiveId", ictv.incentiveId,
		"targetPoolPath", targetPoolPath,
		"refundee", refundee.String(),
		"refundToken", rewardToken,
		"refundAmount", formatUint(refund),
		"refundGnsAmount", formatUint(ictv.depositGnsAmount),
		"isRequestUnwrap", formatBool(ictv.rewardToken == consts.WUGNOT_PATH),
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

func gnsBalance(addr std.Address) uint64 {
	return gns.BalanceOf(addr)
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
