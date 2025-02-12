package uint256

import (
	"testing"
)

func TestMulDiv_1(t *testing.T) {
	// reverts if denominator is 0
	Q128 := MustFromDecimal("340282366920938463463374607431768211456") // 2**128

	x := NewUint(5)
	y := Zero()

	shouldPanic(
		t,
		func() {
			MulDiv(Q128, x, y)
		},
	)
}

func TestMulDiv_2(t *testing.T) {
	// reverts if denominator is 0 and numerator overflows
	Q128 := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	x := NewUint(5)
	y := Zero()

	shouldPanic(
		t,
		func() {
			MulDiv(Q128, Q128, y)
		},
	)
}

func TestMulDiv_3(t *testing.T) {
	// reverts if output overflows uint256
	Q128 := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	y := One()

	shouldPanic(
		t,
		func() {
			MulDiv(Q128, Q128, y)
		},
	)
}

func TestMulDiv_4(t *testing.T) {
	// reverts on overflow with all max inputs
	MaxUint256 := MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	MaxUint256Sub1 := MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639934")
	if MaxUint256.ToString() != "115792089237316195423570985008687907853269984665640564039457584007913129639935" {
		t.Errorf("MustFromHex is Failed")
	}
	shouldPanic(
		t,
		func() {
			MulDiv(MaxUint256, MaxUint256, MaxUint256Sub1)
		},
	)
}

func TestMulDiv_5(t *testing.T) {
	// all max inputs
	MaxUint256 := MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	got := MulDiv(MaxUint256, MaxUint256, MaxUint256)

	rst := got.Eq(MaxUint256)
	if !rst {
		t.Errorf("MaxUint256 * MaxUint256 / MaxUint256 is not same to MaxUint256")
	}
}

func TestMulDiv_6(t *testing.T) {
	denom := NewUint(100)

	// accurate without phantom overflow #1
	Q128 := MustFromDecimal("340282366920938463463374607431768211456")      // 2**128
	Q128Const := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	expected := Q128.Div(Q128, NewUint(3))
	x := NewUint(50)
	x1 := MulDiv(x, Q128Const, denom)
	y := NewUint(150)
	y1 := MulDiv(y, Q128Const, denom)

	got := MulDiv(Q128Const, x1, y1)

	rst := got.Eq(expected)
	if !rst {
		t.Errorf("Q128/3 is not smae to Q128 * (50*Q128/100) / (150*Q128/100)")
	}
}

func TestMulDiv_6_1(t *testing.T) {
	denom := NewUint(100)

	// accurate without phantom overflow #1_1
	Q128 := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	expected := new(Uint).Div(Q128, NewUint(3))
	x := NewUint(50)
	x1 := MulDiv(x, Q128, denom)
	y := NewUint(150)
	y1 := MulDiv(y, Q128, denom)

	got := MulDiv(Q128, x1, y1)

	rst := got.Eq(expected)
	if !rst {
		t.Errorf("Q128/3 is not smae to Q128 * (50*Q128/100) / (150*Q128/100)")
	}
}

func TestMulDiv_7(t *testing.T) {
	// accurate with phantom overflow #2
	Q128 := MustFromDecimal("340282366920938463463374607431768211456")      // 2**128
	Q128Const := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	x := NewUint(4375)
	expected := MulDiv(x, Q128Const, NewUint(1000))

	y1 := NewUint(35)
	y1.Mul(y1, Q128Const)

	denom := NewUint(8)
	denom.Mul(denom, Q128Const)

	got := MulDiv(Q128Const, y1, denom)

	rst := got.Eq(expected)
	if !rst {
		t.Errorf("4375*Q128/1000 is not smae to Q128*35*Q128/(8*Q128)")
	}
}

func TestMulDiv_8(t *testing.T) {
	// accurate with phantom overflow and repeating decimal
	Q128 := MustFromDecimal("340282366920938463463374607431768211456")      // 2**128
	Q128Const := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	expected := MulDiv(One(), Q128Const, NewUint(3))

	y1 := NewUint(1000)
	y1.Mul(y1, Q128Const)

	denom := NewUint(3000)
	denom.Mul(denom, Q128Const)

	got := MulDiv(Q128Const, y1, denom)

	rst := got.Eq(expected)
	if !rst {
		t.Errorf("1*Q128/3 is not smae to Q128*1000*Q128/(3000*Q128)")
	}
}

func TestMulDivRoundingUp_1(t *testing.T) {
	// reverts if denominator is 0
	Q128 := MustFromDecimal("340282366920938463463374607431768211456") // 2**128

	x := NewUint(5)
	y := Zero()

	shouldPanic(
		t,
		func() {
			MulDivRoundingUp(Q128, x, y)
		},
	)
}

func TestMulDivRoundingUp_2(t *testing.T) {
	// reverts if denominator is 0 and numerator overflows

	Q128 := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	y := Zero()

	shouldPanic(
		t,
		func() {
			MulDivRoundingUp(Q128, Q128, y)
		},
	)
}

func TestMulDivRoundingUp_3(t *testing.T) {
	// reverts if output overflows uint256

	Q128 := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	y := One()

	shouldPanic(
		t,
		func() {
			MulDivRoundingUp(Q128, Q128, y)
		},
	)
}

func TestMulDivRoundingUp_4(t *testing.T) {
	// reverts on overflow with all max inputs

	MaxUint256 := MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	MaxUint256Sub1 := MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639934")

	shouldPanic(
		t,
		func() {
			MulDivRoundingUp(MaxUint256, MaxUint256, MaxUint256Sub1)
		},
	)
}

func TestMulDivRoundingUp_5(t *testing.T) {
	// reverts if mulDiv overflows 256 bits after rounding up

	x := MustFromDecimal("535006138814359")
	y := MustFromDecimal("432862656469423142931042426214547535783388063929571229938474969")

	shouldPanic(
		t,
		func() {
			MulDivRoundingUp(x, y, NewUint(2))
		},
	)
}

func TestMulDivRoundingUp_6(t *testing.T) {
	// reverts if mulDiv overflows 256 bits after rounding up case 2

	x := MustFromDecimal("115792089237316195423570985008687907853269984659341747863450311749907997002549")
	y := MustFromDecimal("115792089237316195423570985008687907853269984659341747863450311749907997002550")
	z := MustFromDecimal("115792089237316195423570985008687907853269984653042931687443039491902864365164")

	shouldPanic(
		t,
		func() {
			MulDivRoundingUp(x, y, z)
		},
	)
}

func TestMulDivRoundingUp_7(t *testing.T) {
	// all max inputs
	MaxUint256 := MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	got := MulDivRoundingUp(MaxUint256, MaxUint256, MaxUint256)

	rst := got.Eq(MaxUint256)

	if !rst {
		t.Errorf("MaxUin256*MaxUint256/MaxUint256 RoudingUp is not same to MaxUint256")
	}
}

func TestMulDivRoundingUp_8(t *testing.T) {
	// accurate without phantom overflow
	denom := NewUint(100)

	// accurate without phantom overflow #1
	Q128 := MustFromDecimal("340282366920938463463374607431768211456")      // 2**128
	Q128Const := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	Q128.Div(Q128, NewUint(3))
	expected := Q128.Add(Q128, NewUint(1))

	x1 := NewUint(50)
	x1.Mul(x1, Q128Const)

	y1 := NewUint(150)
	y1.Mul(y1, Q128Const)

	got := MulDivRoundingUp(Q128Const, x1, y1)

	rst := got.Eq(expected)
	if !rst {
		t.Errorf("Q128*50*Q128/100/(150*Q128/100) should be equal to Q128/3")
	}
}

func TestMulDivRoundingUp_9(t *testing.T) {
	// accurate with phantom overflow #2
	Q128 := MustFromDecimal("340282366920938463463374607431768211456")      // 2**128
	Q128Const := MustFromDecimal("340282366920938463463374607431768211456") // 2**128
	x := NewUint(4375)
	expected := MulDiv(x, Q128Const, NewUint(1000))

	y1 := NewUint(35)
	y1.Mul(y1, Q128Const)

	denom := NewUint(8)
	denom.Mul(denom, Q128Const)

	got := MulDiv(Q128Const, y1, denom)

	rst := got.Eq(expected)
	if !rst {
		t.Errorf("4375*Q128/1000 is not smae to Q128*35*Q128/(8*Q128)")
	}
}

func TestMulDivRoundingUp_10(t *testing.T) {
	// accurate with phantom overflow and repeating decimal

	Q128 := MustFromDecimal("340282366920938463463374607431768211456")      // 2**128
	Q128Const := MustFromDecimal("340282366920938463463374607431768211456") // 2**128

	expected := MulDiv(One(), Q128Const, NewUint(3))
	expected.Add(expected, One())

	y1 := NewUint(1000)
	y1.Mul(y1, Q128Const)

	denom := NewUint(3000)
	denom.Mul(denom, Q128Const)

	got := MulDivRoundingUp(Q128Const, y1, denom)

	rst := got.Eq(expected)
	if !rst {
		t.Errorf("Q128*100*Q128/300*Q128 should be eq to 1*Q128+1")
	}
}
