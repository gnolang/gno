package overflow

// Checked conversions out of int256 and uint256. A Gno conversion such as
// int64(x) keeps the low bits and wraps, as in Go; these report whether the
// value fits the target type instead. The result is the wrapped conversion,
// so it is the converted value whenever the boolean is true. The "p" versions
// panic when the value does not fit.

// Uint256ToInt64 converts x to int64, reporting whether it fits.
func Uint256ToInt64(x uint256) (int64, bool) {
	return int64(x), x <= 1<<63-1
}

// Uint256ToUint64 converts x to uint64, reporting whether it fits.
func Uint256ToUint64(x uint256) (uint64, bool) {
	return uint64(x), x <= 1<<64-1
}

// Uint256ToInt256 converts x to int256, reporting whether it fits.
func Uint256ToInt256(x uint256) (int256, bool) {
	return int256(x), x < minInt256Bits
}

// Int256ToInt64 converts x to int64, reporting whether it fits.
func Int256ToInt64(x int256) (int64, bool) {
	return int64(x), x >= -1<<63 && x <= 1<<63-1
}

// Int256ToUint64 converts x to uint64, reporting whether it fits.
func Int256ToUint64(x int256) (uint64, bool) {
	return uint64(x), x >= 0 && uint256(x) <= 1<<64-1
}

// Int256ToUint256 converts x to uint256, reporting whether it fits.
func Int256ToUint256(x int256) (uint256, bool) {
	return uint256(x), x >= 0
}

// Uint256ToInt64p converts x to int64, panicking if it does not fit.
func Uint256ToInt64p(x uint256) int64 {
	r, ok := Uint256ToInt64(x)
	if !ok {
		panic("conversion overflow")
	}
	return r
}

// Uint256ToUint64p converts x to uint64, panicking if it does not fit.
func Uint256ToUint64p(x uint256) uint64 {
	r, ok := Uint256ToUint64(x)
	if !ok {
		panic("conversion overflow")
	}
	return r
}

// Uint256ToInt256p converts x to int256, panicking if it does not fit.
func Uint256ToInt256p(x uint256) int256 {
	r, ok := Uint256ToInt256(x)
	if !ok {
		panic("conversion overflow")
	}
	return r
}

// Int256ToInt64p converts x to int64, panicking if it does not fit.
func Int256ToInt64p(x int256) int64 {
	r, ok := Int256ToInt64(x)
	if !ok {
		panic("conversion overflow")
	}
	return r
}

// Int256ToUint64p converts x to uint64, panicking if it does not fit.
func Int256ToUint64p(x int256) uint64 {
	r, ok := Int256ToUint64(x)
	if !ok {
		panic("conversion overflow")
	}
	return r
}

// Int256ToUint256p converts x to uint256, panicking if it does not fit.
func Int256ToUint256p(x int256) uint256 {
	r, ok := Int256ToUint256(x)
	if !ok {
		panic("conversion overflow")
	}
	return r
}
