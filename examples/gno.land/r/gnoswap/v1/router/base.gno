package router

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/demo/wugnot"
)

const (
	SINGLE_HOP_ROUTE int = 1

	INITIAL_WUGNOT_BALANCE uint64 = 0
)

// swap can be done by multiple pools
// to separate each pool, we use POOL_SEPARATOR
const (
	POOL_SEPARATOR = "*POOL*"
)

type RouterOperation interface {
	Validate() error
	Process() (*SwapResult, error)
}

func executeSwapOperation(op RouterOperation) (*SwapResult, error) {
	if err := op.Validate(); err != nil {
		return nil, err
	}

	result, err := op.Process()
	if err != nil {
		return nil, err
	}

	return result, nil
}

type BaseSwapParams struct {
	InputToken  string
	OutputToken string
	RouteArr    string
	QuoteArr    string
	Deadline    int64
}

// common swap operation
type baseSwapOperation struct {
	routes                  []string
	quotes                  []string
	amountSpecified         *i256.Int
	userBeforeWugnotBalance uint64
	userWrappedWugnot       uint64
}

func (op *baseSwapOperation) handleNativeTokenWrapping(
	inputToken string,
	outputToken string,
	swapType SwapType,
	specifiedAmount *i256.Int,
) error {
	// no native token
	if inputToken == consts.WUGNOT_PATH || outputToken == consts.WUGNOT_PATH {
		return nil
	}

	// save current user's WGNOT amount
	op.userBeforeWugnotBalance = wugnot.BalanceOf(std.PreviousRealm().Address())

	if swapType == ExactIn && inputToken == consts.GNOT {
		sent := std.OriginSend()

		ugnotSentByUser := uint64(sent.AmountOf("ugnot"))
		amountSpecified := specifiedAmount.Uint64()

		if ugnotSentByUser != amountSpecified {
			return ufmt.Errorf("ugnot sent by user(%d) is not equal to amountSpecified(%d)", ugnotSentByUser, amountSpecified)
		}

		// wrap user's WUGNOT
		if ugnotSentByUser > 0 {
			wrap(ugnotSentByUser)
		}

		op.userWrappedWugnot = ugnotSentByUser
	}

	return nil
}

func (op *baseSwapOperation) validateRouteQuote(quote string, i int) (*i256.Int, error) {
	qt, err := strconv.Atoi(quote)
	if err != nil {
		return nil, ufmt.Errorf("invalid quote(%s) at index(%d)", quote, i)
	}

	// calculate amount to swap for this route
	toSwap := i256.Zero().Mul(op.amountSpecified, i256.NewInt(int64(qt)))
	toSwap = i256.Zero().Div(toSwap, i256.NewInt(100))

	return toSwap, nil
}

func (op *baseSwapOperation) processRoutes(swapType SwapType) (*u256.Uint, *u256.Uint, error) {
	resultAmountIn := u256.Zero()
	resultAmountOut := u256.Zero()

	for i, route := range op.routes {
		toSwap, err := op.validateRouteQuote(op.quotes[i], i)
		if err != nil {
			return nil, nil, err
		}

		amountIn, amountOut, err := op.processRoute(route, toSwap, swapType)
		if err != nil {
			return nil, nil, err
		}

		resultAmountIn = new(u256.Uint).Add(resultAmountIn, amountIn)
		resultAmountOut = new(u256.Uint).Add(resultAmountOut, amountOut)
	}

	return resultAmountIn, resultAmountOut, nil
}

func (op *baseSwapOperation) processRoute(
	route string,
	toSwap *i256.Int,
	swapType SwapType,
) (*u256.Uint, *u256.Uint, error) {
	numHops := strings.Count(route, POOL_SEPARATOR) + 1
	assertHopsInRange(numHops)

	var amountIn, amountOut *u256.Uint

	switch numHops {
	case SINGLE_HOP_ROUTE:
		amountIn, amountOut = handleSingleSwap(route, toSwap)
	default:
		amountIn, amountOut = handleMultiSwap(swapType, route, numHops, toSwap)
	}

	if amountIn == nil || amountOut == nil {
		return nil, nil, ufmt.Errorf("swap failed to process route(%s)", route)
	}

	return amountIn, amountOut, nil
}
