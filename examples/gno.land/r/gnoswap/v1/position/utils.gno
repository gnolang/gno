package position

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gnft"

	"gno.land/r/gnoswap/v1/halt"

	u256 "gno.land/p/gnoswap/uint256"
)

// GetOrigPkgAddr returns the original package address.
// In position contract, original package address is the position address.
func GetOrigPkgAddr() std.Address {
	return positionAddr
}

// assertTokenExists checks if a token with the specified positionId exists in the system.
//
// This function verifies the existence of a token by its positionId. If the token does not exist,
// it triggers a panic with a descriptive error message.
//
// Parameters:
// - positionId (uint64): The unique identifier of the token to check.
//
// Panics:
// - If the token does not exist, it panics with the error `errDataNotFound`.
func assertTokenExists(positionId uint64) {
	if !exists(positionId) {
		panic(newErrorWithDetail(
			errDataNotFound,
			ufmt.Sprintf("positionId(%d) doesn't exist", positionId),
		))
	}
}

// assertOnlyOwnerOfToken ensures that only the owner of the specified token can perform operations on it.
//
// This function checks if the caller is the owner of the token identified by `positionId`.
// If the caller is not the owner, the function triggers a panic to prevent unauthorized access.
//
// Parameters:
// - positionId (uint64): The unique identifier of the token to check ownership.
// - caller (std.Address): The address of the entity attempting to modify the token.
//
// Panics:
// - If the caller is not the owner of the token, the function panics with an `errNoPermission` error.
func assertOnlyOwnerOfToken(positionId uint64, caller std.Address) {
	owner, err := gnft.OwnerOf(positionIdFrom(positionId))
	if err != nil {
		panic(newErrorWithDetail(
			errDataNotFound,
			ufmt.Sprintf("positionId(%d) doesn't exist", positionId),
		))
	}
	assertCallerIsOwner(positionId, owner, caller)
}

func assertCallerIsOwner(positionId uint64, owner, caller std.Address) {
	if owner != caller {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf("caller(%s) is not owner(%s) for positionId(%d)", caller, owner, positionId),
		))
	}
}

// assertOnlyUserOrStaker panics if the caller is not a user or staker.
func assertOnlyUserOrStaker(caller std.Realm) {
	if !caller.IsUser() {
		if err := common.StakerOnly(caller.Address()); err != nil {
			panic(newErrorWithDetail(
				errNoPermission,
				ufmt.Sprintf("from (%s)", caller.Address()),
			))
		}
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

		if halt.IsContractHalted(halt.OpTypePosition) {
			panic("position contract operations are currently disabled")
		}
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

// assertOnlyValidAddress panics if the address is invalid or previous address is not
// different from the other address.
func assertOnlyValidAddressWith(prevAddr, otherAddr std.Address) {
	assertOnlyValidAddress(prevAddr)
	assertOnlyValidAddress(otherAddr)

	if prevAddr != otherAddr {
		panic(newErrorWithDetail(
			errInvalidAddress,
			ufmt.Sprintf("(%s, %s)", prevAddr, otherAddr),
		))
	}
}

// assertValidNumberString verifies that the input string represents a valid integer.
//
// This function checks each character in the string to ensure it falls within the numeric ASCII range ('0' to '9').
// The first character is allowed to be a negative sign ('-'), indicating a negative number. If the input does not
// meet these criteria, the function panics with a detailed error message.
//
// Parameters:
//   - input (string): The string to validate.
//
// Panics:
//   - If the input string is empty.
//   - If the input contains non-numeric characters (excluding an optional leading '-').
//   - If the input is not a valid integer representation.
//
// Example:
//
//	assertValidNumberString("12345")       -> Pass (valid positive number)
//	assertValidNumberString("-98765")      -> Pass (valid negative number)
//	assertValidNumberString("12a45")       -> Panic (invalid character 'a')
//	assertValidNumberString("")            -> Panic (empty input)
//	assertValidNumberString("++123")       -> Panic (invalid leading '+')
func assertValidNumberString(input string) {
	if len(input) == 0 {
		panic(newErrorWithDetail(
			errInvalidInput,
			ufmt.Sprintf("input is empty")))
	}

	bytes := []byte(input)
	for i, b := range bytes {
		if i == 0 && b == '-' {
			continue // Allow if the first character is a negative sign (-)
		}
		if b < '0' || b > '9' {
			panic(newErrorWithDetail(
				errInvalidInput,
				ufmt.Sprintf("input string : %s", input)))
		}
	}
}

func assertValidLiquidityAmount(liquidity string) {
	if u256.MustFromDecimal(liquidity).IsZero() {
		panic(newErrorWithDetail(
			errZeroLiquidity,
			ufmt.Sprintf("liquidity amount must be greater than 0, got %s", liquidity),
		))
	}
}

// derivePkgAddr derives the Realm address from it's pkgpath parameter
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

// getPrev returns the address and package path of the previous realm.
func getPrevAsString() (string, string) {
	prev := getPrevRealm()
	return prev.Address().String(), prev.PkgPath()
}

// isUserCall returns true if the caller is a user.
func isUserCall() bool {
	return std.PreviousRealm().IsUser()
}

// checkDeadline checks if the deadline is expired.
// If the deadline is expired, it panics.
// The deadline is expired if the current time is greater than the deadline.
// Input:
//   - deadline: the deadline to check
func checkDeadline(deadline int64) {
	now := time.Now().Unix()
	if now > deadline {
		panic(newErrorWithDetail(
			errExpired,
			ufmt.Sprintf("transaction too old, now(%d) > deadline(%d)", now, deadline),
		))
	}
}

// positionIdFrom converts positionId to grc721.TokenID type
// NOTE: input parameter positionId can be string, int, uint64, or grc721.TokenID
// if positionId is nil or not supported, it will panic
// if positionId is not found, it will panic
// input: positionId any
// output: grc721.TokenID
func positionIdFrom(positionId any) grc721.TokenID {
	if positionId == nil {
		panic(newErrorWithDetail(errInvalidInput, "positionId is nil"))
	}

	switch positionId.(type) {
	case string:
		return grc721.TokenID(positionId.(string))
	case int:
		return grc721.TokenID(strconv.Itoa(positionId.(int)))
	case uint64:
		return grc721.TokenID(strconv.Itoa(int(positionId.(uint64))))
	case grc721.TokenID:
		return positionId.(grc721.TokenID)
	default:
		panic(newErrorWithDetail(errInvalidInput, "unsupported positionId type"))
	}
}

// exists checks whether positionId exists
// If positionId doesn't exist, return false, otherwise return true
// input: positionId uint64
// output: bool
func exists(positionId uint64) bool {
	return gnft.Exists(positionIdFrom(positionId))
}

// isOwner checks whether the caller is the owner of the positionId
// If the caller is the owner of the positionId, return true, otherwise return false
// input: positionId uint64, addr std.Address
// output: bool
func isOwner(positionId uint64, addr std.Address) bool {
	owner, err := gnft.OwnerOf(positionIdFrom(positionId))
	if err == nil {
		if owner == addr {
			return true
		}
	}
	return false
}

// isOperator checks whether the caller is the approved operator of the positionId
// If the caller is the approved operator of the positionId, return true, otherwise return false
// input: positionId uint64, addr std.Address
// output: bool
func isOperator(positionId uint64, addr std.Address) bool {
	operator, err := gnft.GetApproved(positionIdFrom(positionId))
	if err == nil && operator == addr {
		return true
	}
	return false
}

// isStaked checks whether positionId is staked
// If positionId is staked, owner of positionId is staker contract
// If positionId is staked, return true, otherwise return false
// input: positionId grc721.TokenID
// output: bool
func isStaked(positionId grc721.TokenID) bool {
	exist := gnft.Exists(positionId)
	if exist {
		owner, err := gnft.OwnerOf(positionId)
		if err == nil && owner == stakerAddr {
			return true
		}
	}
	return false
}

// isOwnerOrOperator checks whether the caller is the owner or approved operator of the positionId
// If the caller is the owner or approved operator of the positionId, return true, otherwise return false
// input: addr std.Address, positionId uint64
// output: bool
func isOwnerOrOperator(addr std.Address, positionId uint64) bool {
	assertOnlyValidAddress(addr)
	if !exists(positionId) {
		return false
	}
	if isOwner(positionId, addr) || isOperator(positionId, addr) {
		return true
	}
	if isStaked(positionIdFrom(positionId)) {
		position, exist := GetPosition(positionId)
		if exist && addr == position.operator {
			return true
		}
	}
	return false
}

func assertAuthorizedForToken(caller std.Address, positionId uint64) {
	if !isOwnerOrOperator(caller, positionId) {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf("caller(%s) is not owner or approved operator of positionId(%d)", caller, positionId),
		))
	}
}

// splitOf divides poolKey into pToken0, pToken1, and pFee
// If poolKey is invalid, it will panic
//
// input: poolKey string
// output:
// - token0Path string
// - token1Path string
// - fee uint32
func splitOf(poolKey string) (string, string, uint32) {
	res, err := common.Split(poolKey, ":", 3)
	if err != nil {
		panic(newErrorWithDetail(errInvalidInput, ufmt.Sprintf("invalid poolKey(%s)", poolKey)))
	}
	pToken0, pToken1, pFeeStr := res[0], res[1], res[2]

	pFee, err := strconv.Atoi(pFeeStr)
	if err != nil {
		panic(newErrorWithDetail(errInvalidInput, ufmt.Sprintf("invalid fee(%s)", pFeeStr)))
	}
	return pToken0, pToken1, uint32(pFee)
}

func verifyTokenIdAndOwnership(positionId uint64) {
	assertTokenExists(positionId)
	assertOnlyOwnerOfToken(positionId, getPrevAddr())
}

func verifySlippageAmounts(amount0, amount1, amount0Min, amount1Min *u256.Uint) {
	if !(amount0.Gte(amount0Min) && amount1.Gte(amount1Min)) {
		panic(newErrorWithDetail(
			errSlippage,
			ufmt.Sprintf("amount0(%s) >= amount0Min(%s) && amount1(%s) >= amount1Min(%s)",
				amount0.ToString(), amount0Min.ToString(), amount1.ToString(), amount1Min.ToString()),
		))
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

func formatBool(v bool) string {
	return strconv.FormatBool(v)
}
