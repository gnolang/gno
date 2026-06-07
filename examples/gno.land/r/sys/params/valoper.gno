package params

import (
	prms "sys/params"
)

// Valoper sys-param keys consumed by r/gnops/valopers (Register +
// UpdateSigningKey). Governance can update them via the generic
// NewSysParam*PropRequest factories — no realm-side gate is needed
// because the values only affect fees and throttle inside valopers.
const (
	valoperSubmodule = "valoper"

	valoperRegisterFeeKey          = "register_fee"
	valoperRotationFeeKey          = "rotation_fee"
	valoperRotationPeriodBlocksKey = "rotation_period_blocks"
)

// Default values used when the sys-param has never been set by
// governance. Zero fees (GNOT transfers are disabled chain-wide
// pre-fork) and a ~1-hour throttle at 6s/block.
const (
	defaultValoperRegisterFee          = uint64(0)
	defaultValoperRotationFee          = uint64(0)
	defaultValoperRotationPeriodBlocks = int64(600)
)

// GetValoperRegisterFee returns the fee (in ugnot) required to call
// valopers.Register. Defaults to 0 if governance hasn't set it.
func GetValoperRegisterFee() uint64 {
	v, ok := prms.GetSysParamUint64(nodeModulePrefix, valoperSubmodule, valoperRegisterFeeKey)
	if !ok {
		return defaultValoperRegisterFee
	}
	return v
}

// GetValoperRotationFee returns the fee (in ugnot) required to call
// valopers.UpdateSigningKey. Defaults to 0.
func GetValoperRotationFee() uint64 {
	v, ok := prms.GetSysParamUint64(nodeModulePrefix, valoperSubmodule, valoperRotationFeeKey)
	if !ok {
		return defaultValoperRotationFee
	}
	return v
}

// GetValoperRotationPeriodBlocks returns the per-operator rotation
// throttle (in blocks). Defaults to ~1h worth at 6s/block (600).
// This is the primary anti-spam defense pre-fee while rotation_fee
// stays at 0; tightens further once non-zero fees become enforceable.
func GetValoperRotationPeriodBlocks() int64 {
	v, ok := prms.GetSysParamInt64(nodeModulePrefix, valoperSubmodule, valoperRotationPeriodBlocksKey)
	if !ok {
		return defaultValoperRotationPeriodBlocks
	}
	return v
}
