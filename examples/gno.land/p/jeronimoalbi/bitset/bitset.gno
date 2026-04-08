package bitset

import (
	"strconv"
	"strings"
)

const bitsPerWord = 64

// BitSet implements an arbitrary-size bit array.
type BitSet struct {
	words []uint
}

// New creates a new BitSet pre-allocated to hold at least size bits.
func New(size uint) BitSet {
	if size == 0 {
		return BitSet{}
	}
	return BitSet{words: make([]uint, wordIndex(size-1)+1)}
}

// Set turns on the bit at a given position.
func (b *BitSet) Set(i uint) {
	idx := wordIndex(i)
	if idx >= len(b.words) {
		b.grow(idx)
	}
	b.words[idx] |= bitMask(i) // Turn bit on
}

// Clear turns off the bit at a given position.
func (b *BitSet) Clear(i uint) {
	idx := wordIndex(i)
	if idx < len(b.words) {
		b.words[idx] &^= bitMask(i) // Turn bit off
	}
}

// ClearAll turns off all bits.
func (b *BitSet) ClearAll() {
	for i := range b.words {
		b.words[i] = 0
	}
}

// Test reports whether the bit at a given position is set.
func (b BitSet) Test(i uint) bool {
	idx := wordIndex(i)
	if idx >= len(b.words) {
		return false // Bit has not been setted yet
	}
	return b.words[idx]&bitMask(i) != 0
}

// Size returns the total number of bits (including unset) currently stored.
// It's the total space allocated within the set in bits.
func (b BitSet) Size() int {
	return len(b.words) * bitsPerWord
}

// Len returns the number of set bits.
func (b BitSet) Len() int {
	var count int
	for _, w := range b.words {
		count += countSetBits(w)
	}
	return count
}

// And performs an in-place AND with other bitset.
// Bits beyond the other bitset are cleared.
func (b *BitSet) And(other BitSet) {
	for i := range b.words {
		if i < len(other.words) {
			b.words[i] &= other.words[i]
		} else {
			b.words[i] = 0
		}
	}
}

// Or performs an in-place OR (union) with other bitset.
// Current bitset grows if the other bitset is bigger.
func (b *BitSet) Or(other BitSet) {
	if len(other.words) > len(b.words) {
		b.grow(len(other.words) - 1)
	}

	for i := range other.words {
		b.words[i] |= other.words[i]
	}
}

// Xor performs an in-place XOR with other.
// Current bitset grows if the other bitset is bigger.
func (b *BitSet) Xor(other BitSet) {
	if len(other.words) > len(b.words) {
		b.grow(len(other.words) - 1)
	}

	for i := range other.words {
		b.words[i] ^= other.words[i]
	}
}

// Equals checks whether other bitset have exactly the same bits set.
func (b BitSet) Equals(other BitSet) bool {
	short, long := b.words, other.words
	if len(short) > len(long) {
		short, long = long, short
	}

	for i := range short {
		if short[i] != long[i] {
			return false
		}
	}

	// Any trailing words in the longer set must be zero
	for i := len(short); i < len(long); i++ {
		if long[i] != 0 {
			return false
		}
	}
	return true
}

// String returns a binary representation of the bitset.
// Representation uses MSB-first binary representation.
func (b BitSet) String() string {
	var s string
	for i := len(b.words) - 1; i >= 0; i-- {
		s += strconv.FormatUint(uint64(b.words[i]), 2)
	}
	return s
}

// PaddedString returns a binary representation of the bitset with zero padding.
// Representation uses MSB-first binary representation.
func (b BitSet) PaddedString() string {
	var s string
	for i := len(b.words) - 1; i >= 0; i-- {
		// Append zero padding when needed to complete word size
		word := strconv.FormatUint(uint64(b.words[i]), 2)
		if len(word) < bitsPerWord {
			s += strings.Repeat("0", bitsPerWord-len(word))
		}

		s += strconv.FormatUint(uint64(b.words[i]), 2)
	}
	return s
}

func (b *BitSet) grow(idx int) {
	words := make([]uint, idx+1)
	copy(words, b.words)
	b.words = words
}

func wordIndex(i uint) int {
	return int(i / bitsPerWord)
}

func bitMask(i uint) uint {
	return 1 << (i % bitsPerWord)
}

func countSetBits(x uint) int {
	// Count the number of set bits using Kernighan's algorithm
	var count int
	for x != 0 {
		x &= x - 1
		count++
	}
	return count
}
