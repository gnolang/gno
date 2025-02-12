package launchpad

import (
	"errors"

	"gno.land/p/demo/ufmt"
)

var (
	errNoLeftReward        = errors.New("[GNOSWAP-LAUNCHPAD-001] no left reward")
	errInvalidAddress      = errors.New("[GNOSWAP-LAUNCHPAD-002] invalid address")
	errDataNotFound        = errors.New("[GNOSWAP-LAUNCHPAD-003] requested data not found")
	errActiveProject       = errors.New("[GNOSWAP-LAUNCHPAD-004] project is active")
	errInactiveProject     = errors.New("[GNOSWAP-LAUNCHPAD-005] project is inactive")
	errInactiveTier        = errors.New("[GNOSWAP-LAUNCHPAD-006] pool is inactive")
	errInvalidInput        = errors.New("[GNOSWAP-LAUNCHPAD-007] invalid input data")
	errDuplicateProject    = errors.New("[GNOSWAP-LAUNCHPAD-008] can not create same project in same block")
	errInvalidTier         = errors.New("[GNOSWAP-LAUNCHPAD-009] invalid pool")
	errInvalidTierRatio    = errors.New("[GNOSWAP-LAUNCHPAD-010] invalid pool ratio")
	errInvalidLength       = errors.New("[GNOSWAP-LAUNCHPAD-011] invalid length")
	errNotEnoughBalance    = errors.New("[GNOSWAP-LAUNCHPAD-012] not enough balance")
	errInvalidCondition    = errors.New("[GNOSWAP-LAUNCHPAD-013] invalid transfer condition")
	errConvertFail         = errors.New("[GNOSWAP-LAUNCHPAD-014] convert fail")
	errNotUserCaller       = errors.New("[GNOSWAP-LAUNCHPAD-015] only user caller")
	errInvalidData         = errors.New("[GNOSWAP-LAUNCHPAD-016] invalid data")
	errInvalidAmount       = errors.New("[GNOSWAP-LAUNCHPAD-017] invalid amount")
	errDuplicateDeposit    = errors.New("[GNOSWAP-LAUNCHPAD-018] duplicate deposit")
	errInvalidRewardState  = errors.New("[GNOSWAP-LAUNCHPAD-019] invalid reward state")
	errNotExistDeposit     = errors.New("[GNOSWAP-LAUNCHPAD-020] not exist deposit")
	errAlreadyExistDeposit = errors.New("[GNOSWAP-LAUNCHPAD-021] already exist deposit")
	errInvalidProjectId    = errors.New("[GNOSWAP-LAUNCHPAD-022] invalid project id")
	errAlreadyCollected    = errors.New("[GNOSWAP-LAUNCHPAD-023] already collected")
	errNotYetClaimReward   = errors.New("[GNOSWAP-LAUNCHPAD-024] not yet claim reward")
	errInvalidCaller       = errors.New("[GNOSWAP-LAUNCHPAD-025] invalid caller")
	errInvalidOwner        = errors.New("[GNOSWAP-LAUNCHPAD-026] invalid owner")
	errInvalidAvgBlockTime = errors.New("[GNOSWAP-LAUNCHPAD-027] invalid average block time")
	errInvalidHeight       = errors.New("[GNOSWAP-LAUNCHPAD-028] invalid height")
	errTierHasParticipants = errors.New("[GNOSWAP-LAUNCHPAD-029] tier has participants")
)

func addDetailToError(err error, detail string) string {
	finalErr := ufmt.Errorf("%s || %s", err.Error(), detail)
	return finalErr.Error()
}
