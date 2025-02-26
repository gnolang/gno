package pool

import (
	"testing"

	"gno.land/p/demo/ufmt"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestLiquidityMathAddDelta(t *testing.T) {
	tests := []struct {
		name      string
		fn        func()
		wantPanic string
	}{
		{
			name: "x is nil",
			fn: func() {
				var y *i256.Int
				y = i256.MustFromDecimal("100")
				liquidityMathAddDelta(nil, y)
			},
			wantPanic: newErrorWithDetail(errInvalidInput, "x or y is nil"),
		},
		{
			name: "y is nil",
			fn: func() {
				var x *u256.Uint
				x = u256.MustFromDecimal("100")
				liquidityMathAddDelta(x, nil)
			},
			wantPanic: newErrorWithDetail(errInvalidInput, "x or y is nil"),
		},
		{
			name: "underflow panic with sub delta",
			fn: func() {
				x := u256.NewUint(0)
				y := i256.MustFromDecimal("-100")
				liquidityMathAddDelta(x, y)
			},
			wantPanic: newErrorWithDetail(
				errLiquidityCalculation,
				ufmt.Sprintf("Condition failed: (z must be < x) (x: 0, y: -100, z:115792089237316195423570985008687907853269984665640564039457584007913129639836)")),
		},
		{
			name: "overflow panic with add delta",
			fn: func() {
				x := u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935") // 2^256 - 1
				y := i256.MustFromDecimal("100")
				liquidityMathAddDelta(x, y)
			},
			wantPanic: newErrorWithDetail(
				errLiquidityCalculation,
				ufmt.Sprintf("Condition failed: (z must be >= x) (x: 115792089237316195423570985008687907853269984665640564039457584007913129639935, y: 100, z:99)")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Errorf("%s: expected panic but fot none", tt.name)
					return
				}
				if r.(string) != tt.wantPanic {
					t.Errorf("%s: got panic %v, want %v", tt.name, r, tt.wantPanic)
				}
			}()

			tt.fn()
		})
	}
}
