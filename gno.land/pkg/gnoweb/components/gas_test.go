package components

import (
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// Gas is written in the unit its rounded value reads best in: 999,600 is
// 1.0 M, never "1000 k".
func TestFormatGas(t *testing.T) {
	t.Parallel()

	for in, want := range map[int64]string{
		0: "0", 950: "950", 50_000: "50 k", 999_600: "1.0 M", 1_900_000: "1.9 M",
		9_960_000: "10 M", 154_000_000: "154 M", 999_600_000: "1.0 bn", 76_040_000_000: "76.0 bn",
	} {
		if got := FormatGas(in); got != want {
			t.Errorf("FormatGas(%d) = %q, want %q", in, got, want)
		}
	}
}

// The fee is TxGasWanted at the price plus the margin, rounded up, in the
// price's denom; a price that cannot be used falls back and says so.
func TestNewTxGas(t *testing.T) {
	t.Parallel()

	price := func(amount, gas int64, denom string) std.GasPrice {
		return std.GasPrice{Gas: gas, Price: std.Coin{Denom: denom, Amount: amount}}
	}
	for name, tc := range map[string]struct {
		price     std.GasPrice
		fee       string
		estimated bool
	}{
		"chain default":   {price(1, 1000, "ugnot"), "15000ugnot", false},
		"rounds up":       {price(1, 7, "ugnot"), "2142858ugnot", false}, // 2,142,857.14
		"risen price":     {price(3, 1000, "ugnot"), "45000ugnot", false},
		"other denom":     {price(2, 1000, "foo"), "30000foo", false},
		"huge gas, cheap": {price(1, math.MaxInt64, "ugnot"), "1ugnot", false},
		"zero value":      {std.GasPrice{}, "15000ugnot", true},
		"zero amount":     {price(0, 1000, "ugnot"), "15000ugnot", true},
		"zero gas":        {price(1, 0, "ugnot"), "15000ugnot", true},
		"negative gas":    {price(1, -1000, "ugnot"), "15000ugnot", true},
		"invalid denom":   {price(1, 1000, "u gnot"), "15000ugnot", true},
		"fee overflows":   {price(math.MaxInt64, 1, "ugnot"), "15000ugnot", true},
	} {
		got := NewTxGas(tc.price)
		if got.Fee.String() != tc.fee || got.Estimated != tc.estimated || got.Wanted != TxGasWanted {
			t.Errorf("%s: NewTxGas(%v) = %+v, want fee %s, estimated %v", name, tc.price, got, tc.fee, tc.estimated)
		}
		if tc.estimated && got.Price != FallbackGasPrice {
			t.Errorf("%s: price %v, want the fallback", name, got.Price)
		}
	}
}

// What the command and the note under it show.
func TestTxGasText(t *testing.T) {
	t.Parallel()

	g := NewTxGas(FallbackGasPrice)
	for got, want := range map[string]string{
		g.WantedFlag(): "10_000_000",
		g.WantedText(): "10,000,000",
		g.FeeText():    "0.015 GNOT",
		g.PriceText():  "1 ugnot per 1,000 gas",
	} {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	for amount, want := range map[int64]string{
		0: "0 GNOT", 1: "0.000001 GNOT", 15_000: "0.015 GNOT", 1_000_000: "1 GNOT",
		1_500_000: "1.5 GNOT", 1_234_567_890: "1,234.56789 GNOT",
	} {
		if got := coinText(std.NewCoin("ugnot", amount)); got != want {
			t.Errorf("coinText(%dugnot) = %q, want %q", amount, got, want)
		}
	}
	if got := coinText(std.NewCoin("foo", 1_234_567)); got != "1,234,567 foo" {
		t.Errorf("coinText(1234567foo) = %q", got)
	}
}
