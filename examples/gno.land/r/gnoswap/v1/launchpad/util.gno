package launchpad

import (
	"std"
	"strconv"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gns"
)

// Converts a string to an integer. Panics if the conversion fails.
func strToInt(str string) int {
	res, err := strconv.Atoi(str)
	if err != nil {
		panic(err.Error())
	}
	return res
}

// Converts a string to a uint64 using a decimal representation.
func strToU256U64(str string) uint64 {
	strValue := u256.MustFromDecimal(str)
	return strValue.Uint64()
}

// Checks if a string is present in a slice of strings.
// Returns true if the string is found, otherwise false.
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// Marshals a JSON node into a string. Panics if the marshalling fails.
func marshal(data *json.Node) string {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}

// Returns the minimum of two uint64 values.
func minU64(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

// containsString returns true if the string is in the slice.
func containsString(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

// GetOrigPkgAddr returns the original package address.
// In position contract, original package address is the position address.
func GetOrigPkgAddr() std.Address {
	return consts.LAUNCHPAD_ADDR
}

// isUserCall returns true if the caller is a user.
func isUserCall() bool {
	return std.PrevRealm().IsUser()
}

// getPrev returns the address and package path of the previous realm.
func getPrev() (string, string) {
	prev := std.PrevRealm()
	return prev.Addr().String(), prev.PkgPath()
}

// getPrevRealm returns object of the previous realm.
func getPrevRealm() std.Realm {
	return std.PrevRealm()
}

// getPrevAddr returns the address of the previous realm.
func getPrevAddr() std.Address {
	return std.PrevRealm().Addr()
}

// getPrevPkgPath returns the package path of the previous realm.
func getPrevPkgPath() string {
	return std.PrevRealm().PkgPath()
}

// formatUint returns the string representation of the uint64 value.
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

// formatInt returns the string representation of the int64 value.
func formatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

// assertOnlyNotHalted panics if the contract is halted.
func assertOnlyNotHalted() {
	common.IsHalted()
}

// assertOnlyAdmin panics if the caller is not an admin.
func assertOnlyAdmin() {
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(addDetailToError(errInvalidAddress, err.Error()))
	}
}

// assertOnlyValidAddress panics if the address is invalid.
func assertOnlyValidAddress(addr std.Address) {
	if !addr.IsValid() {
		panic(addDetailToError(
			errInvalidAddress,
			ufmt.Sprintf("(%s)", addr),
		))
	}
}

// assertValidAmount panics if the amount is zero.
func assertValidAmount(amount uint64) {
	if amount < MINIMUM_DEPOSIT_AMOUNT {
		panic(addDetailToError(
			errInvalidAmount, ufmt.Sprintf("amount(%d) should greater than minimum deposit amount(%d)", amount, MINIMUM_DEPOSIT_AMOUNT)))
	}
	if (amount % MINIMUM_DEPOSIT_AMOUNT) != 0 {
		panic(addDetailToError(
			errInvalidAmount, ufmt.Sprintf("amount(%d) must be a multiple of 1_000_000", amount)))
	}
}

// assertValidProjectId panics if the project ID is invalid.
func assertValidProjectId(projectId string) {
	if projectId == "" {
		panic(addDetailToError(
			errInvalidProjectId, ufmt.Sprintf("(%s)", projectId)))
	}
}

// convertTimeToHeight converts the given timestamp to block height based on the average block time.
func convertTimeToHeight(timestamp uint64) uint64 {
	avgBlockTimeMs := uint64(gns.GetAvgBlockTimeInMs())
	if avgBlockTimeMs == 0 {
		panic(addDetailToError(
			errInvalidAvgBlockTime,
			ufmt.Sprintf("%d", avgBlockTimeMs)))
	}
	return timestamp * 1000 / avgBlockTimeMs
}
