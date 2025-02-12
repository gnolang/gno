package staker

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	errNoPermission                = errors.New("[GNOSWAP-STAKER-001] caller has no permission")
	errPoolNotFound                = errors.New("[GNOSWAP-STAKER-002] pool not found")
	errAlreadyRegistered           = errors.New("[GNOSWAP-STAKER-003] already registered token")
	errInsufficientReward          = errors.New("[GNOSWAP-STAKER-004] insufficient reward")
	errWrapUnwrap                  = errors.New("[GNOSWAP-STAKER-005] wrap, unwrap failed")
	errWugnotMinimum               = errors.New("[GNOSWAP-STAKER-006] can not wrapless than minimum amount")
	errInvalidInput                = errors.New("[GNOSWAP-STAKER-007] invalid input data")
	errInvalidUnstakingFee         = errors.New("[GNOSWAP-STAKER-008] invalid unstaking fee")
	errAlreadyStaked               = errors.New("[GNOSWAP-STAKER-009] already staked position")
	errNonIncentivizedPool         = errors.New("[GNOSWAP-STAKER-010] pool is not incentivized")
	errOutOfRange                  = errors.New("[GNOSWAP-STAKER-011] out of range")
	errCannotEndIncentive          = errors.New("[GNOSWAP-STAKER-012] can not end incentive")
	errInvalidIncentiveStartTime   = errors.New("[GNOSWAP-STAKER-013] invalid incentive start time")
	errInvalidIncentiveEndTime     = errors.New("[GNOSWAP-STAKER-014] invalid incentive end time")
	errCannotUseForExternalReward  = errors.New("[GNOSWAP-STAKER-015] can not use for external reward")
	errMinTier                     = errors.New("[GNOSWAP-STAKER-016] emission minimum tier is 1")
	errDefaultPoolTier1            = errors.New("[GNOSWAP-STAKER-017] can not delete default pool tier 1")
	errDefaultExternalToken        = errors.New("[GNOSWAP-STAKER-018] can not delete default external token")
	errInvalidPoolPath             = errors.New("[GNOSWAP-STAKER-019] invalid pool path")
	errInvalidPoolTier             = errors.New("[GNOSWAP-STAKER-020] invalid pool tier")
	errAlreadyHasTier              = errors.New("[GNOSWAP-STAKER-021] pool already has emission target")
	errDataNotFound                = errors.New("[GNOSWAP-STAKER-022] requested data not found")
	errCalculationError            = errors.New("[GNOSWAP-STAKER-023] unexpected calculation error")
	errZeroLiquidity               = errors.New("[GNOSWAP-STAKER-024] zero liquidity")
	errInvalidIncentiveDuration    = errors.New("[GNOSWAP-STAKER-025] invalid incentive duration")
	errNotAllowedForExternalReward = errors.New("[GNOSWAP-STAKER-026] not allowed for external reward")
	errInvalidWarmUpPercent        = errors.New("[GNOSWAP-STAKER-027] invalid warm-up duration")
	errInvalidTickCross            = errors.New("[GNOSWAP-STAKER-028] invalid tick cross")
	errIncentiveAlreadyExists      = errors.New("[GNOSWAP-STAKER-029] incentive already exists")
	errIncentiveNotFound           = errors.New("[GNOSWAP-STAKER-030] incentive not found")
	errWarmUpAmountNotFound        = errors.New("[GNOSWAP-STAKER-031] warm-up amount not found")
	errOverflow                    = errors.New("[GNOSWAP-STAKER-032] overflow")
)

func addDetailToError(err error, detail string) string {
	finalErr := ufmt.Errorf("%s || %s", err.Error(), detail)
	return finalErr.Error()
}
