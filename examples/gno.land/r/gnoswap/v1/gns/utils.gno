package gns

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

func getPrev() (string, string) {
	prev := std.PrevRealm()
	return prev.Addr().String(), prev.PkgPath()
}

func getPrevAddr() std.Address {
	return std.PrevRealm().Addr()
}

func assertShouldNotBeHalted() {
	common.IsHalted()
}

func assertCallerIsEmission() {
	caller := getPrevAddr()
	if err := common.EmissionOnly(caller); err != nil {
		panic(err)
	}
}

func assertCallerIsAdmin() {
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err)
	}
}

func assertCallerIsGovernance() {
	caller := getPrevAddr()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err)
	}
}

func assertValidYear(year int64) {
	if year < HALVING_START_YEAR || year > HALVING_END_YEAR {
		panic(addDetailToError(errInvalidYear, ufmt.Sprintf("year: %d", year)))
	}
}

func i64Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func secToMs(sec int64) int64 {
	return sec * consts.MILLISECONDS_PER_SECOND
}

func milliToSec(ms int64) int64 {
	return ms / consts.MILLISECONDS_PER_SECOND
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
