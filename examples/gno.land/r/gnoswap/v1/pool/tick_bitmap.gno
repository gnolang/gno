package pool

import (
	"strconv"

	"gno.land/p/demo/ufmt"
	plp "gno.land/p/gnoswap/gnsmath"

	u256 "gno.land/p/gnoswap/uint256"
)

// tickBitmapPosition calculates the word and bit position for a given tick
func tickBitmapPosition(tick int32) (int16, uint8) {
	wordPos := int16(tick >> 8) // tick / 256
	bitPos := uint8(tick % 256)
	return wordPos, bitPos
}

// tickBitmapFlipTick flips the state of a tick in the tick bitmap.
//
// This function toggles the "initialized" state of a tick in the tick bitmap.
// It ensures that the tick aligns with the specified tick spacing and then
// flips the corresponding bit in the bitmap representation.
//
// Parameters:
//   - tick: int32, the tick index to toggle.
//   - tickSpacing: int32, the spacing between valid ticks.
//     The tick must align with this spacing.
//
// Workflow:
//  1. Validates that the `tick` aligns with `tickSpacing` using `checkTickSpacing`.
//  2. Computes the position of the bit in the tick bitmap:
//     - `wordPos`: Determines which word in the bitmap contains the bit.
//     - `bitPos`: Identifies the position of the bit within the word.
//  3. Creates a bitmask using `Lsh` (Left Shift) to target the bit at `bitPos`.
//  4. Toggles (flips) the bit using XOR with the current value of the tick bitmap.
//  5. Updates the tick bitmap with the modified word.
//
// Behavior:
//   - If the bit is `0` (uninitialized), it will be flipped to `1` (initialized).
//   - If the bit is `1` (initialized), it will be flipped to `0` (uninitialized).
//
// Example:
//
//	pool.tickBitmapFlipTick(120, 60)
//	// This flips the bit for tick 120 with a tick spacing of 60.
//
// Notes:
//   - The `tick` must be divisible by `tickSpacing`. If not, the function will panic.
func (p *Pool) tickBitmapFlipTick(
	tick int32,
	tickSpacing int32,
) {
	checkTickSpacing(tick, tickSpacing)
	wordPos, bitPos := tickBitmapPosition(tick / tickSpacing)

	mask := new(u256.Uint).Lsh(u256.One(), uint(bitPos))
	p.setTickBitmap(wordPos, new(u256.Uint).Xor(p.getTickBitmap(wordPos), mask))
}

// tickBitmapNextInitializedTickWithInOneWord finds the next initialized tick within
// one word of the bitmap.
func (p *Pool) tickBitmapNextInitializedTickWithInOneWord(
	tick int32,
	tickSpacing int32,
	lte bool,
) (int32, bool) {
	compress := tick / tickSpacing
	if tick < 0 && tick%tickSpacing != 0 {
		compress--
	}

	wordPos, bitPos := getWordAndBitPos(compress, lte)
	mask := getMaskBit(uint(bitPos), lte)
	masked := new(u256.Uint).And(p.getTickBitmap(wordPos), mask)
	initialized := !(masked.IsZero())

	nextTick := getNextTick(lte, initialized, compress, bitPos, tickSpacing, masked)
	return nextTick, initialized
}

// getTickBitmap gets the tick bitmap for the given word position
// if the tick bitmap is not initialized, initialize it to zero
func (p *Pool) getTickBitmap(wordPos int16) *u256.Uint {
	wordPosStr := strconv.Itoa(int(wordPos))

	if !p.tickBitmaps.Has(wordPosStr) {
		p.initTickBitmap(wordPos)
	}

	iU256, exist := p.tickBitmaps.Get(wordPosStr)
	if !exist {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("tickBitmap(%d) does not exist", wordPos),
		))
	}

	return iU256.(*u256.Uint)
}

// setTickBitmap sets the tick bitmap for the given word position
func (p *Pool) setTickBitmap(wordPos int16, tickBitmap *u256.Uint) {
	wordPosStr := strconv.Itoa(int(wordPos))
	p.tickBitmaps.Set(wordPosStr, tickBitmap)
}

// initTickBitmap initializes the tick bitmap for the given word position
func (p *Pool) initTickBitmap(wordPos int16) {
	p.setTickBitmap(wordPos, u256.Zero())
}

// getWordAndBitPos gets tick's wordPos and bitPos depending on the swap direction
func getWordAndBitPos(tick int32, lte bool) (int16, uint8) {
	if lte {
		return tickBitmapPosition(tick)
	}

	tick++
	return tickBitmapPosition(tick)
}

// bMap is a map that maps boolean values to uint values.
// true maps to 1, and false maps to 0.
var bMap = map[bool]uint{
	true:  1,
	false: 0,
}

// getMaskBit generates a mask based on the provided bit position (bitPos) and a boolean flag (lte).
// The function constructs a bitmask with a shift depending on the bit position and the boolean value.
// It either returns the mask or its negation, based on the value of 'lte' (swap direction).
func getMaskBit(bitPos uint, lte bool) *u256.Uint {
	// Shift the number 1 to the left by (bitPos + bMap[lte]) positions.
	// If lte is true, the shift will be bitPos + 1; if false, it will be just bitPos.
	shifted := new(u256.Uint).Lsh(u256.One(), bitPos+bMap[lte])

	// Subtract 1 from the shifted value to create a mask.
	mask := new(u256.Uint).Sub(shifted, u256.One())

	// If lte is false, return the negation of the mask.
	if !lte {
		return new(u256.Uint).Not(mask)
	}
	// Otherwise, return the mask itself.
	return mask
}

// getNextTick gets the next tick depending on the initialized state and the swap direction
func getNextTick(lte, initialized bool, compress int32, bitPos uint8, tickSpacing int32, masked *u256.Uint) int32 {
	if initialized {
		return getTickIfInitialized(compress, tickSpacing, bitPos, masked, lte)
	}

	return getTickIfNotInitialized(compress, tickSpacing, bitPos, lte)
}

// getTickIfInitialized gets the next tick if the tick bitmap is initialized
func getTickIfInitialized(compress, tickSpacing int32, bitPos uint8, masked *u256.Uint, lte bool) int32 {
	if lte {
		return (compress - int32(bitPos-plp.BitMathMostSignificantBit(masked))) * tickSpacing
	}

	return (compress + 1 + int32(plp.BitMathLeastSignificantBit(masked)-bitPos)) * tickSpacing
}

// getTickIfNotInitialized gets the next tick if the tick bitmap is not initialized
func getTickIfNotInitialized(compress, tickSpacing int32, bitPos uint8, lte bool) int32 {
	if lte {
		return (compress - int32(bitPos)) * tickSpacing
	}

	return (compress + 1 + int32(255-bitPos)) * tickSpacing
}
