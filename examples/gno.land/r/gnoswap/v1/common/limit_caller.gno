package common

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
)

// limitCaller is a global boolean flag that controls whether function calls are restricted.
// Default value is true, meaning call restrictions are enabled by default.
var limitCaller bool = true

// GetLimitCaller returns the current state of the limitCaller flag.
// If true, call restrictions are active; if false, call restrictions are disabled.
//
// Returns:
// - bool: Current state of the limitCaller (true if active, false if inactive).
func GetLimitCaller() bool {
	return limitCaller
}

// SetLimitCaller updates the limitCaller flag to either enable or disable call restrictions.
// This function can only be called by an admin. If a non-admin attempts to call this function,
// the function will panic.
//
// Parameters:
// - v (bool): The new state for the limitCaller flag (true to enable, false to disable).
//
// Panics:
// - If the caller is not an admin, the function panics with an errNoPermission error.
func SetLimitCaller(v bool) {
	caller := getPrevAddr()
	if err := AdminOnly(caller); err != nil {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf(
				"only Admin can set halt, called from %s",
				caller,
			)),
		)
	}

	limitCaller = v

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"SetLimitCaller",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"limitCaller", strconv.FormatBool(v),
	)
}
