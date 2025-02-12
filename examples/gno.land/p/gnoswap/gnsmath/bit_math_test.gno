package gnsmath

import (
	"testing"

	"gno.land/p/demo/uassert"

	u256 "gno.land/p/gnoswap/uint256"
)

func TestBitMathMostSignificantBit(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		uassert.PanicsWithMessage(t, errMSBZeroInput.Error(), func() {
			BitMathMostSignificantBit(u256.Zero())
		})
	})

	t.Run("1", func(t *testing.T) {
		uassert.Equal(t, BitMathMostSignificantBit(u256.One()), uint8(0))
	})

	t.Run("2", func(t *testing.T) {
		uassert.Equal(t, BitMathMostSignificantBit(u256.NewUint(2)), uint8(1))
	})

	t.Run("all powers of 2", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			num := u256.Zero()
			num.Lsh(u256.One(), uint(i))
			uassert.Equal(t, BitMathMostSignificantBit(num), uint8(i))
		}
	})

	t.Run("uint256(-1)", func(t *testing.T) {
		// BigNumber.from(2).pow(256).sub(1))
		uassert.Equal(t, BitMathMostSignificantBit(u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935")), uint8(255))
	})
}

func TestBitMathLeastSignificantBit(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		uassert.PanicsWithMessage(t, errLSBZeroInput.Error(), func() {
			BitMathLeastSignificantBit(u256.Zero())
		})
	})

	t.Run("1", func(t *testing.T) {
		uassert.Equal(t, BitMathLeastSignificantBit(u256.One()), uint8(0))
	})

	t.Run("2", func(t *testing.T) {
		uassert.Equal(t, BitMathLeastSignificantBit(u256.NewUint(2)), uint8(1))
	})

	t.Run("all powers of 2", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			num := u256.Zero()
			num.Lsh(u256.One(), uint(i))
			uassert.Equal(t, BitMathLeastSignificantBit(num), uint8(i))
		}
	})

	t.Run("uint256(-1)", func(t *testing.T) {
		// BigNumber.from(2).pow(256).sub(1))
		uassert.Equal(t, BitMathLeastSignificantBit(u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935")), uint8(0))
	})
}
