package router

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/demo/wugnot"
)

var one = u256.One()

// ErrorMessages define all error message templates used throughout the router
const (
	// slippage validation
	errExactOutAmountExceeded = "Received more than requested in [EXACT_OUT] requested=%s, actual=%s"

	// route validation
	errInvalidRouteLength = "route length(%d) must be 1~7"

	// quote validation
	errRoutesQuotesMismatch = "mismatch between routes(%d) and quotes(%d) length"
	errInvalidQuote         = "invalid quote(%s) at index(%d)"
	errInvalidQuoteValue    = "quote(%s) at index(%d) must be positive value"
	errQuoteExceedsMax      = "quote(%s) at index(%d) must be less than or equal to %d"
	errQuoteSumExceedsMax   = "quote sum exceeds 100 at index(%d)"
	errInvalidQuoteSum      = "quote sum(%d) must be 100"
)

// GnotSwapHandler encapsulates methods for handling GNOT token swaps
type GnotSwapHandler struct {
	BeforeBalance int64
	WrappedAmount int64
	NewBalance    int64
}

// newGnotSwapHandler creates a new handler for GNOT swaps
func newGnotSwapHandler(beforeBalance, wrappedAmount int64) *GnotSwapHandler {
	return &GnotSwapHandler{
		BeforeBalance: beforeBalance,
		WrappedAmount: wrappedAmount,
	}
}

// UpdateNewBalance updates the current balance after swap operations
func (h *GnotSwapHandler) UpdateNewBalance() {
	h.NewBalance = wugnot.BalanceOf(std.PreviousRealm().Address())
}

// HandleInputSwap manages unwrapping logic for GNOT input tokens
// Returns error if validation fails
func (h *GnotSwapHandler) HandleInputSwap() error {
	totalBefore := h.BeforeBalance + h.WrappedAmount
	spend := totalBefore - h.NewBalance

	if spend > h.WrappedAmount {
		return ufmt.Errorf("too much wugnot spent (wrapped: %d, spend: %d)",
			h.WrappedAmount, spend)
	}

	toUnwrap := h.WrappedAmount - spend
	unwrap(toUnwrap)
	return nil
}

// HandleOutputSwap manages unwrapping logic for GNOT output tokens
func (h *GnotSwapHandler) HandleOutputSwap() {
	userRecvWugnot := h.NewBalance - h.BeforeBalance - h.WrappedAmount
	unwrap(userRecvWugnot)
}

// SwapValidator provides validation methods for swap operations
type SwapValidator struct{}

// exactOutAmount checks if output amount meets specified requirements
func (v *SwapValidator) exactOutAmount(resultAmount, specifiedAmount *u256.Uint) error {
	if resultAmount.Gte(specifiedAmount) {
		return nil
	}

	diff := u256.Zero().Sub(specifiedAmount, resultAmount)
	if diff.Gt(one) {
		return ufmt.Errorf(errExactOutAmountExceeded, specifiedAmount.ToString(), resultAmount.ToString())
	}
	return nil
}

// slippage ensures swap amounts meet slippage requirements
func (v *SwapValidator) slippage(swapType SwapType, amountIn, amountOut, limit *u256.Uint) error {
	switch swapType {
	case ExactIn:
		if amountOut.Lt(limit) {
			return ufmt.Errorf("ExactIn: too few received (min:%s, got:%s)",
				limit.ToString(), amountOut.ToString())
		}
	case ExactOut:
		if amountIn.Gt(limit) {
			return ufmt.Errorf("ExactOut: too much spent (max:%s, used:%s)",
				limit.ToString(), amountIn.ToString())
		}
	default:
		return errInvalidSwapType
	}
	return nil
}

// swapType ensures the swap type string is valid
func (v *SwapValidator) swapType(swapTypeStr string) (SwapType, error) {
	swapType, err := trySwapTypeFromStr(swapTypeStr)
	if err != nil {
		return Unknown, errInvalidSwapType
	}
	return swapType, nil
}

// amount ensures the amount is properly formatted and positive
func (v *SwapValidator) amount(amount string) (*i256.Int, error) {
	parsedAmount := i256.MustFromDecimal(amount)
	if parsedAmount.Lt(i256.Zero()) {
		return nil, ufmt.Errorf(ErrInvalidPositiveAmount, amount)
	}
	return parsedAmount, nil
}

// amountLimit ensures the amount limit is properly formatted and non-zero
func (v *SwapValidator) amountLimit(amountLimit string) (*i256.Int, error) {
	parsedLimit := i256.MustFromDecimal(amountLimit)
	if parsedLimit.IsZero() {
		return nil, ufmt.Errorf(ErrInvalidZeroAmountLimit, amountLimit)
	}
	return parsedLimit, nil
}

// RouteParser handles parsing and validation of routes and quotes
type RouteParser struct{}

func NewRouteParser() *RouteParser {
	return &RouteParser{}
}

// ParseRoutes parses route and quote strings into slices and validates them
func (p *RouteParser) ParseRoutes(routes, quotes string) ([]string, []string, error) {
	routesArr := splitSingleChar(routes, ',')
	quotesArr := splitSingleChar(quotes, ',')

	if err := p.ValidateRoutesAndQuotes(routesArr, quotesArr); err != nil {
		return nil, nil, err
	}

	return routesArr, quotesArr, nil
}

// ValidateRoutesAndQuotes ensures routes and quotes meet required criteria
func (p *RouteParser) ValidateRoutesAndQuotes(routes, quotes []string) error {
	rr := len(routes)
	qq := len(quotes)

	if rr < 1 || rr > 7 {
		return ufmt.Errorf(errInvalidRouteLength, rr)
	}

	if rr != qq {
		return ufmt.Errorf(errRoutesQuotesMismatch, rr, qq)
	}

	return p.ValidateQuoteSum(quotes)
}

// ValidateQuoteSum ensures all quotes add up to 100%
func (p *RouteParser) ValidateQuoteSum(quotes []string) error {
	const (
		maxQuote int8 = 100
		minQuote int8 = 0
	)

	var sum int8

	for i, quote := range quotes {
		qt, err := strconv.ParseInt(quote, 10, 8)
		if err != nil {
			return ufmt.Errorf(errInvalidQuote, quote, i)
		}
		intQuote := int8(qt)

		// Even if quoteArr itself contains 0, there's no problem as long as the sum equals 100,
		// but since quote generally won't be 0, we check if it's less than or equal to minQuote.
		if intQuote <= minQuote {
			return ufmt.Errorf(errInvalidQuoteValue, quote, i)
		}

		if intQuote > maxQuote {
			return ufmt.Errorf(errQuoteExceedsMax, quote, i, maxQuote)
		}

		if sum > maxQuote-intQuote {
			return ufmt.Errorf(errQuoteSumExceedsMax, i)
		}

		sum += intQuote
	}

	if sum != maxQuote {
		return ufmt.Errorf(errInvalidQuoteSum, sum)
	}

	return nil
}

// FinalizeSwap handles post-swap operations and validations
// Returns input and output amount strings after processing
func finalizeSwap(
	inputToken, outputToken string,
	resultAmountIn, resultAmountOut *u256.Uint,
	swapType SwapType,
	tokenAmountLimit *u256.Uint,
	userBeforeWugnotBalance, userWrappedWugnot int64,
	amountSpecified *u256.Uint,
) (string, string) {
	validator := &SwapValidator{}

	// Validate exact out amount if applicable
	if swapType == ExactOut {
		if err := validator.exactOutAmount(resultAmountOut, amountSpecified); err != nil {
			panic(addDetailToError(errSlippage, err.Error()))
		}
	}

	// Handle swap fee
	resultAmountOutWithoutFee := handleSwapFee(outputToken, resultAmountOut)

	// Handle GNOT token swaps
	handler := newGnotSwapHandler(userBeforeWugnotBalance, userWrappedWugnot)
	handler.UpdateNewBalance()

	var err error
	if inputToken == consts.GNOT {
		err = handler.HandleInputSwap()
	} else if outputToken == consts.GNOT {
		handler.HandleOutputSwap()
	}

	if err != nil {
		panic(addDetailToError(errSlippage, err.Error()))
	}

	if err := validator.slippage(swapType, resultAmountIn, resultAmountOutWithoutFee, tokenAmountLimit); err != nil {
		panic(addDetailToError(errSlippage, err.Error()))
	}

	// calculate final amounts
	intAmountOut := i256.FromUint256(resultAmountOutWithoutFee)
	negativeOut := i256.Zero().Neg(intAmountOut).ToString()

	return resultAmountIn.ToString(), negativeOut
}

// validateRoutesAndQuotes is a convenience function that parses and validates routes in one call
func validateRoutesAndQuotes(routes, quotes string) ([]string, []string, error) {
	return NewRouteParser().ParseRoutes(routes, quotes)
}
