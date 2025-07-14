package router

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/common"
)

var PERCENTAGE_DENOMINATOR = i256.NewInt(100)

// QuoteConstraints defines the valid range for swap quote percentages
const (
	MaxQuotePercentage = 100
	MinQuotePercentage = 0
)

// ErrorMessages for DrySwapRoute operations
const (
	ErrOverflowResultAmountIn      = "overflow in resultAmountIn"
	ErrOverflowResultAmountOut     = "overflow in resultAmountOut"
	ErrUnknownSwapType             = "unknown swapType(%s)"
	ErrInvalidPositiveAmount       = "invalid amount(%s), must be positive"
	ErrInvalidZeroAmountLimit      = "invalid amountLimit(%s), should not be zero"
	ErrInvalidQuoteRange           = "quote(%d) must be %d~%d"
	ErrOverflowCalculateSwapAmount = "overflow in calculateSwapAmount"
)

// SwapProcessor handles the execution of swap operations
type SwapProcessor struct{}

// ProcessSwapAmount calculates the exact amount to swap based on quote percentage
func (p *SwapProcessor) ProcessSwapAmount(amountSpecified *i256.Int, quote int) (*i256.Int, error) {
	if quote < MinQuotePercentage || quote > MaxQuotePercentage {
		return nil, ufmt.Errorf(ErrInvalidQuoteRange, quote, MinQuotePercentage, MaxQuotePercentage)
	}

	toSwap := i256.Zero().Mul(amountSpecified, i256.NewInt(int64(quote)))
	if toSwap.IsOverflow() {
		return nil, ufmt.Errorf(ErrOverflowCalculateSwapAmount)
	}

	return new(i256.Int).Div(toSwap, PERCENTAGE_DENOMINATOR), nil
}

// ProcessSingleSwap handles a single-hop swap simulation
func (p *SwapProcessor) ProcessSingleSwap(route string, amountSpecified *i256.Int) (*u256.Uint, *u256.Uint, error) {
	input, output, fee := getDataForSinglePath(route)
	singleParams := SingleSwapParams{
		tokenIn:         input,
		tokenOut:        output,
		fee:             fee,
		amountSpecified: amountSpecified,
	}

	amountIn, amountOut := singleDrySwap(&singleParams)
	return amountIn, amountOut, nil
}

// ProcessMultiSwap handles a multi-hop swap simulation
func (p *SwapProcessor) ProcessMultiSwap(
	swapType SwapType,
	route string,
	numHops int,
	amountSpecified *i256.Int,
) (*u256.Uint, *u256.Uint, error) {
	recipient := std.PreviousRealm().Address()
	pathIndex := getPathIndex(swapType, numHops)

	input, output, fee := getDataForMultiPath(route, pathIndex)
	swapParams := newSwapParams(input, output, fee, recipient, amountSpecified)

	switch swapType {
	case ExactIn:
		return multiDrySwap(*swapParams, numHops, route)
	case ExactOut:
		return multiDrySwapNegative(*swapParams, pathIndex, route)
	default:
		return nil, nil, ufmt.Errorf(ErrUnknownSwapType, swapType)
	}
}

// ValidateSwapResults checks if the swap results meet the required constraints
func (p *SwapProcessor) ValidateSwapResults(
	swapType SwapType,
	resultAmountIn, resultAmountOut *u256.Uint,
	amountSpecified, amountLimit *i256.Int,
) (string, string, bool) {
	switch swapType {
	case ExactIn:
		if i256.FromUint256(resultAmountIn).Gt(amountSpecified) {
			return resultAmountIn.ToString(), resultAmountOut.ToString(), false
		}
		if i256.FromUint256(resultAmountOut).Lt(amountLimit) {
			return resultAmountIn.ToString(), resultAmountOut.ToString(), false
		}
		return resultAmountIn.ToString(), resultAmountOut.ToString(), true

	case ExactOut:
		if i256.FromUint256(resultAmountOut).Lt(amountSpecified) {
			return resultAmountIn.ToString(), resultAmountOut.ToString(), false
		}
		if i256.FromUint256(resultAmountIn).Gt(amountLimit) {
			return resultAmountIn.ToString(), resultAmountOut.ToString(), false
		}
		return resultAmountIn.ToString(), resultAmountOut.ToString(), true

	default:
		// This should never happen since we validate the swap type earlier
		return "", "", false
	}
}

// AddSwapResults safely adds swap result amounts, checking for overflow
func (p *SwapProcessor) AddSwapResults(
	resultAmountIn, resultAmountOut, amountIn, amountOut *u256.Uint,
) (*u256.Uint, *u256.Uint, error) {
	newAmountIn := new(u256.Uint).Add(resultAmountIn, amountIn)
	if newAmountIn.IsOverflow() {
		return nil, nil, ufmt.Errorf(ErrOverflowResultAmountIn)
	}

	newAmountOut := new(u256.Uint).Add(resultAmountOut, amountOut)
	if newAmountOut.IsOverflow() {
		return nil, nil, ufmt.Errorf(ErrOverflowResultAmountOut)
	}

	return newAmountIn, newAmountOut, nil
}

// DrySwapRoute simulates a token swap route without actually executing the swap.
// It calculates the expected outcome based on the current state of liquidity pools.
//
// Parameters:
//   - inputToken: the symbol of the input token
//   - outputToken: the symbol of the output token
//   - specifiedAmount: the amount specified by the user (input or output, depending on swapKind)
//   - swapTypeStr: the type of swap ("exactIn" or "exactOut")
//   - strRouteArr: comma-separated swap route strings
//   - quoteArr: comma-separated quote percentages for each route
//   - tokenAmountLimit: a limit for token amount depending on swapKind
//
// Returns:
//   - amountInStr: the calculated input amount as a string
//   - amountOutStr: the calculated output amount as a string
//   - success: true if the simulated swap meets all constraints
func DrySwapRoute(
	inputToken, outputToken string,
	specifiedAmount string,
	swapTypeStr string,
	strRouteArr, quoteArr string,
	tokenAmountLimit string,
) (string, string, bool) {
	drySwapRouteWithCrossFn := func(cur realm) (string, string, bool) {
		return drySwapRoute(inputToken, outputToken, specifiedAmount, swapTypeStr, strRouteArr, quoteArr, tokenAmountLimit)
	}

	return drySwapRouteWithCrossFn(cross)
}

// drySwapRoute is a function for applying cross realm
func drySwapRoute(
	inputToken, outputToken string,
	specifiedAmount string,
	swapTypeStr string,
	strRouteArr, quoteArr string,
	tokenAmountLimit string,
) (string, string, bool) {
	common.MustRegistered(inputToken, outputToken)
	// initialize components
	validator := &SwapValidator{}
	processor := &SwapProcessor{}
	parser := &RouteParser{}

	// validate and parse inputs
	swapType, err := validator.swapType(swapTypeStr)
	if err != nil {
		panic(addDetailToError(errInvalidSwapType, err.Error()))
	}

	amountSpecified, err := validator.amount(specifiedAmount)
	if err != nil {
		panic(addDetailToError(errInvalidInput, err.Error()))
	}

	amountLimit, err := validator.amountLimit(tokenAmountLimit)
	if err != nil {
		panic(addDetailToError(errInvalidInput, err.Error()))
	}

	routes, quotes, err := parser.ParseRoutes(strRouteArr, quoteArr)
	if err != nil {
		panic(addDetailToError(errInvalidRoutesAndQuotes, err.Error()))
	}

	// adjust amount sign for exact out swaps
	if swapType == ExactOut {
		amountSpecified = i256.Zero().Neg(amountSpecified)
	}

	// initialize accumulators for swap results
	resultAmountIn, resultAmountOut := zero, zero

	// Process each route
	for i, route := range routes {
		// calculate the amount to swap for this route
		quoteValue, err := strconv.Atoi(quotes[i])
		if err != nil {
			panic(addDetailToError(errInvalidInput, err.Error()))
		}

		toSwap, err := processor.ProcessSwapAmount(amountSpecified, quoteValue)
		if err != nil {
			panic(addDetailToError(errInvalidInput, err.Error()))
		}

		// determine the number of hops and validate
		numHops := strings.Count(route, POOL_SEPARATOR) + 1
		assertHopsInRange(numHops)

		// execute the appropriate swap type
		var amountIn, amountOut *u256.Uint
		if numHops == 1 {
			amountIn, amountOut, err = processor.ProcessSingleSwap(route, toSwap)
		} else {
			amountIn, amountOut, err = processor.ProcessMultiSwap(swapType, route, numHops, toSwap)
		}

		if err != nil {
			panic(addDetailToError(errInvalidSwapType, err.Error()))
		}

		// update accumulated results
		resultAmountIn, resultAmountOut, err = processor.AddSwapResults(resultAmountIn, resultAmountOut, amountIn, amountOut)
		if err != nil {
			panic(addDetailToError(errInvalidInput, err.Error()))
		}
	}

	return processor.ValidateSwapResults(swapType, resultAmountIn, resultAmountOut, amountSpecified, amountLimit)
}

// Helper function to get path index based on swap type and number of hops
func getPathIndex(swapType SwapType, numHops int) int {
	switch swapType {
	case ExactIn:
		// first data for exact input swaps
		return 0
	case ExactOut:
		// last data for exact output swaps
		return numHops - 1
	default:
		panic("should not happen")
	}
}
