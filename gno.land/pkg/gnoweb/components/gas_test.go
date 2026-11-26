package components

import "testing"

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
