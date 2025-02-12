package common

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	errNoPermission   = errors.New("[GNOSWAP-COMMON-001] caller has no permission")
	errHalted         = errors.New("[GNOSWAP-COMMON-002] halted")
	errOutOfRange     = errors.New("[GNOSWAP-COMMON-003] value out of range")
	errNotRegistered  = errors.New("[GNOSWAP-COMMON-004] token is not registered")
	errInvalidAddr    = errors.New("[GNOSWAP-COMMON-005] invalid address")
	errOverflow       = errors.New("[GNOSWAP-COMMON-006] overflow")
	errInvalidTokenId = errors.New("[GNOSWAP-COMMON-007] invalid tokenId")
	errInvalidInput   = errors.New("[GNOSWAP-COMMON-008] invalid input data")
	errOverFlow       = errors.New("[GNOSWAP-COMMON-009] overflow")
	errIdenticalTicks = errors.New("[GNOSWAP-COMMON-010] identical ticks")
)

// newErrorWithDetail appends additional context or details to an existing error message.
//
// Parameters:
//   - err: The original error (error).
//   - detail: Additional context or detail to append to the error message (string).
//
// Returns:
//   - string: The combined error message in the format "<original_error> || <detail>".
func newErrorWithDetail(err error, detail string) string {
	return ufmt.Errorf("%s || %s", err.Error(), detail).Error()
}
