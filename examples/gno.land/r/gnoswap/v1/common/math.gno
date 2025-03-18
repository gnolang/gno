package common

import (
	"gno.land/p/demo/ufmt"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
)

// I64Min returns the minimum of two int64 values
func I64Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// I64Max returns the maximum of two int64 values
func I64Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// U64Min returns the minimum of two uint64 values
func U64Min(x, y uint64) uint64 {
	if x < y {
		return x
	}
	return y
}

// U64Max returns the maximum of two uint64 values
func U64Max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

// I256Min returns the minimum of two Int256 values
func I256Min(x, y *i256.Int) *i256.Int {
	if x.Cmp(y) < 0 {
		return x
	}
	return y
}

// I256Max returns the maximum of two Int256 values
func I256Max(x, y *i256.Int) *i256.Int {
	if x.Cmp(y) > 0 {
		return x
	}
	return y
}

// U256Min returns the minimum of two Uint256 values
func U256Min(x, y *u256.Uint) *u256.Uint {
	if x.Cmp(y) < 0 {
		return x
	}
	return y
}

// U256Max returns the maximum of two Uint256 values
func U256Max(x, y *u256.Uint) *u256.Uint {
	if x.Cmp(y) > 0 {
		return x
	}
	return y
}

// SafeConvertUint256ToInt256 converts a uint256.Uint to int256.Int and returns it.
// If the value is greater than the maximum int256 value, it panics.
func SafeConvertUint256ToInt256(x *u256.Uint) *i256.Int {
	if x.Gt(u256.MustFromDecimal(consts.MAX_INT256)) {
		panic(newErrorWithDetail(
			errOverflow,
			ufmt.Sprintf("can not convert %s to int256", x.ToString()),
		))
	}
	return i256.FromUint256(x)
}

// SafeConvertUint256ToUint64 converts a uint256.Uint to uint64 and returns it.
// If the value is greater than the maximum uint64 value, it panics.
func SafeConvertUint256ToUint64(x *u256.Uint) uint64 {
	value, overflow := x.Uint64WithOverflow()
	if overflow {
		panic(newErrorWithDetail(
			errOverflow,
			ufmt.Sprintf("can not convert %s to uint64", x.ToString()),
		))
	}

	return value
}
