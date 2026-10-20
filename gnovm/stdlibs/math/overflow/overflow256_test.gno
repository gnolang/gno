package overflow_test

import (
	"math"
	"math/overflow"
	"testing"
)

func TestInt256(t *testing.T) {
	const (
		max = math.MaxInt256
		min = math.MinInt256
	)
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
		{"Mul", overflow.Mul256, 1 << 127, 1 << 127, 1 << 254, true},
		{"Mul", overflow.Mul256, 1 << 128, 1 << 127, min, false},
		{"Mul", overflow.Mul256, -(1 << 128), 1 << 127, min, true},
		{"Div", overflow.Div256, 7, 2, 3, true},
		{"Div", overflow.Div256, -7, 2, -3, true},
		{"Div", overflow.Div256, min, -1, min, false},
		{"Div", overflow.Div256, min, 1, min, true},
		{"Div", overflow.Div256, 1, 0, 0, false},
	}
	for _, tt := range tests {
		got, ok := tt.fn(tt.a, tt.b)
		if ok != tt.ok || (ok && got != tt.want) {
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
		if ok != tt.ok || (ok && got != tt.want) {
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
	mustPanic("Add256p", func() { overflow.Add256p(math.MaxInt256, 1) })
	mustPanic("Sub256p", func() { overflow.Sub256p(math.MinInt256, 1) })
	mustPanic("Mul256p", func() { overflow.Mul256p(math.MinInt256, -1) })
	mustPanic("Div256p", func() { overflow.Div256p(math.MinInt256, -1) })
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
