package overflow_test

import (
	"math"
	"math/overflow"
	"testing"
)

// An int256 constant must fit an int64, so wide int256 test values are
// built at run time.
var (
	maxBits uint256 = math.MaxInt256
	i256Max         = int256(maxBits)
	i256Min         = int256(maxBits + 1)
)

func pow2(n uint) int256 {
	var one int256 = 1
	return one << n
}

func TestInt256(t *testing.T) {
	max, min := i256Max, i256Min
	type op func(a, b int256) (int256, bool)
	tests := []struct {
		name string
		fn   op
		a, b int256
		want int256
		ok   bool
	}{
		{"Add", overflow.Add256, 1, 2, 3, true},
		{"Add", overflow.Add256, max, 0, max, true},
		{"Add", overflow.Add256, max, 1, min, false},
		{"Add", overflow.Add256, min, -1, max, false},
		{"Add", overflow.Add256, min, max, -1, true},
		{"Add", overflow.Add256, max, max, -2, false},
		{"Sub", overflow.Sub256, 3, 2, 1, true},
		{"Sub", overflow.Sub256, min, 1, max, false},
		{"Sub", overflow.Sub256, max, -1, min, false},
		{"Sub", overflow.Sub256, 0, min, min, false},
		{"Sub", overflow.Sub256, -1, min, max, true},
		{"Mul", overflow.Mul256, 0, max, 0, true},
		{"Mul", overflow.Mul256, -1, max, -max, true},
		{"Mul", overflow.Mul256, -1, min, min, false},
		{"Mul", overflow.Mul256, min, -1, min, false},
		{"Mul", overflow.Mul256, pow2(127), pow2(127), pow2(254), true},
		{"Mul", overflow.Mul256, pow2(128), pow2(127), min, false},
		{"Mul", overflow.Mul256, -pow2(128), pow2(127), min, true},
		{"Div", overflow.Div256, 7, 2, 3, true},
		{"Div", overflow.Div256, -7, 2, -3, true},
		{"Div", overflow.Div256, min, -1, min, false},
		{"Div", overflow.Div256, min, 1, min, true},
		{"Div", overflow.Div256, 1, 0, 0, false},
	}
	for _, tt := range tests {
		got, ok := tt.fn(tt.a, tt.b)
		// On failure the result is still defined: the wrapped value, or 0
		// for a zero divisor.
		if ok != tt.ok || got != tt.want {
			t.Errorf("%s256(%d, %d) = %d, %v; want %d, %v", tt.name, tt.a, tt.b, got, ok, tt.want, tt.ok)
		}
	}
}

func TestUint256(t *testing.T) {
	const max = math.MaxUint256
	type op func(a, b uint256) (uint256, bool)
	tests := []struct {
		name string
		fn   op
		a, b uint256
		want uint256
		ok   bool
	}{
		{"Add", overflow.Addu256, 1, 2, 3, true},
		{"Add", overflow.Addu256, max, 0, max, true},
		{"Add", overflow.Addu256, max, 1, 0, false},
		{"Add", overflow.Addu256, 1 << 255, 1 << 255, 0, false},
		{"Sub", overflow.Subu256, 3, 2, 1, true},
		{"Sub", overflow.Subu256, 0, 1, max, false},
		{"Sub", overflow.Subu256, max, max, 0, true},
		{"Mul", overflow.Mulu256, 0, max, 0, true},
		{"Mul", overflow.Mulu256, 1 << 128, 1 << 127, 1 << 255, true},
		{"Mul", overflow.Mulu256, 1 << 128, 1 << 128, 0, false},
		{"Mul", overflow.Mulu256, max, 2, max - 1, false},
		{"Div", overflow.Divu256, max, 2, max / 2, true},
		{"Div", overflow.Divu256, 1, 0, 0, false},
	}
	for _, tt := range tests {
		got, ok := tt.fn(tt.a, tt.b)
		// On failure the result is still defined: the wrapped value, or 0
		// for a zero divisor.
		if ok != tt.ok || got != tt.want {
			t.Errorf("%su256(%d, %d) = %d, %v; want %d, %v", tt.name, tt.a, tt.b, got, ok, tt.want, tt.ok)
		}
	}
}

func Test256Panics(t *testing.T) {
	mustPanic := func(name string, fn func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s: expected panic", name)
			}
		}()
		fn()
	}
	mustPanic("Add256p", func() { overflow.Add256p(i256Max, 1) })
	mustPanic("Sub256p", func() { overflow.Sub256p(i256Min, 1) })
	mustPanic("Mul256p", func() { overflow.Mul256p(i256Min, -1) })
	mustPanic("Div256p", func() { overflow.Div256p(i256Min, -1) })
	mustPanic("Addu256p", func() { overflow.Addu256p(math.MaxUint256, 1) })
	mustPanic("Subu256p", func() { overflow.Subu256p(0, 1) })
	mustPanic("Mulu256p", func() { overflow.Mulu256p(1<<200, 1<<56) })
	mustPanic("Divu256p", func() { overflow.Divu256p(1, 0) })

	if got := overflow.Add256p(1, 2); got != 3 {
		t.Errorf("Add256p(1, 2) = %d", got)
	}
	if got := overflow.Mulu256p(1<<200, 1<<55); got != 1<<255 {
		t.Errorf("Mulu256p(1<<200, 1<<55) = %d", got)
	}
}

func TestConvert256(t *testing.T) {
	var u64max uint256 = math.MaxUint64
	checkI64 := func(name string, got int64, ok bool, want int64, wok bool) {
		if got != want || ok != wok {
			t.Errorf("%s = %d, %v; want %d, %v", name, got, ok, want, wok)
		}
	}
	checkU64 := func(name string, got uint64, ok bool, want uint64, wok bool) {
		if got != want || ok != wok {
			t.Errorf("%s = %d, %v; want %d, %v", name, got, ok, want, wok)
		}
	}

	// The wrapping conversion the reviewer pointed at: 1<<200 + 1 keeps its
	// low bits, so it reads as 1, and the check reports it.
	var big uint256 = 1<<200 + 1
	got, ok := overflow.Uint256ToInt64(big)
	checkI64("Uint256ToInt64(1<<200+1)", got, ok, 1, false)
	got, ok = overflow.Uint256ToInt64(math.MaxInt64)
	checkI64("Uint256ToInt64(MaxInt64)", got, ok, math.MaxInt64, true)
	got, ok = overflow.Uint256ToInt64(math.MaxInt64 + 1)
	checkI64("Uint256ToInt64(MaxInt64+1)", got, ok, math.MinInt64, false)

	ugot, ok := overflow.Uint256ToUint64(u64max)
	checkU64("Uint256ToUint64(MaxUint64)", ugot, ok, math.MaxUint64, true)
	ugot, ok = overflow.Uint256ToUint64(u64max + 1)
	checkU64("Uint256ToUint64(MaxUint64+1)", ugot, ok, 0, false)

	var minI64 int256 = math.MinInt64
	var maxI64 int256 = math.MaxInt64
	got, ok = overflow.Int256ToInt64(minI64)
	checkI64("Int256ToInt64(MinInt64)", got, ok, math.MinInt64, true)
	got, ok = overflow.Int256ToInt64(minI64 - 1)
	checkI64("Int256ToInt64(MinInt64-1)", got, ok, math.MaxInt64, false)
	got, ok = overflow.Int256ToInt64(maxI64 + 1)
	checkI64("Int256ToInt64(MaxInt64+1)", got, ok, math.MinInt64, false)
	got, ok = overflow.Int256ToInt64(i256Min)
	checkI64("Int256ToInt64(MinInt256)", got, ok, 0, false)

	ugot, ok = overflow.Int256ToUint64(int256(u64max))
	checkU64("Int256ToUint64(MaxUint64)", ugot, ok, math.MaxUint64, true)
	ugot, ok = overflow.Int256ToUint64(int256(u64max + 1))
	checkU64("Int256ToUint64(MaxUint64+1)", ugot, ok, 0, false)
	ugot, ok = overflow.Int256ToUint64(-1)
	checkU64("Int256ToUint64(-1)", ugot, ok, math.MaxUint64, false)

	if r, ok := overflow.Int256ToUint256(i256Max); r != maxBits || !ok {
		t.Errorf("Int256ToUint256(MaxInt256) = %d, %v", r, ok)
	}
	if r, ok := overflow.Int256ToUint256(-1); r != math.MaxUint256 || ok {
		t.Errorf("Int256ToUint256(-1) = %d, %v", r, ok)
	}
	if r, ok := overflow.Uint256ToInt256(maxBits); r != i256Max || !ok {
		t.Errorf("Uint256ToInt256(MaxInt256) = %d, %v", r, ok)
	}
	if r, ok := overflow.Uint256ToInt256(maxBits + 1); r != i256Min || ok {
		t.Errorf("Uint256ToInt256(MaxInt256+1) = %d, %v", r, ok)
	}
}

func TestConvert256Panics(t *testing.T) {
	mustPanic := func(name string, fn func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("%s: expected panic", name)
			}
		}()
		fn()
	}
	var u64max uint256 = math.MaxUint64
	mustPanic("Uint256ToInt64p", func() { overflow.Uint256ToInt64p(1<<200 + 1) })
	mustPanic("Uint256ToUint64p", func() { overflow.Uint256ToUint64p(u64max + 1) })
	mustPanic("Uint256ToInt256p", func() { overflow.Uint256ToInt256p(maxBits + 1) })
	mustPanic("Int256ToInt64p", func() { overflow.Int256ToInt64p(i256Min) })
	mustPanic("Int256ToUint64p", func() { overflow.Int256ToUint64p(-1) })
	mustPanic("Int256ToUint256p", func() { overflow.Int256ToUint256p(-1) })

	if got := overflow.Uint256ToInt64p(42); got != 42 {
		t.Errorf("Uint256ToInt64p(42) = %d", got)
	}
	if got := overflow.Int256ToUint256p(i256Max); got != maxBits {
		t.Errorf("Int256ToUint256p(MaxInt256) = %d", got)
	}
}
