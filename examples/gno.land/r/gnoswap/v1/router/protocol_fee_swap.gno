package router

import (
	"std"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/p/demo/ufmt"

	u256 "gno.land/p/gnoswap/uint256"

	pf "gno.land/r/gnoswap/v1/protocol_fee"
)

const (
	defaultSwapFeeBPS = uint64(15) // 0.15%
)

var (
	swapFee = defaultSwapFeeBPS
)

// GetSwapFee returns current rate of swap fee
// ref: https://docs.gnoswap.io/contracts/router/protocol_fee_swap.gno#getswapfee
func GetSwapFee() uint64 {
	return swapFee
}

// SetSwapFeeByAdmin modifies the swap fee
// Only admin can execute this function
func SetSwapFeeByAdmin(fee uint64) {
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err)
	}

	prevSwapFee := swapFee
	if err := setSwapFee(fee); err != nil {
		panic(err)
	}

	prevAddr, prevPkgPath := getPrev()

	std.Emit(
		"SetSwapFeeByAdmin",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"newFee", formatUint(fee),
		"prevFee", formatUint(prevSwapFee),
	)
}

// SetSwapFee modifies the swap fee
// Only governance contract can execute this function via proposal
// ref: https://docs.gnoswap.io/contracts/router/protocol_fee_swap.gno#setswapfee
func SetSwapFee(fee uint64) {
	caller := getPrevAddr()
	if err := common.GovernanceOnly(caller); err != nil {
		panic(err)
	}

	prevSwapFee := swapFee
	if err := setSwapFee(fee); err != nil {
		panic(err)
	}

	prevAddr, prevPkgPath := getPrev()

	std.Emit(
		"SetSwapFee",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"newFee", formatUint(fee),
		"prevFee", formatUint(prevSwapFee),
	)
}

func setSwapFee(fee uint64) error {
	common.IsHalted()

	// 10000 (bps) = 100%
	if fee > 10000 {
		return ufmt.Errorf(
			"%s: fee must be in range 0 to 10000. got %d",
			errInvalidSwapFee.Error(), fee,
		)
	}

	swapFee = fee
	return nil
}

// handleSwapFee deducts the protocol fee from the amount and sends it to the protocol_fee contract
// and returns the amount after deducting the protocol fee
func handleSwapFee(
	outputToken string,
	amount *u256.Uint,
) *u256.Uint {
	if swapFee <= 0 {
		return amount
	}

	feeAmount := new(u256.Uint).Mul(amount, u256.NewUint(swapFee))
	feeAmount.Div(feeAmount, u256.NewUint(10000))
	feeAmountUint64 := feeAmount.Uint64()

	outputTeller := common.GetTokenTeller(outputToken)
	outputTeller.TransferFrom(std.PrevRealm().Addr(), consts.PROTOCOL_FEE_ADDR, feeAmountUint64)
	pf.AddToProtocolFee(outputToken, feeAmountUint64)

	prevAddr, prevPkgPath := getPrev()

	std.Emit(
		"SwapRouteFee",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"tokenPath", outputToken,
		"amount", formatUint(feeAmountUint64),
	)

	toUserAfterProtocol := new(u256.Uint).Sub(amount, feeAmount)
	return toUserAfterProtocol
}
