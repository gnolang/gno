package router

import (
	"std"

	"gno.land/p/demo/ufmt"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/referral"
)

type ExactInSwapOperation struct {
	baseSwapOperation
	params ExactInParams
}

func NewExactInSwapOperation(pp ExactInParams) *ExactInSwapOperation {
	return &ExactInSwapOperation{
		params: pp,
		baseSwapOperation: baseSwapOperation{
			userWrappedWugnot: INITIAL_WUGNOT_BALANCE,
		},
	}
}

// ExactInSwapRoute is a function that performs an exact-in swap operation.
// exact-in swap means that the user wants to swap an exact amount of input tokens for output tokens.
func ExactInSwapRoute(cur realm,
	inputToken string,
	outputToken string,
	amountIn string,
	RouteArr string,
	quoteArr string,
	amountOutMin string,
	deadline int64,
	referrer string,
) (string, string) {
	if err := checkDeadline(deadline); err != nil {
		panic(addDetailToError(errExpired, err.Error()))
	}

	if err := commonSwapSetup(); err != nil {
		panic(addDetailToError(errInvalidInput, err.Error()))
	}

	baseParams := BaseSwapParams{
		InputToken:  inputToken,
		OutputToken: outputToken,
		RouteArr:    RouteArr,
		QuoteArr:    quoteArr,
	}

	pp := NewExactInParams(
		baseParams,
		amountIn,
		amountOutMin,
	)

	op := NewExactInSwapOperation(pp)

	result, err := executeSwapOperation(op)
	if err != nil {
		panic(addDetailToError(
			errInvalidInput,
			ufmt.Sprintf("invalid ExactInSwapOperation: %s", err.Error()),
		))
	}

	inputAmount, outputAmount := finalizeSwap(
		pp.InputToken,
		pp.OutputToken,
		result.AmountIn,
		result.AmountOut,
		ExactIn,
		u256.MustFromDecimal(pp.AmountOutMin),
		op.userBeforeWugnotBalance,
		op.userWrappedWugnot,
		result.AmountSpecified.Abs(),
	)

	prevAddr, prevPkgPath := getPrev()
	success := referral.TryRegister(cross, std.Address(prevAddr), referrer)

	// get the actual referral address that was used
	actualReferrer := referrer
	if !success {
		actualReferrer = referral.GetReferral(prevAddr)
	}

	std.Emit(
		"ExactInSwap",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"input", pp.InputToken,
		"output", pp.OutputToken,
		"exactAmount", amountIn,
		"route", pp.RouteArr,
		"quote", pp.QuoteArr,
		"resultInputAmount", inputAmount,
		"resultOutputAmount", outputAmount,
		"referrer", actualReferrer,
	)

	return inputAmount, outputAmount
}

func (op *ExactInSwapOperation) Validate() error {
	amountIn := i256.MustFromDecimal(op.params.AmountIn)
	if amountIn.IsZero() || amountIn.IsNeg() {
		return ufmt.Errorf("invalid amountIn(%s), must be positive", amountIn.ToString())
	}

	// when `SwapType` is `ExactIn`, assign `amountSpecified` the `amountIn`
	// obtained from above.
	op.amountSpecified = amountIn

	routes, quotes, err := validateRoutesAndQuotes(op.params.RouteArr, op.params.QuoteArr)
	if err != nil {
		return err
	}

	op.routes = routes
	op.quotes = quotes

	return nil
}

func (op *ExactInSwapOperation) Process() (*SwapResult, error) {
	if err := op.handleNativeTokenWrapping(); err != nil {
		return nil, err
	}

	resultAmountIn, resultAmountOut, err := op.processRoutes(ExactIn)
	if err != nil {
		return nil, err
	}

	return &SwapResult{
		AmountIn:        resultAmountIn,
		AmountOut:       resultAmountOut,
		Routes:          op.routes,
		Quotes:          op.quotes,
		AmountSpecified: op.amountSpecified,
	}, nil
}

func (op *ExactInSwapOperation) handleNativeTokenWrapping() error {
	return op.baseSwapOperation.handleNativeTokenWrapping(
		op.params.InputToken,
		op.params.OutputToken,
		op.amountSpecified,
	)
}
