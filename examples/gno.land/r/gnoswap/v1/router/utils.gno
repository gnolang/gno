package router

import (
	"bytes"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/emission"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/r/gnoswap/v1/halt"
)

var (
	errRouterHalted = "router contract operations are currently disabled"
	errTxExpired    = "transaction too old, now(%d) > deadline(%d)"
)

// assertHopsInRange panics if the number of hops is not in the range 1~3.
func assertHopsInRange(hops int) {
	switch hops {
	case 1, 2, 3:
		return
	default:
		panic(errHopsOutOfRange)
	}
}

// getDataForSinglePath extracts token0, token1, and fee from a single path.
func getDataForSinglePath(poolPath string) (token0, token1 string, fee uint32) {
	poolPathSplit, err := common.Split(poolPath, ":", 3)
	if err != nil {
		panic(addDetailToError(
			errInvalidPoolPath,
			ufmt.Sprintf("len(poolPathSplit) != 3, poolPath: %s", poolPath),
		))
	}

	f, err := strconv.Atoi(poolPathSplit[2])
	if err != nil {
		panic(ufmt.Sprintf("invalid fee: %s", poolPathSplit[2]))
	}

	return poolPathSplit[0], poolPathSplit[1], uint32(f)
}

// getDataForMultiPath extracts token0, token1, and fee from a multi path.
func getDataForMultiPath(possiblePath string, poolIdx int) (token0, token1 string, fee uint32) {
	pools := strings.Split(possiblePath, POOL_SEPARATOR)

	switch poolIdx {
	case 0:
		return getDataForSinglePath(pools[0])
	case 1:
		return getDataForSinglePath(pools[1])
	case 2:
		return getDataForSinglePath(pools[2])
	default:
		return "", "", uint32(0)
	}
}

func i256MinMax(x, y *i256.Int) (min, max *u256.Uint) {
	if x.Lt(y) || x.Eq(y) {
		return x.Abs(), y.Abs()
	}
	return y.Abs(), x.Abs()
}

// getPrevAddr returns the address of the previous realm.
func getPrevAddr() std.Address {
	return std.PreviousRealm().Address()
}

// getPrev returns the address and package path of the previous realm.
func getPrev() (string, string) {
	prev := std.PreviousRealm()
	return prev.Address().String(), prev.PkgPath()
}

// haltCheck performs security validations before any swap operation
func haltCheck() error {
	currentLevel := halt.GetCurrentHaltLevel()

	if currentLevel != halt.LvMainnetSafeMode {
		// check for withdraw operation halt
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			return err
		}

		// check for router contract halt
		if halt.IsContractHalted(halt.OpTypeRouter) {
			return ufmt.Errorf(errRouterHalted)
		}
	}

	return nil
}

// commonSwapSetup performs all necessary validations and setup for a swap operation
func commonSwapSetup() error {
	if err := haltCheck(); err != nil {
		return err
	}

	if common.GetLimitCaller() && std.PreviousRealm().PkgPath() != "" {
		return ufmt.Errorf("%v: only user can call this function", errNoPermission)
	}

	emission.MintAndDistributeGns(cross)
	return nil
}

// checkDeadline checks if the deadline is expired.
// If the deadline is expired, it panics.
// The deadline is expired if the current time is greater than the deadline.
// Input:
//   - deadline: the deadline to check
func checkDeadline(deadline int64) error {
	now := time.Now().Unix()
	if now <= deadline {
		return nil
	}

	return ufmt.Errorf(errTxExpired, now, deadline)
}

// splitSingleChar splits a string by a single character separator.
//
// This function is optimized for splitting strings with a single-byte separator.
// Unlike `strings.Split`, it:
// 1. Performs direct byte comparison instead of substring matching
// 2. Avoids additional string allocations by using slicing
// 3. Makes only one allocation for the result slice
//
// The main differences from `strings.Split` are:
// - Only works with single-byte separators
// - More memory efficient as it doesn't need to handle multi-byte separators
// - Faster for small to medium strings due to simpler byte comparison
//
// Performance:
//   - Up to 5x faster than `strings.Split` for small strings (in Go)
//   - For gno (run test with `-print-runtime-metrics` option):
//     |     Function     |     Cycles       | Allocations
//     |------------------|------------------|--------------|
//     | strings.Split    |      1.1M        | 808.1K       |
//     | splitSingleChar  |      1.0M        | 730.4K       |
//   - Uses zero allocations except for the initial result slice
//   - Most effective for strings under 1KB with simple single-byte delimiters
//     (* This test result was measured without the `uassert` package)
//
// Parameters:
//
//	s (string): source string to split
//	sep (byte): single byte separator to split on
//
// Returns:
//
//	[]string: slice containing the split string parts
func splitSingleChar(s string, sep byte) []string {
	if s == "" {
		return []string{""}
	}

	result := make([]string, 0, bytes.Count([]byte(s), []byte{sep})+1)
	start := 0
	for i := range s {
		if s[i] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
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

func formatInt64(v any) string {
	switch v := v.(type) {
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		panic(ufmt.Sprintf("invalid type %T", v))
	}
}
