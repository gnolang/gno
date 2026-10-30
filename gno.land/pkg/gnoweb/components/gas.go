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
// calls (a wugnot Deposit uses 7 to 9 M on onyx, a small addpkg ~2.6 M);
// gnokey's default simulation stops a call that needs more before anything
// is charged.
const TxGasWanted = 10_000_000

// txFeeHeadroom multiplies the fee so the command still pays enough if the
// gas price rises between page load and broadcast. tm2 moves the price by at
// least one unit per busy block, so at 1ugnot/1000gas one block doubles it.
const txFeeHeadroom = 3

// Headroom is the factor the fee is multiplied by over the current price.
func (g TxGas) Headroom() int { return txFeeHeadroom }

// maxUgnotFee is the old hardcoded fee, 1 GNOT; a price asking more is unusable.
const maxUgnotFee = 1_000_000

// FallbackGasPrice is gno.land's usual gas price, used when the node's
// cannot be read or is unusable.
var FallbackGasPrice = std.GasPrice{Gas: 1_000, Price: std.NewCoin("ugnot", 1)}

// minTxGasWanted is the least gas a command sized from recent calls asks for.
const minTxGasWanted = 1_000_000

// gasWantedStep is what gas sized from recent calls is rounded up to.
const gasWantedStep = 100_000

// minGasUseCalls is how many recent calls it takes to size the gas from them.
const minGasUseCalls = 3

// GasUse is what a function's recent successful calls used, as an indexer
// reports it.
type GasUse struct {
	Max   int64 // the most gas one of them used
	Calls int
}

// GasFee is the gas a gnokey command asks for and the fee it offers.
type GasFee struct {
	Wanted int64
	Fee    std.Coin
}

// TxGas is the gas and fee the copied gnokey commands offer, and the price
// the fees were worked out from.
type TxGas struct {
	// Secure is what the secure-mode commands offer: TxGasWanted, since they
	// broadcast without simulating.
	Secure GasFee
	// Fast is what the fast-mode command offers: Secure, or less when Used
	// sized it. gnokey simulates it first, so too little gas costs nothing.
	Fast  GasFee
	Price std.GasPrice
	// Estimated is set when Price is FallbackGasPrice, not the chain's.
	Estimated bool
	// Used is what Fast was sized from, zero when Fast is Secure.
	Used GasUse
}

// NewTxGas prices TxGasWanted at price times txFeeHeadroom, rounded up; a
// zero, invalid, overflowing or too costly price falls back to
// FallbackGasPrice. With minGasUseCalls recent calls or more, the fast-mode
// command asks for twice the most gas one used, rounded up to gasWantedStep
// and at least minTxGasWanted, at the same price.
func NewTxGas(price std.GasPrice, use GasUse) TxGas {
	g := TxGas{Price: price}
	fee, ok := txFee(price, TxGasWanted)
	if !ok {
		fee, _ = txFee(FallbackGasPrice, TxGasWanted)
		g.Price, g.Estimated = FallbackGasPrice, true
	}
	g.Secure = GasFee{Wanted: TxGasWanted, Fee: fee}
	g.Fast = g.Secure
	// Indexer data is not consensus data: it may lower the gas, never raise it.
	if use.Calls >= minGasUseCalls && use.Max < TxGasWanted/2 {
		w := max((2*use.Max+gasWantedStep-1)/gasWantedStep*gasWantedStep, minTxGasWanted)
		if w < TxGasWanted {
			fee, _ := txFee(g.Price, w) // below TxGasWanted's, which was usable
			g.Fast, g.Used = GasFee{Wanted: w, Fee: fee}, use
		}
	}
	return g
}

// txFee is ceil(wanted × amount × txFeeHeadroom / gas).
func txFee(p std.GasPrice, wanted int64) (std.Coin, bool) {
	if p.Gas <= 0 || p.Price.Amount <= 0 || !p.Price.IsValid() {
		return std.Coin{}, false
	}
	num := big.NewInt(wanted * txFeeHeadroom)
	num.Mul(num, big.NewInt(p.Price.Amount))
	den := big.NewInt(p.Gas)
	fee := num.Add(num, den).Sub(num, big.NewInt(1)).Quo(num, den)
	if !fee.IsInt64() || (p.Price.Denom == "ugnot" && fee.Int64() > maxUgnotFee) {
		return std.Coin{}, false
	}
	return std.NewCoin(p.Price.Denom, fee.Int64()), true
}

// WantedFlag is the -gas-wanted value, grouped like Go literals: 10_000_000.
func (g GasFee) WantedFlag() string {
	s := strconv.FormatInt(g.Wanted, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "_" + s[i:]
	}
	return s
}

// FeeText is the fee for a reader: GNOT for ugnot, else the raw coin.
func (g GasFee) FeeText() string {
	if g.Fee.Denom != "ugnot" {
		return g.Fee.String()
	}
	return strconv.FormatFloat(float64(g.Fee.Amount)/1e6, 'f', -1, 64) + " GNOT"
}
