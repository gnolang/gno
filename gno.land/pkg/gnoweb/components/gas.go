package components

import (
	"math"
	"math/big"
	"strconv"

	"github.com/gnolang/gno/tm2/pkg/std"
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

// TxGasWanted is the gas a copied gnokey command asks for. It covers most
// calls (a wugnot Deposit uses ~1.4 M, a small addpkg ~2.6 M); gnokey's
// default simulation stops a call that needs more before anything is charged.
const TxGasWanted = 10_000_000

// txFeeMargin is added to the fee, in percent, so the command still pays
// enough if the gas price rises between page load and broadcast.
const txFeeMargin = 50

// Margin is the percent added to the fee over the price.
func (g TxGas) Margin() int { return txFeeMargin }

// maxUgnotFee is the old hardcoded fee, 1 GNOT; a price asking more is unusable.
const maxUgnotFee = 1_000_000

// FallbackGasPrice is gno.land's usual gas price, used when the node's
// cannot be read or is unusable.
var FallbackGasPrice = std.GasPrice{Gas: 1_000, Price: std.NewCoin("ugnot", 1)}

// TxGas is the gas and fee a copied gnokey command offers, and the price the
// fee was worked out from.
type TxGas struct {
	Wanted int64
	Fee    std.Coin
	Price  std.GasPrice
	// Estimated is set when Price is FallbackGasPrice, not the chain's.
	Estimated bool
}

// NewTxGas prices TxGasWanted at price plus txFeeMargin, rounded up. A zero,
// invalid, overflowing or too costly price falls back to FallbackGasPrice.
func NewTxGas(price std.GasPrice) TxGas {
	if fee, ok := txFee(price); ok {
		return TxGas{Wanted: TxGasWanted, Fee: fee, Price: price}
	}
	fee, _ := txFee(FallbackGasPrice)
	return TxGas{Wanted: TxGasWanted, Fee: fee, Price: FallbackGasPrice, Estimated: true}
}

// txFee is ceil(TxGasWanted × amount / gas × (100+txFeeMargin) / 100).
func txFee(p std.GasPrice) (std.Coin, bool) {
	if p.Gas <= 0 || p.Price.Amount <= 0 || !p.Price.IsValid() {
		return std.Coin{}, false
	}
	num := big.NewInt(TxGasWanted * (100 + txFeeMargin))
	num.Mul(num, big.NewInt(p.Price.Amount))
	den := new(big.Int).Mul(big.NewInt(p.Gas), big.NewInt(100))
	fee := num.Add(num, den).Sub(num, big.NewInt(1)).Quo(num, den)
	if !fee.IsInt64() || (p.Price.Denom == "ugnot" && fee.Int64() > maxUgnotFee) {
		return std.Coin{}, false
	}
	return std.NewCoin(p.Price.Denom, fee.Int64()), true
}

// WantedFlag is the -gas-wanted value, grouped like Go literals: 10_000_000.
func (g TxGas) WantedFlag() string {
	s := strconv.FormatInt(g.Wanted, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "_" + s[i:]
	}
	return s
}

// FeeText is the fee for a reader: GNOT for ugnot, else the raw coin.
func (g TxGas) FeeText() string {
	if g.Fee.Denom != "ugnot" {
		return g.Fee.String()
	}
	return strconv.FormatFloat(float64(g.Fee.Amount)/1e6, 'f', -1, 64) + " GNOT"
}
