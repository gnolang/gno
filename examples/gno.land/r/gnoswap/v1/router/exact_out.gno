package router

import (
	"std"

	"gno.land/p/demo/ufmt"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/referral"
)

type ExactOutSwapOperation struct {
	baseSwapOperation
	params ExactOutParams
}

func NewExactOutSwapOperation(pp ExactOutParams) *ExactOutSwapOperation {
	return &ExactOutSwapOperation{
		params: pp,
		baseSwapOperation: baseSwapOperation{
			userWrappedWugnot: INITIAL_WUGNOT_BALANCE,
		},
	}
}

// ExactOutSwapRoute is a function that performs an exact-out swap operation.
// exact-out swap means that the user wants to swap an exact amount of output tokens for input tokens.
func ExactOutSwapRoute(cur realm,
	inputToken string,
	outputToken string,
	amountOut string,
	routeArr string,
	quoteArr string,
	amountInMax string,
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
		RouteArr:    routeArr,
		QuoteArr:    quoteArr,
	}

	pp := NewExactOutParams(baseParams, amountOut, amountInMax)
	op := NewExactOutSwapOperation(pp)

	result, err := executeSwapOperation(op)
	if err != nil {
		panic(addDetailToError(errInvalidInput, err.Error()))
	}

	inputAmount, outputAmount := finalizeSwap(
		pp.InputToken,
		pp.OutputToken,
		result.AmountIn,
		result.AmountOut,
		ExactOut,
		u256.MustFromDecimal(pp.AmountInMax),
		op.userBeforeWugnotBalance,
		op.userWrappedWugnot,
		result.AmountSpecified.Abs(),
	)

	prevAddr, prevPkgPath := getPrev()
	success := referral.TryRegister(cross, std.Address(prevAddr), referrer)
	actualReferrer := referrer
	if !success {
		actualReferrer = referral.GetReferral(prevAddr)
	}
	std.Emit(
		"ExactOutSwap",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"input", pp.InputToken,
		"output", pp.OutputToken,
		"exactAmount", amountOut,
		"route", pp.RouteArr,
		"quote", pp.QuoteArr,
		"resultInputAmount", inputAmount,
		"resultOutputAmount", outputAmount,
		"referrer", actualReferrer,
	)

	return inputAmount, outputAmount
}

func (op *ExactOutSwapOperation) Validate() error {
	amountOut := i256.MustFromDecimal(op.params.AmountOut)
	if amountOut.IsZero() || amountOut.IsNeg() {
		return ufmt.Errorf("invalid amountOut(%s), must be positive", amountOut.ToString())
	}

	// assign a signed reversed `amountOut` to `amountSpecified`
	// when it's an ExactOut
	op.amountSpecified = i256.Zero().Neg(amountOut)

	routes, quotes, err := validateRoutesAndQuotes(op.params.RouteArr, op.params.QuoteArr)
	if err != nil {
		return err
	}

	op.routes = routes
	op.quotes = quotes

	return nil
}

func (op *ExactOutSwapOperation) Process() (*SwapResult, error) {
	if err := op.handleNativeTokenWrapping(); err != nil {
		return nil, err
	}

	resultAmountIn, resultAmountOut, err := op.processRoutes(ExactOut)
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

func (op *ExactOutSwapOperation) handleNativeTokenWrapping() error {
	return op.baseSwapOperation.handleNativeTokenWrapping(
		op.params.InputToken,
		op.params.OutputToken,
		i256.MustFromDecimal(op.params.AmountInMax),
	)
}
