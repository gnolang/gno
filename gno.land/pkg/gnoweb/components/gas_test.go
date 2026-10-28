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

// The fee is TxGasWanted at the price times the headroom, rounded up, in the
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
		"chain default":       {price(1, 1000, "ugnot"), "30000ugnot", false},
		"rounds up":           {price(1, 700, "ugnot"), "42858ugnot", false}, // 42,857.14
		"at the cap":          {price(1, 30, "ugnot"), "1000000ugnot", false},
		"over the cap":        {price(1, 29, "ugnot"), "30000ugnot", true},
		"other denom, no cap": {price(1000, 1, "foo"), "30000000000foo", false},
		"risen price":         {price(3, 1000, "ugnot"), "90000ugnot", false},
		"other denom":         {price(2, 1000, "foo"), "60000foo", false},
		"huge gas, cheap":     {price(1, math.MaxInt64, "ugnot"), "1ugnot", false},
		"zero value":          {std.GasPrice{}, "30000ugnot", true},
		"zero amount":         {price(0, 1000, "ugnot"), "30000ugnot", true},
		"zero gas":            {price(1, 0, "ugnot"), "30000ugnot", true},
		"negative gas":        {price(1, -1000, "ugnot"), "30000ugnot", true},
		"invalid denom":       {price(1, 1000, "u gnot"), "30000ugnot", true},
		"fee overflows":       {price(math.MaxInt64, 1, "ugnot"), "30000ugnot", true},
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
	fee := func(c std.Coin) string { return TxGas{Fee: c}.FeeText() }
	for _, tc := range []struct{ name, got, want string }{
		{"flag", g.WantedFlag(), "10_000_000"},
		{"fee", g.FeeText(), "0.03 GNOT"},
		{"price", g.Price.String(), "1ugnot/1000gas"},
		{"zero", fee(std.NewCoin("ugnot", 0)), "0 GNOT"},
		{"one ugnot", fee(std.NewCoin("ugnot", 1)), "0.000001 GNOT"},
		{"one GNOT", fee(std.NewCoin("ugnot", 1_000_000)), "1 GNOT"},
		{"fraction", fee(std.NewCoin("ugnot", 1_234_567_890)), "1234.56789 GNOT"},
		{"other denom", fee(std.NewCoin("foo", 1_234_567)), "1234567foo"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}
