package emission

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/gns"
)

const (
	LIQUIDITY_STAKER int = iota + 1
	DEVOPS
	COMMUNITY_POOL
	GOV_STAKER
)

var (
	// Stores the percentage (in basis points) for each distribution target
	// 1 basis point = 0.01%
	distributionBpsPct *avl.Tree

	distributedToStaker        int64 // can be cleared by staker contract
	distributedToDevOps        int64
	distributedToCommunityPool int64
	distributedToGovStaker     int64 // can be cleared by governance staker

	// Historical total distributions (never reset)
	accuDistributedToStaker        int64
	accuDistributedToDevOps        int64
	accuDistributedToCommunityPool int64
	accuDistributedToGovStaker     int64
)

func getDistributionBpsPct(key string) int64 {
	val, exists := distributionBpsPct.Get(key)
	if !exists {
		panic(addDetailToError(
			errInvalidEmissionTarget,
			ufmt.Sprintf("invalid target(%s)", key),
		))
	}
	return val.(int64)
}

// Initialize default distribution percentages:
// - Liquidity Stakers: 75%
// - DevOps: 20%
// - Community Pool: 5%
// - Governance Stakers: 0%
//
// ref: https://docs.gnoswap.io/gnoswap-token/emission
func init() {
	distributionBpsPct = avl.NewTree()
	distributionBpsPct.Set(strconv.Itoa(LIQUIDITY_STAKER), int64(7500))
	distributionBpsPct.Set(strconv.Itoa(DEVOPS), int64(2000))
	distributionBpsPct.Set(strconv.Itoa(COMMUNITY_POOL), int64(500))
	distributionBpsPct.Set(strconv.Itoa(GOV_STAKER), int64(0))
}

// ChangeDistributionPctByAdmin changes the distribution percentage for the given targets.
// Panics if following conditions are not met:
// - caller is not admin
// - invalid target
// - duplicate targets
// - sum of percentages is not 10000
// - swap is halted
func ChangeDistributionPctByAdmin(cur realm,
	target01 int, pct01 int64,
	target02 int, pct02 int64,
	target03 int, pct03 int64,
	target04 int, pct04 int64,
) {
	assertOnlyAdmin()
	targets := []int{target01, target02, target03, target04}
	for _, target := range targets {
		assertDistributionTarget(target)
	}
	assertNoDuplicateTargets(target01, target02, target03, target04)
	assertSumDistributionPct(pct01, pct02, pct03, pct04)
	assertOnlyNotHalted()

	changeDistributionPcts(
		cur,
		target01, pct01,
		target02, pct02,
		target03, pct03,
		target04, pct04,
	)
}

// ChangeDistributionPct changes the distribution percentage for the given targets.
// Panics if following conditions are not met:
// - caller is not governance
// - invalid target
// - duplicate targets
// - sum of percentages is not 10000
// - swap is halted
func ChangeDistributionPct(cur realm,
	target01 int, pct01 int64,
	target02 int, pct02 int64,
	target03 int, pct03 int64,
	target04 int, pct04 int64,
) {
	assertOnlyGovernance()
	targets := []int{target01, target02, target03, target04}
	for _, target := range targets {
		assertDistributionTarget(target)
	}
	assertNoDuplicateTargets(target01, target02, target03, target04)
	assertSumDistributionPct(pct01, pct02, pct03, pct04)
	assertOnlyNotHalted()

	changeDistributionPcts(
		cur,
		target01, pct01,
		target02, pct02,
		target03, pct03,
		target04, pct04,
	)
}

func changeDistributionPcts(
	cur realm,
	target01 int, pct01 int64,
	target02 int, pct02 int64,
	target03 int, pct03 int64,
	target04 int, pct04 int64,
) {
	// First, cache the percentage of the staker just before it changes Callback if needed
	// (check if the LIQUIDITY_STAKER was located between target01 and 04)
	setDistributionBpsPct(cur, target01, pct01)
	setDistributionBpsPct(cur, target02, pct02)
	setDistributionBpsPct(cur, target03, pct03)
	setDistributionBpsPct(cur, target04, pct04)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"ChangeDistributionPct",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"target01", targetToStr(target01),
		"pct01", formatInt(pct01),
		"target02", targetToStr(target02),
		"pct02", formatInt(pct02),
		"target03", targetToStr(target03),
		"pct03", formatInt(pct03),
		"target04", targetToStr(target04),
		"pct04", formatInt(pct04),
	)
}

// distributeToTarget splits an amount according to the configured percentages
// and sends tokens to each target address. Returns the total amount distributed.
func distributeToTarget(cur realm, amount int64) int64 {
	totalSent := int64(0)

	distributionBpsPct.Iterate("", "", func(targetStr string, iPct any) bool {
		targetInt, err := strconv.Atoi(targetStr)
		if err != nil {
			panic(addDetailToError(
				errInvalidEmissionTarget,
				ufmt.Sprintf("invalid target(%s)", targetStr),
			))
		}

		pct := iPct.(int64)
		distAmount := calculateAmount(amount, pct)
		totalSent += distAmount

		transferToTarget(cross, targetInt, distAmount)

		return false
	})

	leftAmount := amount - totalSent

	// guard for totalSent > amount case
	// this should never happen.
	if leftAmount < 0 {
		panic(addDetailToError(
			errInvalidEmissionTarget,
			ufmt.Sprintf("total distributed amount (%d) exceeds original amount (%d)", totalSent, amount),
		))
	}

	setLeftGNSAmount(leftAmount)

	return totalSent
}

// calculateAmount converts a basis point percentage to actual token amount
// bptPct is in basis points (1/100th of 1%)
// Example: 7500 basis points = 75%
func calculateAmount(amount, bptPct int64) int64 {
	if amount < 0 || bptPct < 0 || bptPct > 10000 {
		panic("invalid amount or bptPct")
	}

	// prevent multiplication overflow
	if amount > 9223372036854775807/10000 {
		panic("amount too large")
	}

	return amount * bptPct / 10000
}

// transferToTarget sends tokens to the appropriate address based on target type
// and updates both current and accumulated distribution tracking
func transferToTarget(cur realm, target int, amount int64) {
	switch target {
	case LIQUIDITY_STAKER:
		gns.Transfer(cross, stakerAddr, amount)
		distributedToStaker += amount
		accuDistributedToStaker += amount

	case DEVOPS:
		gns.Transfer(cross, devOpsAddr, amount)
		distributedToDevOps += amount
		accuDistributedToDevOps += amount

	case COMMUNITY_POOL:
		gns.Transfer(cross, communityPoolAddr, amount)
		distributedToCommunityPool += amount
		accuDistributedToCommunityPool += amount

	case GOV_STAKER:
		gns.Transfer(cross, govStakerAddr, amount)
		distributedToGovStaker += amount
		accuDistributedToGovStaker += amount

	default:
		panic(addDetailToError(
			errInvalidEmissionTarget,
			ufmt.Sprintf("invalid target(%d)", target),
		))
	}
}

func GetDistributionBpsPct(target int) int64 {
	assertDistributionTarget(target)
	if distributionBpsPct == nil {
		panic("distributionBpsPct is nil")
	}
	iInt64, exist := distributionBpsPct.Get(strconv.Itoa(target))
	if !exist {
		panic(addDetailToError(
			errInvalidEmissionTarget,
			ufmt.Sprintf("invalid target(%d)", target),
		))
	}

	return iInt64.(int64)
}

func GetDistributedToStaker() int64 {
	return distributedToStaker
}

func GetDistributedToDevOps() int64 {
	return distributedToDevOps
}

func GetDistributedToCommunityPool() int64 {
	return distributedToCommunityPool
}

func GetDistributedToGovStaker() int64 {
	return distributedToGovStaker
}

func GetAccuDistributedToStaker() int64 {
	return accuDistributedToStaker
}

func GetAccuDistributedToDevOps() int64 {
	return accuDistributedToDevOps
}

func GetAccuDistributedToCommunityPool() int64 {
	return accuDistributedToCommunityPool
}

func GetAccuDistributedToGovStaker() int64 {
	return accuDistributedToGovStaker
}

func ClearDistributedToStaker(cur realm) {
	assertStakerOnly()
	distributedToStaker = 0
}

func ClearDistributedToGovStaker(cur realm) {
	assertOnlyGovStaker()
	distributedToGovStaker = 0
}

// setDistributionBpsPct changes percentage of each target for how much gns it will get by emission.
func setDistributionBpsPct(cur realm, target int, pct int64) {
	if distributionBpsPct == nil {
		distributionBpsPct = avl.NewTree()
	}
	if target == LIQUIDITY_STAKER {
		oldPct, exist := distributionBpsPct.Get(strconv.Itoa(target))
		if !exist {
			panic("should not happen")
		}

		oldPctInt64, ok := oldPct.(int64)
		if !ok {
			panic("cannot convert oldPct to int64")
		}
		if oldPctInt64 != pct {
			changedPct := calculateAmount(gns.GetEmission(), pct)
			if callbackStakerEmissionChange != nil {
				callbackStakerEmissionChange(changedPct)
			}
		}
	}

	distributionBpsPct.Set(strconv.Itoa(target), pct)
}

func GetEmission() int64 {
	return calculateAmount(gns.GetEmission(), GetDistributionBpsPct(LIQUIDITY_STAKER))
}

// When there is a staker % change, it will trigger the StakerCallback first, which should "flush" the halving blocks first(by poolTier calling the GetHalvingBlocksInRange before it updates the emission).
// Therefore, if there are halving blocks exist from what we got from gns side, there must be no % change event between those blocks, which means we can use the current distribution pct to calculate the emissions for the past halving blocks.
func GetHalvingBlocksInRange(start, end int64) ([]int64, []int64) {
	halvingBlocks, halvingEmissions := gns.GetHalvingBlocksInRange(start, end)
	for i := range halvingBlocks {
		// Applying staker ratio for past halving blocks
		halvingEmissions[i] = calculateAmount(halvingEmissions[i], GetDistributionBpsPct(LIQUIDITY_STAKER))
	}
	return halvingBlocks, halvingEmissions
}

func targetToStr(target int) string {
	switch target {
	case LIQUIDITY_STAKER:
		return "LIQUIDITY_STAKER"
	case DEVOPS:
		return "DEVOPS"
	case COMMUNITY_POOL:
		return "COMMUNITY_POOL"
	case GOV_STAKER:
		return "GOV_STAKER"
	default:
		return "UNKNOWN"
	}
}
