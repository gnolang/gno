package common

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

var (
	tickRatioTree = NewTickRatioTree()
	binaryLogTree = NewBinaryLogTree()

	shift1By32Left = u256.MustFromDecimal("4294967296") // (1 << 32)
)

// TreeValue wraps u256.Uint to implement custom value type for avl.Tree
type TreeValue struct {
	value *u256.Uint
}

func (tv TreeValue) String() string {
	return tv.value.ToString()
}

// TickRatioTree manages tick ratios
type TickRatioTree struct {
	tree *avl.Tree
}

// NewTickRatioTree initializes a new TickRatioTree with predefined values
func NewTickRatioTree() *TickRatioTree {
	tree := avl.NewTree()

	ratios := []struct {
		key   int32
		value string
	}{
		{0x1, "340265354078544963557816517032075149313"},
		{0x2, "340248342086729790484326174814286782778"},
		{0x4, "340214320654664324051920982716015181260"},
		{0x8, "340146287995602323631171512101879684304"},
		{0x10, "340010263488231146823593991679159461444"},
		{0x20, "339738377640345403697157401104375502016"},
		{0x40, "339195258003219555707034227454543997025"},
		{0x80, "338111622100601834656805679988414885971"},
		{0x100, "335954724994790223023589805789778977700"},
		{0x200, "331682121138379247127172139078559817300"},
		{0x400, "323299236684853023288211250268160618739"},
		{0x800, "307163716377032989948697243942600083929"},
		{0x1000, "277268403626896220162999269216087595045"},
		{0x2000, "225923453940442621947126027127485391333"},
		{0x4000, "149997214084966997727330242082538205943"},
		{0x8000, "66119101136024775622716233608466517926"},
		{0x10000, "12847376061809297530290974190478138313"},
		{0x20000, "485053260817066172746253684029974020"},
		{0x40000, "691415978906521570653435304214168"},
		{0x80000, "1404880482679654955896180642"},
	}

	for _, ratio := range ratios {
		tick := ufmt.Sprintf("%d", ratio.key)
		value := TreeValue{u256.MustFromDecimal(ratio.value)}
		tree.Set(tick, value)
	}

	return &TickRatioTree{tree}
}

// GetRatio retrieves the ratio for a given tick
func (t *TickRatioTree) GetRatio(key int32) (*u256.Uint, bool) {
	strKey := ufmt.Sprintf("%d", key)
	value, exists := t.tree.Get(strKey)

	if !exists {
		return nil, false
	}

	tv, ok := value.(TreeValue)
	if ok {
		return tv.value, true
	}
	println("[TickRatioTree.GetRatio] value is not TreeValue")

	return nil, false
}

// BinaryLogTree manages binary log constants
type BinaryLogTree struct {
	tree *avl.Tree
}

// NewBinaryLogTree initializes a new BinaryLogTree
func NewBinaryLogTree() *BinaryLogTree {
	tree := avl.NewTree()

	logs := [8]string{
		"0",
		"3",
		"15",
		"255",
		"65535",
		"4294967295",
		"18446744073709551615",
		"340282366920938463463374607431768211455",
	}

	for i, value := range logs {
		key := ufmt.Sprintf("%d", i)
		tree.Set(key, TreeValue{u256.MustFromDecimal(value)})
	}

	return &BinaryLogTree{tree}
}

// GetLog retrieves the binary log constant at given index
func (t *BinaryLogTree) GetLog(idx int) (*u256.Uint, bool) {
	strKey := ufmt.Sprintf("%d", idx)
	value, exists := t.tree.Get(strKey)

	if !exists {
		return nil, false
	}

	if tv, ok := value.(TreeValue); ok {
		return tv.value, true
	}

	return nil, false
}

// TickMathGetSqrtRatioAtTick calculates the square root price ratio for a given tick.
//
// This function computes the square root ratio (sqrt(price)) at a specific tick,
// using a precomputed mapping of ratios. The result is returned as a 160-bit
// fixed-point value (Q64.96 format).
//
// Parameters:
//   - tick (int32): The tick index for which the square root ratio is calculated.
//
// Returns:
//   - *u256.Uint: The square root price ratio at the given tick, represented as a 160-bit unsigned integer.
//
// Behavior:
//  1. Validates that the tick is within the acceptable range by asserting its absolute value.
//  2. Initializes the ratio based on whether the least significant bit of the tick is set, using a lookup table (`tickRatioMap`).
//  3. Iteratively adjusts the ratio by multiplying and right-shifting with precomputed values for each relevant bit set in the tick value.
//  4. If the tick is positive, the ratio is inverted by dividing a maximum uint256 value by the computed ratio.
//  5. The result is split into upper 128 bits and lower 32 bits for precision handling.
//  6. If the lower 32 bits are non-zero, the upper 128 bits are incremented by 1 to ensure rounding up.
//
// Example:
//   - For a tick of `0`, the ratio represents `1.0000` in Q64.96 format.
//   - For a tick of `-887272` (minimum tick), the ratio represents the smallest possible price.
//   - For a tick of `887272` (maximum tick), the ratio represents the highest possible price.
//
// Panics:
//   - If the absolute tick value exceeds the maximum allowed tick range.
//
// Notes:
//   - The function relies on a precomputed map `tickRatioMap` to optimize calculations.
//   - Handles rounding by adding 1 if the remainder of the division is non-zero.
func TickMathGetSqrtRatioAtTick(tick int32) *u256.Uint { // uint160 sqrtPriceX96
	assertValidTickRange(tick)
	absTick := abs(tick)

	ratio := u256.MustFromDecimal(consts.Q128)
	initialBit := int32(0x1)

	if val, exists := tickRatioTree.GetRatio(initialBit); exists && (absTick&initialBit) != 0 {
		ratio = val
	}

	calculateRatio := func(mask int32) *u256.Uint {
		value, exists := tickRatioTree.GetRatio(mask)
		if exists && absTick&mask != 0 {
			return new(u256.Uint).Rsh(
				new(u256.Uint).Mul(ratio, value),
				128,
			)
		} else if exists == false {
			println("[TickMathGetSqrtRatioAtTick] ratio not found for mask:", mask)
		}
		return ratio
	}

	for mask := int32(0x2); mask <= 0x80000; mask *= 2 {
		ratio = calculateRatio(mask)
	}

	if tick > 0 {
		maxUint256 := u256.MustFromDecimal(consts.MAX_UINT256)
		if ratio.IsZero() {
			println("[TickMathGetSqrtRatioAtTick] ratio is zero")
			return u256.Zero()
		}
		ratio = new(u256.Uint).Div(maxUint256, ratio)
	}

	// direct modification of ratio objects is restricted
	ratioClone := ratio.Clone()
	upper128Bits := ratioClone.Rsh(ratioClone, 32)  // ratio >> 32
	lower32Bits := ratio.Mod(ratio, shift1By32Left) // ratio % (1 << 32)

	var roundUp *u256.Uint
	if lower32Bits.IsZero() {
		roundUp = u256.Zero()
	} else {
		roundUp = u256.One()
	}

	return new(u256.Uint).Add(upper128Bits, roundUp)
}

func TickMathGetTickAtSqrtRatio(sqrtPriceX96 *u256.Uint) int32 {
	sqrtPrice := new(u256.Uint).Set(sqrtPriceX96)
	isAboveMin := sqrtPrice.Gte(u256.MustFromDecimal(consts.MIN_SQRT_RATIO))
	isBelowMax := sqrtPrice.Lt(u256.MustFromDecimal(consts.MAX_SQRT_RATIO))
	if !(isAboveMin && isBelowMax) {
		panic(newErrorWithDetail(
			errOutOfRange,
			ufmt.Sprintf("sqrtPriceX96(%s) is out of range", sqrtPrice.ToString()),
		))
	}

	scaledPrice := new(u256.Uint).Lsh(sqrtPrice, 32)
	leadingBitPosition, normalizedPrice := findMSB(scaledPrice)
	normalizedPrice = adjustRatio(scaledPrice, leadingBitPosition)
	logBase2Value := calculateLog2(leadingBitPosition, normalizedPrice)
	tickValue := getTickValue(logBase2Value, sqrtPrice)

	return tickValue
}

// findMSB computes the MSB (most significant bit) of the given ratio.
func findMSB(ratio *u256.Uint) (*u256.Uint, *u256.Uint) {
	msb := u256.Zero()

	calculateMSB := func(i int) (*u256.Uint, *u256.Uint) {
		if logConst, exists := binaryLogTree.GetLog(i); exists {
			f := new(u256.Uint).Lsh(gt(ratio, logConst), uint(i))
			msb = new(u256.Uint).Or(msb, f)
			ratio = new(u256.Uint).Rsh(ratio, uint(f.Uint64()))
		}
		return msb, ratio
	}

	for i := 7; i >= 1; i-- {
		msb, ratio = calculateMSB(i)
	}

	// handle the remaining bits
	{
		f := gt(ratio, u256.One()) // 0x1
		// msb = msb | f
		msb = new(u256.Uint).Or(msb, f)
	}

	return msb, ratio
}

// adjustRatio adjusts the given ratio based on the MSB found.
//
// This adjustment ensures that the ratio falls within the specific range.
func adjustRatio(ratio, msb *u256.Uint) *u256.Uint {
	if msb.Gte(u256.NewUint(128)) {
		return new(u256.Uint).Rsh(ratio, uint(msb.Uint64()-127))
	}

	return new(u256.Uint).Lsh(ratio, uint(127-msb.Uint64()))
}

// calculateLog2 calculates the binary logarith, of the adjusted ratio using a fixed-point arithmetic.
//
// This function iteratively squares the ratio and adjusts the result to compute the log base 2, which will determine the tick value.
func calculateLog2(msb, ratio *u256.Uint) *i256.Int {
	msbInt := i256.FromUint256(msb)
	constant128 := i256.NewInt(128)

	log2Result := i256.Zero().Sub(msbInt, constant128)
	log2Result = log2Result.Lsh(log2Result, 64)

	for bitPosition := 63; bitPosition >= 51; bitPosition-- {
		ratio = new(u256.Uint).Mul(ratio, ratio)
		ratio = ratio.Rsh(ratio, 127)

		shiftFactor := i256.FromUint256(new(u256.Uint).Rsh(ratio, 128))

		// log2Result = log2Result | (shiftFactor << bitPosition)
		log2Result = i256.Zero().Or(log2Result, i256.Zero().Lsh(shiftFactor, uint(bitPosition)))

		// ratio = ratio >> uint64(shiftFactor)
		ratio = ratio.Rsh(ratio, uint(shiftFactor.Uint64()))
	}

	// handle the remaining bits
	{
		// ratio = ratio * ratio >> 127
		ratio = new(u256.Uint).Mul(ratio, ratio)
		ratio = new(u256.Uint).Rsh(ratio, 127)

		shiftFactor := i256.FromUint256(new(u256.Uint).Rsh(ratio, 128))

		log2Result = i256.Zero().Or(log2Result, i256.Zero().Lsh(shiftFactor, 50))
	}

	return log2Result
}

// getTickValue determines the tick value corresponding to a given sqrtPriveX96.
//
// It calculates the upper and lower bounds for each tick, and selects the appropriate tock value
// based on the given sqrtPriceX96.
func getTickValue(log2 *i256.Int, sqrtPriceX96 *u256.Uint) int32 {
	// ref: https://github.com/Uniswap/v3-core/issues/500
	// 2^64 / log2 (√1.0001) = 255738958999603826347141
	logSqrtRatioMultiplier := i256.Zero().Mul(log2, i256.MustFromDecimal("255738958999603826347141"))

	// ref: https://ethereum.stackexchange.com/questions/113844/how-does-uniswap-v3s-logarithm-library-tickmath-sol-work/113912#113912
	// 0.010000497 x 2^128 = 3402992956809132418596140100660247210
	tickLow256 := i256.Zero().Sub(logSqrtRatioMultiplier, i256.MustFromDecimal("3402992956809132418596140100660247210"))
	tickLow256 = tickLow256.Rsh(tickLow256, 128)
	tickLow := int32(tickLow256.Int64())

	// ref: https://ethereum.stackexchange.com/questions/113844/how-does-uniswap-v3s-logarithm-library-tickmath-sol-work/113912#113912
	// 0.856 x 2^128 = 291339464771989622907027621153398088495
	tickHi256 := i256.Zero().Add(logSqrtRatioMultiplier, i256.MustFromDecimal("291339464771989622907027621153398088495"))
	tickHi256 = tickHi256.Rsh(tickHi256, 128)
	tickHi := int32(tickHi256.Int64())

	var tick int32
	if tickLow == tickHi {
		tick = tickLow
	} else if TickMathGetSqrtRatioAtTick(tickHi).Lte(sqrtPriceX96) {
		tick = tickHi
	} else {
		tick = tickLow
	}

	return tick
}

func gt(x, y *u256.Uint) *u256.Uint {
	if x.Gt(y) {
		return u256.One()
	}

	return u256.Zero()
}

// abs returns the absolute value of the given integer.
func abs(x int32) int32 {
	if x < 0 {
		return -x
	}

	return x
}

// assertValidTickRange validates that the absolute tick value is within the acceptable range.
func assertValidTickRange(tick int32) {
	if tick > consts.MAX_TICK {
		panic(newErrorWithDetail(
			errOutOfRange,
			ufmt.Sprintf("tick is out of range (larger than 887272), tick: %d", tick),
		))
	}
	if tick < consts.MIN_TICK {
		panic(newErrorWithDetail(
			errOutOfRange,
			ufmt.Sprintf("tick is out of range (smaller than -887272), tick: %d", tick),
		))
	}
}
