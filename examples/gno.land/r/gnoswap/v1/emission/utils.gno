package emission

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/access"

	"gno.land/r/gnoswap/v1/halt"
)

const MAX_BPS_PCT uint64 = 10000

// getPrevRealm returns object of the previous realm.
func getPrevRealm() std.Realm {
	return std.PreviousRealm()
}

// getPrevAddr returns the address of the previous realm.
func getPrevAddr() std.Address {
	return std.PreviousRealm().Address()
}

// getPrev returns the address and package path of the previous realm.
func getPrevAsString() (string, string) {
	prev := std.PreviousRealm()
	return prev.Address().String(), prev.PkgPath()
}

// assertOnlyAdmin panics if the caller is not the admin.
func assertOnlyAdmin() {
	caller := getPrevAddr()
	if err := access.AdminOnly(caller); err != nil {
		panic(err.Error())
	}
}

// assertStakerOnly panics if the caller is not the staker.
func assertStakerOnly() {
	caller := getPrevAddr()
	if err := access.StakerOnly(caller); err != nil {
		panic(err.Error())
	}
}

// assertOnlyGovernance panics if the caller is not the governance.
func assertOnlyGovernance() {
	caller := getPrevAddr()
	if err := access.GovernanceOnly(caller); err != nil {
		panic(err.Error())
	}
}

// assertOnlyGovStaker panics if the caller is not the gov staker.
func assertOnlyGovStaker() {
	caller := getPrevAddr()
	if err := access.GovStakerOnly(caller); err != nil {
		panic(err.Error())
	}
}

// assertDistributionTarget panics if the target is invalid.
func assertDistributionTarget(target int) {
	if target != LIQUIDITY_STAKER && target != DEVOPS && target != COMMUNITY_POOL && target != GOV_STAKER {
		panic(addDetailToError(
			errInvalidEmissionTarget,
			ufmt.Sprintf("invalid target(%d)", target),
		))
	}
}

// assertOnlyNotHalted panics if the contract is halted.
func assertOnlyNotHalted() {
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			panic(err.Error())
		}

		if halt.IsContractHalted(halt.OpTypeEmission) {
			panic("emission contract operations are currently disabled")
		}
	}
}

// assertSumDistributionPct ensures the sum of all distribution percentages is 100%
func assertSumDistributionPct(pct01, pct02, pct03, pct04 int64) {
	sum := pct01 + pct02 + pct03 + pct04
	if sum != 10000 {
		panic(addDetailToError(
			errInvalidEmissionTarget,
			ufmt.Sprintf("sum of percentages must be 10000, got %d", sum),
		))
	}
}

// assertNoDuplicateTargets ensures no duplicate targets are provided
func assertNoDuplicateTargets(target01, target02, target03, target04 int) {
	targets := []int{target01, target02, target03, target04}
	seen := make(map[int]bool)

	for _, target := range targets {
		if seen[target] {
			panic(addDetailToError(
				errDuplicateTarget,
				ufmt.Sprintf("target %d appears multiple times", target),
			))
		}
		seen[target] = true
	}
}

func formatUint(v any) string {
	switch v := v.(type) {
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		panic(ufmt.Sprintf("invalid type: %T", v))
	}
}

func formatInt(v any) string {
	switch v := v.(type) {
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	default:
		panic(ufmt.Sprintf("invalid type: %T", v))
	}
}
