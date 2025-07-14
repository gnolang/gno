package pool

import (
	"std"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
	pf "gno.land/r/gnoswap/v1/protocol_fee"
)

var (
	// poolCreationFee is the fee that is charged when a user creates a pool.
	// The fee is denominated in GNS tokens.
	poolCreationFee = int64(100_000_000) // 100_GNS

	// withdrawalFeeBPS is the fee that is charged when a user withdraws their collected fees
	// The fee is denominated in BPS (Basis Points)
	// Example: 100 BPS = 1%
	withdrawalFeeBPS = uint64(100) // 1%
)

const (
	MaxBpsValue = uint64(10000)
	ZeroBps     = uint64(0)
)

// CollectWithdrawalFee withdraws the fee from the user and returns the amount after the fee
// Only position contract can call this function
// Input:
// - positionId: the id of the LP token
// - token0Path: the path of the token0
// - amount0: the amount of token0
// - token1Path: the path of the token1
// - amount1: the amount of token1
// - poolPath: the path of the pool
// - positionCaller: the original caller of the position contract
// Output:
// - the amount of token0 after the fee
// - the amount of token1 after the fee
//
// ref: https://docs.gnoswap.io/contracts/pool/protocol_fee_withdrawal.gno#handlewithdrawalfee
func CollectWithdrawalFee(cur realm,
	positionId uint64,
	token0Path string,
	amount0 string, // uint256
	token1Path string,
	amount1 string, // uint256
	poolPath string,
	positionCaller std.Address,
) (string, string) { // uint256 x2
	assertOnlyNotHalted()
	common.MustRegistered(token0Path, token1Path)

	// only position contract can call this function
	caller := getPrevAddr()
	if err := common.PositionOnly(caller); err != nil {
		panic(newErrorWithDetail(
			errNoPermission,
			ufmt.Sprintf("only position(%s) can call this function, called from %s", positionAddr, caller.String()),
		))
	}

	fee := GetWithdrawalFee()
	if fee == ZeroBps {
		return amount0, amount1
	}

	feeAmount0, afterAmount0 := calculateAmountWithFee(u256.MustFromDecimal(amount0), u256.NewUint(fee))
	feeAmount1, afterAmount1 := calculateAmountWithFee(u256.MustFromDecimal(amount1), u256.NewUint(fee))

	checkTransferError(common.TransferFrom(cross, token0Path, positionCaller, protocolFeeAddr, feeAmount0.Int64()))
	pf.AddToProtocolFee(cross, token0Path, feeAmount0.Uint64())
	checkTransferError(common.TransferFrom(cross, token1Path, positionCaller, protocolFeeAddr, feeAmount1.Int64()))
	pf.AddToProtocolFee(cross, token1Path, feeAmount1.Uint64())

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"WithdrawalFee",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"lpTokenId", formatUint(positionId),
		"poolPath", poolPath,
		"feeAmount0", feeAmount0.ToString(),
		"feeAmount1", feeAmount1.ToString(),
		"amount0WithoutFee", afterAmount0.ToString(),
		"amount1WithoutFee", afterAmount1.ToString(),
	)

	return afterAmount0.ToString(), afterAmount1.ToString()
}

// GetPoolCreationFee returns the poolCreationFee
// ref: https://docs.gnoswap.io/contracts/pool/protocol_fee_pool_creation.gno#getpoolcreationfee
func GetPoolCreationFee() int64 {
	return poolCreationFee
}

// SetPoolCreationFee sets the poolCreationFee.
// Only governance contract can execute this function via proposal
// ref: https://docs.gnoswap.io/contracts/pool/protocol_fee_pool_creation.gno#setpoolcreationfee
func SetPoolCreationFee(cur realm, fee int64) {
	assertOnlyNotHalted()
	caller := getPrevAddr()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err.Error())
	}
	prevPoolCreationFee := GetPoolCreationFee()
	setPoolCreationFee(fee)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"SetPoolCreationFee",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevFee", formatInt(prevPoolCreationFee),
		"newFee", formatInt(fee),
	)
}

// SetPoolCreationFeeByAdmin sets the poolCreationFee by Admin.
// Only admin can execute this function.
func SetPoolCreationFeeByAdmin(cur realm, fee int64) {
	assertOnlyNotHalted()
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err.Error())
	}
	prevPoolCreationFee := GetPoolCreationFee()
	setPoolCreationFee(fee)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"SetPoolCreationFeeByAdmin",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevFee", formatInt(prevPoolCreationFee),
		"newFee", formatInt(fee),
	)
}

// GetWithdrawalFee returns the withdrawal fee
// ref: https://docs.gnoswap.io/contracts/pool/protocol_fee_withdrawal.gno#getwithdrawalfee
func GetWithdrawalFee() uint64 {
	return withdrawalFeeBPS
}

// SetWithdrawalFee sets the withdrawal fee.
// Only governance contract can execute this function via proposal
// ref: https://docs.gnoswap.io/contracts/pool/protocol_fee_withdrawal.gno#setwithdrawalfee
func SetWithdrawalFee(cur realm, fee uint64) {
	assertOnlyNotHalted()
	caller := getPrevAddr()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err.Error())
	}

	prevWithdrawalFee := GetWithdrawalFee()

	setWithdrawalFee(fee)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"SetWithdrawalFee",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevFee", formatUint(prevWithdrawalFee),
		"newFee", formatUint(fee),
	)
}

// SetWithdrawalFeeByAdmin sets the withdrawal fee by Admin.
// Only admin can execute this function.
func SetWithdrawalFeeByAdmin(cur realm, fee uint64) {
	assertOnlyNotHalted()
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err.Error())
	}

	prevWithdrawalFee := GetWithdrawalFee()

	setWithdrawalFee(fee)

	prevAddr, prevPkgPath := getPrevAsString()
	std.Emit(
		"SetWithdrawalFeeByAdmin",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"prevFee", formatUint(prevWithdrawalFee),
		"newFee", formatUint(fee),
	)
}

// calculateAmountWithFee calculates the fee amount and the amount after the fee
//
// Inputs:
// - amount: the amount before the fee
// - fee: the fee in BPS
//
// Outputs:
// - the fee amount
// - the amount after the fee applied
func calculateAmountWithFee(amount *u256.Uint, fee *u256.Uint) (*u256.Uint, *u256.Uint) {
	feeAmount := new(u256.Uint).Mul(amount, fee)
	feeAmount = new(u256.Uint).Div(feeAmount, u256.NewUint(MaxBpsValue))
	afterAmount := new(u256.Uint).Sub(amount, feeAmount)
	return feeAmount, afterAmount
}

// setPoolCreationFee this function is internal function called by SetPoolCreationFee
// And SetPoolCreationFeeByAdmin
func setPoolCreationFee(fee int64) {
	if fee < 0 {
		panic("pool creation fee cannot be negative")
	}
	poolCreationFee = fee
}

// setWithdrawalFee this function is internal function called by SetWithdrawalFee
// function and SetWithdrawalFeeByAdmin function
func setWithdrawalFee(fee uint64) {
	// 10000 (bps) = 100%
	if fee > MaxBpsValue {
		panic(newErrorWithDetail(
			errInvalidWithdrawalFeePct,
			ufmt.Sprintf("fee(%d) must be in range 0 ~ 10000", fee),
		))
	}
	withdrawalFeeBPS = fee
}
