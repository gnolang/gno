package router

import (
	"std"

	"gno.land/p/gnoswap/consts"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func multiSwap(params SwapParams, currentPoolIndex, numPools int, swapPath string) (*u256.Uint, *u256.Uint) { // firstAmountIn, lastAmountOut
	firstAmountIn := u256.Zero()

	payer := getPrevAddr() // user

	for {
		var recipient std.Address
		currentPoolIndex++

		if currentPoolIndex < numPools {
			recipient = consts.ROUTER_ADDR
		} else {
			recipient = params.recipient
		}

		amountIn, amountOut := swapInner(
			params.amountSpecified,
			recipient,
			u256.Zero(),
			SwapCallbackData{
				params.tokenIn,
				params.tokenOut,
				params.fee,
				payer,
			},
		)

		if currentPoolIndex == 1 {
			firstAmountIn = amountIn
		}

		if currentPoolIndex >= numPools {
			return firstAmountIn, amountOut
		}

		payer = consts.ROUTER_ADDR

		nextInput, nextOutput, nextFee := getDataForMultiPath(swapPath, currentPoolIndex)
		params.tokenIn = nextInput
		params.tokenOut = nextOutput
		params.fee = nextFee
		params.amountSpecified = i256.FromUint256(amountOut)
	}
}

func multiSwapNegative(params SwapParams, numPools int, swapPath string) (*u256.Uint, *u256.Uint) { // firstAmountIn, lastAmountOut
	firstAmountIn := u256.Zero()

	swapInfo := []SingleSwapParams{}

	// CALCULATE BACKWARD INFO
	currentPoolIndex := numPools
	for currentPoolIndex >= 0 {
		// Dry-run
		amountIn, _ := singleDrySwap(
			SingleSwapParams{
				tokenIn:         params.tokenIn,
				tokenOut:        params.tokenOut,
				fee:             params.fee,
				amountSpecified: params.amountSpecified,
			},
		)

		thisSwap := SingleSwapParams{
			tokenIn:         params.tokenIn,
			tokenOut:        params.tokenOut,
			fee:             params.fee,
			amountSpecified: params.amountSpecified,
		}
		swapInfo = append(swapInfo, thisSwap)

		if currentPoolIndex == 0 {
			break
		}
		currentPoolIndex--

		nextInput, nextOutput, nextFee := getDataForMultiPath(swapPath, currentPoolIndex)

		intAmountIn := i256.FromUint256(amountIn)
		params.tokenIn = nextInput
		params.tokenOut = nextOutput
		params.fee = nextFee
		params.amountSpecified = i256.Zero().Neg(intAmountIn)
	}

	// PROCESS FORWARD INFO
	currentPoolIndex = len(swapInfo) - 1
	payer := getPrevAddr() // first payer ≈ user
	for currentPoolIndex >= 0 {
		var recipient std.Address
		if currentPoolIndex == 0 {
			recipient = params.recipient // params.recipient ≈ user
		} else {
			recipient = consts.ROUTER_ADDR
		}

		amountIn, amountOut := swapInner(
			swapInfo[currentPoolIndex].amountSpecified,
			recipient,
			u256.Zero(),
			SwapCallbackData{
				swapInfo[currentPoolIndex].tokenIn,
				swapInfo[currentPoolIndex].tokenOut,
				swapInfo[currentPoolIndex].fee,
				payer,
			},
		)

		// save route's first hop's amountIn to check whether crossed limit or not
		if currentPoolIndex == len(swapInfo)-1 {
			firstAmountIn = amountIn
		}

		if currentPoolIndex == 0 {
			return firstAmountIn, amountOut
		}

		swapInfo[currentPoolIndex-1].amountSpecified = i256.FromUint256(amountOut)
		payer = consts.ROUTER_ADDR
		currentPoolIndex--
	}
	return firstAmountIn, u256.Zero()
}

func multiDrySwap(params SwapParams, currentPoolIndex, numPool int, swapPath string) (*u256.Uint, *u256.Uint) { // firstAmountIn, lastAmountOut
	firstAmountIn := u256.Zero()

	payer := getPrevAddr() // user

	for {
		currentPoolIndex++

		amountIn, amountOut := swapDryInner(
			params.amountSpecified,
			u256.Zero(),
			SwapCallbackData{
				params.tokenIn,
				params.tokenOut,
				params.fee,
				payer,
			},
		)

		if currentPoolIndex == 1 {
			firstAmountIn = amountIn
		}

		if currentPoolIndex >= numPool {
			return firstAmountIn, amountOut
		}

		payer = consts.ROUTER_ADDR

		nextInput, nextOutput, nextFee := getDataForMultiPath(swapPath, currentPoolIndex)
		params.tokenIn = nextInput
		params.tokenOut = nextOutput
		params.fee = nextFee
		params.amountSpecified = i256.FromUint256(amountOut)
	}

}

func multiDrySwapNegative(params SwapParams, currentPoolIndex int, swapPath string) (*u256.Uint, *u256.Uint) {
	firstAmountIn := u256.Zero()
	payer := consts.ROUTER_ADDR

	for {
		amountIn, amountOut := swapDryInner(
			params.amountSpecified,
			u256.Zero(),
			SwapCallbackData{
				params.tokenIn,
				params.tokenOut,
				params.fee,
				payer,
			},
		)

		if currentPoolIndex == 0 {
			firstAmountIn = amountIn
		}

		currentPoolIndex--

		if currentPoolIndex == -1 {
			return firstAmountIn, amountOut
		}

		nextInput, nextOutput, nextFee := getDataForMultiPath(swapPath, currentPoolIndex)
		intAmountIn := i256.FromUint256(amountIn)

		params.amountSpecified = i256.Zero().Neg(intAmountIn)
		params.tokenIn = nextInput
		params.tokenOut = nextOutput
		params.fee = nextFee
	}
}
