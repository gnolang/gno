package strconv

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/holiman/uint256"
)

// Status values of parseUint256Digits. Keep in sync with int256.gno.
const (
	digitsOK     = 0
	digitsSyntax = 1
	digitsRange  = 2
)

// X_parseUint256Digits is the digit loop of ParseUint256, run natively: the
// interpreter otherwise runs one iteration of 256-bit arithmetic per digit,
// about 10k gas each. It reads s in base 2 to 36, left to right, and stops at
// the first digit that is invalid (digitsSyntax) or makes the value overflow
// a uint256 (digitsRange), in that order for each digit, exactly as the Gno
// loop did. With skipUnderscores (base 0), '_' is skipped and reported in
// underscores; ParseUint256 validates their placement afterwards. On failure
// the value is 0.
func X_parseUint256Digits(s string, base int, skipUnderscores bool) (_ gno.Uint256, status int, underscores bool) {
	if base < 2 || base > 36 {
		panic("strconv: parseUint256Digits: invalid base")
	}
	var zero, b, cutoff, n, dv uint256.Int
	b.SetUint64(uint64(base))
	// cutoff is the smallest number such that cutoff*base > maxUint256.
	cutoff.SetAllOne()
	cutoff.Div(&cutoff, &b)
	cutoff.AddUint64(&cutoff, 1)

	for i := 0; i < len(s); i++ {
		c := s[i]
		var d byte
		switch lc := c | ('x' - 'X'); {
		case c == '_' && skipUnderscores:
			underscores = true
			continue
		case '0' <= c && c <= '9':
			d = c - '0'
		case 'a' <= lc && lc <= 'z':
			d = lc - 'a' + 10
		default:
			return zero, digitsSyntax, underscores
		}
		if d >= byte(base) {
			return zero, digitsSyntax, underscores
		}
		if !n.Lt(&cutoff) {
			// n*base overflows
			return zero, digitsRange, underscores
		}
		n.Mul(&n, &b)
		dv.SetUint64(uint64(d))
		if _, overflow := n.AddOverflow(&n, &dv); overflow {
			// n+d overflows
			return zero, digitsRange, underscores
		}
	}
	return n, digitsOK, underscores
}
