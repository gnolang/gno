package emission

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/common"
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
	common.IsHalted()
}

// assertSumDistributionPct ensures the sum of all distribution percentages is 100%
func assertSumDistributionPct(pct01, pct02, pct03, pct04 uint64) {
	sum := pct01 + pct02 + pct03 + pct04

	if sum != MAX_BPS_PCT {
		panic(addDetailToError(
			errInvalidEmissionPct,
			ufmt.Sprintf("sum of all pct should be 100%% (10000 bps), got %d", sum),
		))
	}
}

func formatUint(v interface{}) string {
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

func formatInt(v interface{}) string {
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
