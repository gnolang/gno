package calibrate

// strconv native calibration benchmarks.
//
// parseUint256Digits does the same 256-bit multiply-add for every digit
// whatever its value, so the worst case per byte is a string that stays
// valid to its end: leading zeros (a value that never overflows however long
// the string is) followed by the largest base-36 digits that still fit.
// Underscores are cheaper (skipped without arithmetic), so they are left out.
//
// Run:
//   cd gnovm/cmd/calibrate
//   go test -bench=BenchmarkNative_Strconv -benchtime=200ms -count=3 . > strconv_bench_output.txt
//   python3 gen_native_table.py strconv_bench_output.txt

import (
	"strings"
	"testing"
)

func benchParseUint256Digits(b *testing.B, n int) {
	b.Helper()
	// 49 base-36 'z' digits are below 2^256; 50 would overflow.
	k := min(n, 49)
	s := strings.Repeat("0", n-k) + strings.Repeat("z", k)
	m := newDispatchMachine(3)
	setBlockValueFromGo(m, 0, s)
	setBlockValueFromGo(m, 1, 36)
	setBlockValueFromGo(m, 2, false)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "strconv", "parseUint256Digits"), nReturns: 3}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Strconv_ParseUint256Digits_1(b *testing.B)   { benchParseUint256Digits(b, 1) }
func BenchmarkNative_Strconv_ParseUint256Digits_10(b *testing.B)  { benchParseUint256Digits(b, 10) }
func BenchmarkNative_Strconv_ParseUint256Digits_100(b *testing.B) { benchParseUint256Digits(b, 100) }
func BenchmarkNative_Strconv_ParseUint256Digits_1000(b *testing.B) {
	benchParseUint256Digits(b, 1000)
}

func BenchmarkNative_Strconv_ParseUint256Digits_10000(b *testing.B) {
	benchParseUint256Digits(b, 10000)
}

func BenchmarkNative_Strconv_ParseUint256Digits_100000(b *testing.B) {
	benchParseUint256Digits(b, 100000)
}
