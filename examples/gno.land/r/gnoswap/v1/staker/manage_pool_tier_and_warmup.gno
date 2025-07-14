package staker

import (
	"std"

	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/halt"

	en "gno.land/r/gnoswap/v1/emission"
	pl "gno.land/r/gnoswap/v1/pool"
)

const (
	NOT_EMISSION_TARGET_TIER uint64 = 0
)

// SetPoolTierByAdmin sets the tier of a pool by an admin
func SetPoolTierByAdmin(cur realm, poolPath string, tier uint64) {
	assertOnlyAdmin()
	assertMustNotHalted()
	assertPoolMustExist(poolPath)

	setPoolTier(cur, poolPath, tier)
}

// SetPoolTier sets the tier of a pool by governance
func SetPoolTier(cur realm, poolPath string, tier uint64) {
	assertOnlyGovernance()
	assertMustNotHalted()
	assertPoolMustExist(poolPath)

	setPoolTier(cur, poolPath, tier)
}

func setPoolTier(cur realm, poolPath string, tier uint64) {
	en.MintAndDistributeGns(cross)

	currentHeight := std.ChainHeight()
	pools.GetOrCreate(poolPath)
	poolTier.changeTier(currentHeight, pools, poolPath, tier)

	prevAddr, prevRealm := getPrev()
	std.Emit(
		"SetPoolTier",
		"prevAddr", prevAddr,
		"prevRealm", prevRealm,
		"poolPath", poolPath,
		"tier", formatUint(tier),
		"height", formatInt(currentHeight),
	)
}

// ChangePoolTierByAdmin changes the tier of a pool by an admin
func ChangePoolTierByAdmin(cur realm, poolPath string, tier uint64) {
	assertOnlyAdmin()
	assertMustNotHalted()
	assertPoolMustExist(poolPath)

	changePoolTier(cur, poolPath, tier)
}

// ChangePoolTier changes the tier of a pool by governance
func ChangePoolTier(cur realm, poolPath string, tier uint64) {
	assertOnlyGovernance()
	assertMustNotHalted()
	assertPoolMustExist(poolPath)

	changePoolTier(cur, poolPath, tier)
}

func changePoolTier(cur realm, poolPath string, tier uint64) {
	en.MintAndDistributeGns(cross)

	prevTier := poolTier.CurrentTier(poolPath)
	currentHeight := std.ChainHeight()
	poolTier.changeTier(currentHeight, pools, poolPath, tier)

	prevAddr, prevRealm := getPrev()
	std.Emit(
		"ChangePoolTier",
		"prevAddr", prevAddr,
		"prevRealm", prevRealm,
		"poolPath", poolPath,
		"prevTier", formatUint(prevTier),
		"newTier", formatUint(tier),
		"height", formatInt(currentHeight),
	)
}

// RemovePoolTierByAdmin removes the tier of a pool by an admin
func RemovePoolTierByAdmin(cur realm, poolPath string) {
	assertOnlyAdmin()
	assertMustNotHalted()
	assertPoolMustExist(poolPath)

	removePoolTier(cur, poolPath)
}

// RemovePoolTier removes the tier of a pool by governance
func RemovePoolTier(cur realm, poolPath string) {
	assertOnlyGovernance()
	assertMustNotHalted()
	assertPoolMustExist(poolPath)

	removePoolTier(cur, poolPath)
}

func removePoolTier(cur realm, poolPath string) {
	en.MintAndDistributeGns(cross)

	currentHeight := std.ChainHeight()
	poolTier.changeTier(currentHeight, pools, poolPath, NOT_EMISSION_TARGET_TIER)

	prevAddr, prevRealm := getPrev()
	std.Emit(
		"RemovePoolTier",
		"prevAddr", prevAddr,
		"prevRealm", prevRealm,
		"poolPath", poolPath,
		"height", formatInt(currentHeight),
	)
}

// SetWarmUpByAdmin sets the warmup percentage and block duration by an admin
func SetWarmUpByAdmin(cur realm, pct, blockDuration int64) {
	assertOnlyAdmin()
	assertMustNotHalted()

	setWarmUp(cur, pct, blockDuration)
}

// SetWarmUp sets the warmup percentage and block duration by governance
func SetWarmUp(cur realm, pct, blockDuration int64) {
	assertOnlyGovernance()
	assertMustNotHalted()

	setWarmUp(cur, pct, blockDuration)
}

func setWarmUp(cur realm, pct, blockDuration int64) {
	en.MintAndDistributeGns(cross)

	modifyWarmup(pctToIndex(pct), blockDuration)

	prevAddr, prevRealm := getPrev()
	std.Emit(
		"SetWarmUp",
		"prevAddr", prevAddr,
		"prevRealm", prevRealm,
		"pct", formatInt(pct),
		"blockDuration", formatInt(blockDuration),
	)
}

func pctToIndex(pct int64) int {
	switch pct {
	case 30:
		return 0
	case 50:
		return 1
	case 70:
		return 2
	case 100:
		return 3
	default:
		panic("staker.gno__pctToIndex() || pct is not valid")
	}
}

func assertPoolMustExist(poolPath string) {
	if !(pl.DoesPoolPathExist(poolPath)) {
		panic(addDetailToError(
			errInvalidPoolPath,
			ufmt.Sprintf("pool(%s) does not exist", poolPath),
		))
	}
}

func assertOnlyAdmin() {
	access.AdminOnly(getPrevAddr())
}

func assertOnlyGovernance() {
	access.GovernanceOnly(getPrevAddr())
}

func assertMustNotHalted() {
	if err := halt.IsHalted(); err != nil {
		panic(err)
	}
}
