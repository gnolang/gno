package render

import (
	"time"
	"strconv"

	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/p/volos/consts"
)

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
 
func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("02 Jan 2006, 15:04:05")
}

func formatPercentage(wad *u256.Uint) string {
	// Round to 2 decimals and always display two decimal digits.
	// Compute: q = round((wad * 10000) / 1e18), where q = percent * 100
	num := new(u256.Uint).Mul(wad, u256.NewUint(10000))
	q := new(u256.Uint)
	r := new(u256.Uint)
	q, r = q.DivMod(num, consts.WAD, r)

	// Arithmetic (.5 rounds up)
	if new(u256.Uint).Mul(r, u256.NewUint(2)).Gte(consts.WAD) {
		q = new(u256.Uint).Add(q, u256.One())
	}

	intPart := new(u256.Uint).Div(q, u256.NewUint(100))
	frac := new(u256.Uint).Mod(q, u256.NewUint(100))
	intStr := intPart.ToString()
	fracStr := frac.ToString()
	if len(fracStr) == 1 {
		fracStr = "0" + fracStr
	}
	return intStr + "." + fracStr
}

// formatTokenAmount converts a raw token amount (in smallest units) into a
// human-readable decimal string using tokenDecimals. It does not use floats.
// - Keeps full fractional precision implied by tokenDecimals
// - Trims redundant trailing zeros in the fractional part
// - Omits the decimal point if the fractional part becomes empty
func formatTokenAmount(tokenAmount *u256.Uint, tokenDecimals int) string {
	if tokenAmount == nil {
		return "0"
	}
	if tokenDecimals == 0 {
		return tokenAmount.ToString()
	}

	// Build 10^tokenDecimals as uint256
	pow := u256.One()
	for i := 0; i < tokenDecimals; i++ {
		pow = new(u256.Uint).Mul(pow, u256.NewUint(10))
	}

	// Quotient and remainder: tokenAmount / 10^decimals
	q := new(u256.Uint)
	r := new(u256.Uint)
	q, r = q.DivMod(tokenAmount, pow, r)

	intStr := q.ToString()
	if r.IsZero() {
		return intStr
	}

	// Fractional part needs left-padding to exactly tokenDecimals digits
	fracStr := r.ToString()
	for lenFrac := len(fracStr); lenFrac < tokenDecimals; lenFrac++ {
		fracStr = "0" + fracStr
	}

	// Trim trailing zeros
	idx := len(fracStr) - 1
	for idx >= 0 && fracStr[idx] == '0' {
		idx--
	}
	if idx < 0 {
		return intStr
	}
	fracStr = fracStr[:idx+1]

	return intStr + "." + fracStr
}

// pow10Uint returns 10^exp as *u256.Uint. exp>=0.
func pow10Uint(exp int) *u256.Uint {
	p := u256.One()
	if exp <= 0 {
		return p
	}
	ten := u256.NewUint(10)
	for i := 0; i < exp; i++ {
		p = new(u256.Uint).Mul(p, ten)
	}
	return p
}

// formatPrice formats a price by moving the decimal point by places = 36 + loanDecimals - collateralDecimals.
// If places > 0, it divides by 10^places and trims trailing zeros like formatTokenAmount.
// If places <= 0, it multiplies by 10^(-places) and returns the integer string.
func formatPrice(price *u256.Uint, loanDecimals, collateralDecimals int) string {
	if price == nil {
		return "0"
	}

	places := 36 + loanDecimals - collateralDecimals
	if places <= 0 {
		mul := pow10Uint(-places)
		val := new(u256.Uint).Mul(price, mul)
		return val.ToString()
	}

	pow := pow10Uint(places)
	q := new(u256.Uint)
	r := new(u256.Uint)
	q, r = q.DivMod(price, pow, r)
	intStr := q.ToString()
	if r.IsZero() {
		return intStr
	}
	fracStr := r.ToString()
	for lenFrac := len(fracStr); lenFrac < places; lenFrac++ {
		fracStr = "0" + fracStr
	}
	idx := len(fracStr) - 1
	for idx >= 0 && fracStr[idx] == '0' {
		idx--
	}
	if idx < 0 {
		return intStr
	}
	fracStr = fracStr[:idx+1]
	return intStr + "." + fracStr
}
