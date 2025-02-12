package common

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
)

// halted is a global flag that indicates whether the GnoSwap is currently halted.
// When true, most operations are disabled to prevent further actions.
// Default value is false, meaning the GnoSwap is active by default.
var halted bool = false

// GetHalt returns the current halted status of the GnoSwap.
//
// Returns:
// - bool: true if the GnoSwap is halted, false otherwise.
func GetHalt() bool {
	return halted
}

// IsHalted checks if the GnoSwap is currently halted.
// If the GnoSwap is halted, the function panics with an errHalted error.
//
// Panics:
// - If the halted flag is true, indicating that the GnoSwap is inactive.
func IsHalted() {
	if halted {
		panic(newErrorWithDetail(
			errHalted,
			"GnoSwap is halted",
		))
	}
}

// SetHaltByAdmin allows an admin to set the halt status of the GnoSwap.
// Only an admin can execute this function. If a non-admin attempts to call this function,
// the function panics with an errNoPermission error.
//
// Parameters:
// - halt (bool): The new halt status to set (true to halt, false to unhalt).
//
// Panics:
// - If the caller is not an admin, the function will panic with an errNoPermission error.
func SetHaltByAdmin(halt bool) {
	caller := getPrevAddr()
	if err := AdminOnly(caller); err != nil {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf(
				"only admin(%s) can set halt, called from %s",
				consts.ADMIN,
				caller,
			),
		))
	}
	setHalt(halt)
}

// SetHalt allows the governance contract to set the halt status of the GnoSwap.
// Only the governance contract can execute this function through a proposal process.
//
// Parameters:
// - halt (bool): The new halt status to set (true to halt, false to unhalt).
//
// Panics:
// - If the caller is not the governance contract, the function will panic with an errNoPermission error.
func SetHalt(halt bool) {
	caller := getPrevAddr()
	if err := GovernanceOnly(caller); err != nil {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf(
				"only governance(%s) can set halt, called from %s",
				consts.GOV_GOVERNANCE_ADDR,
				caller,
			),
		))
	}
	setHalt(halt)
}

// setHalt updates the halted flag to the specified value.
// This is an internal function that should only be called by SetHalt or SetHaltByAdmin.
//
// Parameters:
// - halt (bool): The new halt status to set.
func setHalt(halt bool) {
	halted = halt

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"setHalt",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"halt", strconv.FormatBool(halt),
	)
}
