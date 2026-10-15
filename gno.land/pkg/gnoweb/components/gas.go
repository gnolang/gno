package components

import (
	"math"
	"strconv"
)

// FormatGas writes an amount of gas in thousands ("50 k"), millions ("1.9 M",
// "154 M") or billions ("76.0 bn"), the unit chosen on the rounded value so
// 999,600 reads 1.0 M, not 1000 k. "bn", not "B", which reads as bytes beside
// a page's storage sizes.
func FormatGas(g int64) string {
	f := float64(g)
	switch {
	case g < 1_000:
		return strconv.FormatInt(g, 10)
	case math.Round(f/1e3) < 1_000:
		return strconv.FormatFloat(f/1e3, 'f', 0, 64) + " k"
	case math.Round(f/1e5) < 100:
		return strconv.FormatFloat(f/1e6, 'f', 1, 64) + " M"
	case math.Round(f/1e6) < 1_000:
		return strconv.FormatFloat(f/1e6, 'f', 0, 64) + " M"
	default:
		return strconv.FormatFloat(f/1e9, 'f', 1, 64) + " bn"
	}
}
