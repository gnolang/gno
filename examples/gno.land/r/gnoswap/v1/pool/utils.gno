package pool

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/halt"
)

// safeConvertToUint64 safely converts a *u256.Uint value to a uint64, ensuring no overflow.
//
// This function attempts to convert the given *u256.Uint value to a uint64. If the value exceeds
// the maximum allowable range for uint64 (`2^64 - 1`), it triggers a panic with a descriptive error message.
//
// Parameters:
// - value (*u256.Uint): The unsigned 256-bit integer to be converted.
//
// Returns:
// - uint64: The converted value if it falls within the uint64 range.
//
// Panics:
//   - If the `value` exceeds the range of uint64, the function will panic with an error indicating
//     the overflow and the original value.
//
// Notes:
// - This function uses the `Uint64WithOverflow` method to detect overflow during the conversion.
// - It is essential to validate large values before calling this function to avoid unexpected panics.
//
// Example:
// safeValue := safeConvertToUint64(u256.MustFromDecimal("18446744073709551615")) // Valid conversion
// safeConvertToUint64(u256.MustFromDecimal("18446744073709551616")) // Panics due to overflow
func safeConvertToUint64(value *u256.Uint) uint64 {
	res, overflow := value.Uint64WithOverflow()
	if overflow {
		panic(ufmt.Sprintf(
			"%v: amount(%s) overflows uint64 range (max %s)",
			errOutOfRange,
			value.ToString(),
			consts.MAX_UINT64,
		))
	}
	return res
}

// safeConvertToInt64 safely converts a *u256.Uint value to an int64, ensuring no overflow.
//
// This function attempts to convert the given *u256.Uint value to an int64. If the value exceeds
// the maximum allowable range for int64 (`2^63 - 1`), it triggers a panic with a descriptive error message.
//
// Parameters:
// - value (*u256.Uint): The unsigned 256-bit integer to be converted.
//
// Returns:
// - int64: The converted value if it falls within the int64 range.
//
// Panics:
//   - If the `value` exceeds the range of int64, the function will panic with an error indicating
//     the overflow and the original value.
//
// Notes:
// - This function first checks for uint64 overflow, then validates the int64 range.
// - The maximum value for int64 is 9,223,372,036,854,775,807 (2^63 - 1).
//
// Example:
// safeValue := safeConvertToInt64(u256.MustFromDecimal("9223372036854775807")) // Valid conversion
// safeConvertToInt64(u256.MustFromDecimal("9223372036854775808")) // Panics due to overflow
func safeConvertToInt64(value *u256.Uint) int64 {
	res, overflow := value.Uint64WithOverflow()
	if overflow || res > uint64(consts.INT64_MAX) {
		panic(ufmt.Sprintf(
			"%v: amount(%s) overflows int64 range (max %s)",
			errOutOfRange,
			value.ToString(),
			consts.MAX_INT64,
		))
	}
	return int64(res)
}

// safeConvertToInt128 safely converts a *u256.Uint value to an *i256.Int, ensuring it does not exceed the int128 range.
//
// This function converts an unsigned 256-bit integer (*u256.Uint) into a signed 256-bit integer (*i256.Int).
// It checks whether the resulting value falls within the valid range of int128 (`-2^127` to `2^127 - 1`).
// If the value exceeds the maximum allowable int128 range, it triggers a panic with a descriptive error message.
//
// Parameters:
// - value (*u256.Uint): The unsigned 256-bit integer to be converted.
//
// Returns:
// - *i256.Int: The converted value if it falls within the int128 range.
//
// Panics:
//   - If the converted value exceeds the maximum int128 value (`2^127 - 1`), the function will panic with an
//     error message indicating the overflow and the original value.
//
// Notes:
// - The function uses `i256.FromUint256` to perform the conversion.
// - The constant `MAX_INT128` is used to define the upper bound of the int128 range (`170141183460469231731687303715884105727`).
//
// Example:
// validInt128 := safeConvertToInt128(u256.MustFromDecimal("170141183460469231731687303715884105727")) // Valid conversion
// safeConvertToInt128(u256.MustFromDecimal("170141183460469231731687303715884105728")) // Panics due to overflow
func safeConvertToInt128(value *u256.Uint) *i256.Int {
	liquidityDelta := i256.FromUint256(value)
	if liquidityDelta.Gt(i256.MustFromDecimal(consts.MAX_INT128)) {
		panic(ufmt.Sprintf(
			"%v: amount(%s) overflows int128 range",
			errOverFlow, value.ToString()))
	}
	return liquidityDelta
}

// toUint128 ensures a *u256.Uint value fits within the uint128 range.
//
// This function validates that the given `value` is properly initialized and checks whether
// it exceeds the maximum value of uint128. If the value exceeds the uint128 range,
// it applies a masking operation to truncate the value to fit within the uint128 limit.
//
// Parameters:
//   - value: *u256.Uint, the value to be checked and possibly truncated.
//
// Returns:
//   - *u256.Uint: A value guaranteed to fit within the uint128 range.
//
// Notes:
//   - The function first checks if the value is not nil to avoid potential runtime errors.
//   - The mask ensures that only the lower 128 bits of the value are retained.
//   - If the input value is already within the uint128 range, it is returned unchanged.
//   - If masking is required, a new instance is returned without modifying the input.
//   - MAX_UINT128 is a constant representing `2^128 - 1`.
func toUint128(value *u256.Uint) *u256.Uint {
	assertOnlyInitializedUint256(value)
	if value.Gt(u256.MustFromDecimal(consts.MAX_UINT128)) {
		mask := new(u256.Uint).Lsh(u256.One(), consts.Q128_RESOLUTION)
		mask = new(u256.Uint).Sub(mask, u256.One())
		return new(u256.Uint).And(value, mask)
	}
	return value
}

// u256Min returns the smaller of two *u256.Uint values.
//
// This function compares two unsigned 256-bit integers and returns the smaller of the two.
// If `num1` is less than `num2`, it returns `num1`; otherwise, it returns `num2`.
//
// Parameters:
// - num1 (*u256.Uint): The first unsigned 256-bit integer.
// - num2 (*u256.Uint): The second unsigned 256-bit integer.
//
// Returns:
// - *u256.Uint: The smaller of `num1` and `num2`.
//
// Notes:
//   - This function uses the `Lt` (less than) method of `*u256.Uint` to perform the comparison.
//   - The function assumes both input values are non-nil. If nil inputs are possible in the usage context,
//     additional validation may be needed.
//
// Example:
// smaller := u256Min(u256.MustFromDecimal("10"), u256.MustFromDecimal("20")) // Returns 10
// smaller := u256Min(u256.MustFromDecimal("30"), u256.MustFromDecimal("20")) // Returns 20
func u256Min(num1, num2 *u256.Uint) *u256.Uint {
	if num1.Lt(num2) {
		return num1
	}
	return num2
}

// derivePkgAddr derives the Realm address from it's pkgPath parameter
func derivePkgAddr(pkgPath string) std.Address {
	return std.DerivePkgAddr(pkgPath)
}

// getPrevRealm returns object of the previous realm.
func getPrevRealm() std.Realm {
	return std.PreviousRealm()
}

// getPrevAddr returns the address of the previous realm.
func getPrevAddr() std.Address {
	return std.PreviousRealm().Address()
}

// getPrevAsString returns the address and package path of the previous realm.
func getPrevAsString() (string, string) {
	prev := getPrevRealm()
	return prev.Address().String(), prev.PkgPath()
}

// isUserCall returns true if the caller is a user.
func isUserCall() bool {
	return std.PreviousRealm().IsUser()
}

// checkTransferError checks transfer error.
func checkTransferError(err error) {
	if err != nil {
		panic(newErrorWithDetail(
			errTransferFailed,
			err.Error(),
		))
	}
}

// checkOverFlowInt128 checks if the value overflows the int128 range.
func checkOverFlowInt128(value *i256.Int) {
	if value.Gt(i256.MustFromDecimal(consts.MAX_INT128)) {
		panic(ufmt.Sprintf(
			"%v: amount(%s) overflows int128 range",
			errOverFlow, value.ToString()))
	}
}

// checkTickSpacing checks if the tick is divisible by the tickSpacing.
func checkTickSpacing(tick, tickSpacing int32) {
	if tick%tickSpacing != 0 {
		panic(newErrorWithDetail(
			errInvalidTickAndTickSpacing,
			ufmt.Sprintf("tick(%d) MOD tickSpacing(%d) != 0(%d)", tick, tickSpacing, tick%tickSpacing),
		))
	}
}

// assertOnlyInitializedUint256 (assertOnlyNotNil) panics if the value is nil.
func assertOnlyInitializedUint256(value *u256.Uint) {
	if value == nil {
		panic(newErrorWithDetail(
			errInvalidInput,
			"value is nil",
		))
	}
}

// assertOnlyValidAddress panics if the address is invalid.
func assertOnlyValidAddress(addr std.Address) {
	if !addr.IsValid() {
		panic(newErrorWithDetail(
			errInvalidAddress,
			ufmt.Sprintf("(%s)", addr),
		))
	}
}

// assertOnlyNotHalted panics if the contract is halted.
func assertOnlyNotHalted() {
	currentLevel := halt.GetCurrentHaltLevel()
	if currentLevel != halt.LvMainnetSafeMode {
		// Check if withdrawals are specifically halted, not just if the system is halted
		if err := halt.IsHalted(halt.OpTypeWithdraw); err != nil {
			panic(err.Error())
		}

		if halt.IsContractHalted(halt.OpTypePool) {
			panic("pool contract operations are currently disabled")
		}
	}
}

// assertOnlyPositionContract panics if the caller is not the position contract.
func assertOnlyPositionContract() {
	caller := getPrevAddr()
	if err := access.PositionOnly(caller); err != nil {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf("only position(%s) can call, called from %s", positionAddr, caller.String()),
		))
	}
}

// assertOnlyAdmin panics if the caller is not the admin.
func assertOnlyAdmin() {
	caller := getPrevAddr()
	if err := access.AdminOnly(caller); err != nil {
		panic(err)
	}
}

// assertOnlyGovernance panics if the caller is not the governance.
func assertOnlyGovernance() {
	caller := getPrevAddr()
	if err := access.GovernanceOnly(caller); err != nil {
		panic(err)
	}
}

// assertOnlyRegistered panics if the token is not registered.
func assertOnlyRegistered(tokenPath string) {
	common.MustRegistered(tokenPath)
}

// formatUint converts various unsigned integer types to string representation.
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

// formatInt converts various signed integer types to string representation.
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

// formatBool converts a boolean value to string representation.
func formatBool(v bool) string {
	return strconv.FormatBool(v)
}