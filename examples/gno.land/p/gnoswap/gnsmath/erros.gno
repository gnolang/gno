package gnsmath

import (
	"errors"
)

var (
	errInvalidPoolSqrtPrice      = errors.New("invalid pool sqrt price calculation: product/amount != sqrtPX96 or numerator1 <= product")
	errNextSqrtPriceOverflow     = errors.New("nextSqrtPrice overflows uint160")
	errSqrtPriceQuotientOverflow = errors.New("GetNextSqrtPriceFromAmount1RoundingDown sqrtPx96 + quotient overflow uint160")
	errSqrtPriceExceedsQuotient  = errors.New("sqrt price exceeds calculated quotient")
	errSqrtPriceZero             = errors.New("sqrtPX96 should not be zero")
	errLiquidityZero             = errors.New("liquidity should not be zero")
	errSqrtRatioAX96NotPositive  = errors.New("sqrtRatioAX96 must be greater than zero")
	errAmount0DeltaOverflow      = errors.New("GetAmount0DeltaStr: overflow")
	errAmount1DeltaOverflow      = errors.New("GetAmount1DeltaStr: overflow")
	errMSBZeroInput              = errors.New("input for MSB calculation should not be zero")
	errLSBZeroInput              = errors.New("input for LSB calculation should not be zero")
)
