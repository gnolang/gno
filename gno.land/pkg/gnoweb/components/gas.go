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

// maxUgnotFee is the old hardcoded fee, 1 GNOT. A fee is capped at it, and a
// price whose gas alone costs more is unusable.
const maxUgnotFee = 1_000_000

// maxTxGasWanted is the old hardcoded gas, the most a measure raises a
// command to.
const maxTxGasWanted = 1_000_000_000

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
	// Secure is what the secure-mode commands offer: TxGasWanted, or more
	// when Used asks for more. They broadcast without simulating, so they
	// never ask for less.
	Secure GasFee
	// Fast is what the fast-mode command offers: Secure, or less when Used
	// sized it. gnokey simulates it first, so too little gas costs nothing.
	Fast  GasFee
	Price std.GasPrice
	// Estimated is set when Price is FallbackGasPrice, not the chain's.
	Estimated bool
	// Used is what the gas was sized from, zero when it is TxGasWanted.
	Used GasUse
}

// NewTxGas prices TxGasWanted at price times txFeeHeadroom, rounded up and
// capped at maxUgnotFee; a zero, invalid or overflowing price, or one whose
// gas alone costs more than maxUgnotFee, falls back to FallbackGasPrice.
// With minGasUseCalls recent calls or more, the commands ask for twice the
// most gas one used, rounded up to gasWantedStep, at least minTxGasWanted and
// at most what maxUgnotFee pays for, up to maxTxGasWanted: the fast-mode one
// for less or more than TxGasWanted, the secure-mode ones for more only.
func NewTxGas(price std.GasPrice, use GasUse) TxGas {
	g := TxGas{Price: price}
	fee, ok := txFee(price, TxGasWanted)
	if !ok {
		fee, _ = txFee(FallbackGasPrice, TxGasWanted)
		g.Price, g.Estimated = FallbackGasPrice, true
	}
	g.Secure = GasFee{Wanted: TxGasWanted, Fee: fee}
	g.Fast = g.Secure
	if use.Calls < minGasUseCalls {
		return g
	}
	// Indexer data is not consensus data. Lowering the gas costs nothing to
	// get wrong, as gnokey simulates; raising it costs at worst the old
	// hardcoded fee, which every call paid before.
	most := maxWanted(g.Price)
	w := most
	if use.Max < most/2 {
		w = min(max((2*use.Max+gasWantedStep-1)/gasWantedStep*gasWantedStep, minTxGasWanted), most)
	}
	fee, _ = txFee(g.Price, w) // w is within what maxUgnotFee pays for
	switch {
	case w < TxGasWanted:
		g.Fast, g.Used = GasFee{Wanted: w, Fee: fee}, use
	case w > TxGasWanted:
		g.Secure, g.Used = GasFee{Wanted: w, Fee: fee}, use
		g.Fast = g.Secure
	}
	return g
}

// maxWanted is the most gas, in gasWantedStep, a command asks for at p:
// maxTxGasWanted, or less when maxUgnotFee pays for less. p is usable, so
// that is at least TxGasWanted. A price in another denom is not raised.
func maxWanted(p std.GasPrice) int64 {
	if p.Price.Denom != "ugnot" {
		return TxGasWanted // no 1 GNOT cap bounds a raise in another denom
	}
	w := new(big.Int).Mul(big.NewInt(maxUgnotFee), big.NewInt(p.Gas))
	w.Quo(w, big.NewInt(p.Price.Amount))
	if !w.IsInt64() || w.Int64() > maxTxGasWanted {
		return maxTxGasWanted
	}
	return w.Int64() / gasWantedStep * gasWantedStep
}

// txFee is ceil(wanted × amount × txFeeHeadroom / gas), a ugnot fee capped at
// maxUgnotFee. It reports false when the price is unusable for wanted: zero,
// invalid, overflowing, or, in ugnot, asking more than maxUgnotFee for the
// gas alone.
func txFee(p std.GasPrice, wanted int64) (std.Coin, bool) {
	if p.Gas <= 0 || p.Price.Amount <= 0 || !p.Price.IsValid() {
		return std.Coin{}, false
	}
	den := big.NewInt(p.Gas)
	ceilTimes := func(k int64) *big.Int { // ceil(wanted × amount × k / gas)
		n := new(big.Int).Mul(big.NewInt(wanted), big.NewInt(p.Price.Amount))
		n.Mul(n, big.NewInt(k))
		return n.Add(n, den).Sub(n, big.NewInt(1)).Quo(n, den)
	}
	fee := ceilTimes(txFeeHeadroom)
	if !fee.IsInt64() {
		return std.Coin{}, false
	}
	if p.Price.Denom == "ugnot" && fee.Int64() > maxUgnotFee {
		// Capped: the headroom shrinks, but the fee still pays the gas at
		// the price read.
		if ceilTimes(1).Cmp(big.NewInt(maxUgnotFee)) > 0 {
			return std.Coin{}, false
		}
		return std.NewCoin(p.Price.Denom, maxUgnotFee), true
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
