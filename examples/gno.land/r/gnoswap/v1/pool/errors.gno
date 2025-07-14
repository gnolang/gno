package pool

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

// Error Messages
// These error messages are intended to inform you when an error occurs while using Contract and what it is.
var (
	errNoPermission              = errors.New("[GNOSWAP-POOL-001] caller has no permission")
	errUnsupportedFeeTier        = errors.New("[GNOSWAP-POOL-002] unsupported fee tier")
	errPoolAlreadyExists         = errors.New("[GNOSWAP-POOL-003] pool already created")
	errInvalidTickMinMaxRange    = errors.New("[GNOSWAP-POOL-004] tickLower and tickUpper are not within the valid range")
	errOutOfRange                = errors.New("[GNOSWAP-POOL-005] out of range for numeric value")
	errInvalidInput              = errors.New("[GNOSWAP-POOL-006] invalid input data")
	errInvalidPositionKey        = errors.New("[GNOSWAP-POOL-007] invalid position key")
	errDataNotFound              = errors.New("[GNOSWAP-POOL-008] requested data not found")
	errLiquidityCalculation      = errors.New("[GNOSWAP-POOL-009] invalid liquidity calculated")
	errZeroLiquidity             = errors.New("[GNOSWAP-POOL-010] zero liquidity")
	errDuplicateTokenInPool      = errors.New("[GNOSWAP-POOL-011] same token used in single pool")
	errTokenSortOrder            = errors.New("[GNOSWAP-POOL-012] tokens must be in lexicographical order")
	errTickLowerInvalid          = errors.New("[GNOSWAP-POOL-013] tickLower is invalid")
	errTickUpperInvalid          = errors.New("[GNOSWAP-POOL-014] tickUpper is invalid")
	errInvalidSwapAmount         = errors.New("[GNOSWAP-POOL-015] invalid swap amount")
	errInvalidProtocolFeePct     = errors.New("[GNOSWAP-POOL-016] invalid protocol fee percentage")
	errInvalidWithdrawalFeePct   = errors.New("[GNOSWAP-POOL-017] invalid withdrawal fee percentage")
	errLockedPool                = errors.New("[GNOSWAP-POOL-018] can't swap while pool is locked")
	errPriceOutOfRange           = errors.New("[GNOSWAP-POOL-019] swap price out of range")
	errMustBeNegative            = errors.New("[GNOSWAP-POOL-020] negative value expected")
	errTransferFailed            = errors.New("[GNOSWAP-POOL-021] token transfer failed")
	errInvalidTickAndTickSpacing = errors.New("[GNOSWAP-POOL-022] invalid tick and tick spacing requested")
	errInvalidAddress            = errors.New("[GNOSWAP-POOL-023] invalid address")
	errInvalidTickRange          = errors.New("[GNOSWAP-POOL-024] tickLower is greater than or equal to tickUpper")
	errUnderflow                 = errors.New("[GNOSWAP-POOL-025] underflow")
	errOverFlow                  = errors.New("[GNOSWAP-POOL-026] overflow")
	errBalanceUpdateFailed       = errors.New("[GNOSWAP-POOL-027] balance update failed")
)

// newErrorWithDetail adds detail to an error message
//
// input
// err: error - the error message that is one of above list
// detail: string - the detail to add to the error message
// Returns: string - the final error message
func newErrorWithDetail(err error, detail string) string {
	finalErr := ufmt.Errorf("%s || %s", err.Error(), detail)
	return finalErr.Error()
}
