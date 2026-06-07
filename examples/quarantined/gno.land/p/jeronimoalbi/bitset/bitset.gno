package bitset

import "strings"

const bitsPerWord = 64

// BitSet implements an arbitrary-size bit array.
type BitSet struct {
	words []uint64
}

// New creates a new BitSet pre-allocated to hold at least size bits.
func New(size uint64) BitSet {
	if size == 0 {
		return BitSet{}
	}
	return BitSet{words: make([]uint64, wordIndex(size-1)+1)}
}

// Set turns on the bit at a given position.
func (b *BitSet) Set(i uint64) {
	idx := wordIndex(i)
	if idx >= len(b.words) {
		b.grow(idx + 1)
	}
	b.words[idx] |= bitMask(i) // Turn bit on
}

// Clear turns off the bit at a given position.
func (b *BitSet) Clear(i uint64) {
	idx := wordIndex(i)
	if idx < len(b.words) {
		b.words[idx] &^= bitMask(i) // Turn bit off
	}
}

// ClearAll turns off all bits.
// The size of the bitset remains unchanged.
func (b *BitSet) ClearAll() {
	for i := range b.words {
		b.words[i] = 0
	}
}

// Compact reclaims memory by removing trailing zero words.
func (b *BitSet) Compact() {
	end := len(b.words)
	for ; end > 0 && b.words[end-1] == 0; end-- {
	}
	words := make([]uint64, end)
	copy(words, b.words)
	b.words = words
}

// IsSet checks whether the bit at a given position is set.
func (b BitSet) IsSet(i uint64) bool {
	idx := wordIndex(i)
	if idx >= len(b.words) {
		return false // Bit has not been set yet
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
	otherLen := len(other.words)
	if otherLen > len(b.words) {
		b.grow(otherLen)
	}

	for i := range other.words {
		b.words[i] |= other.words[i]
	}
}

// Xor performs an in-place XOR with other.
// Current bitset grows if the other bitset is bigger.
func (b *BitSet) Xor(other BitSet) {
	otherLen := len(other.words)
	if otherLen > len(b.words) {
		b.grow(otherLen)
	}

	for i := range other.words {
		b.words[i] ^= other.words[i]
	}
}

// Equal checks whether other bitset have exactly the same bits set.
func (b BitSet) Equal(other BitSet) bool {
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

// PaddedString returns a binary representation of the bitset with zero padding.
// Representation uses MSB-first binary representation.
func (b BitSet) PaddedString() string {
	wordCount := len(b.words)
	if wordCount == 0 {
		return ""
	}

	buf := make([]byte, wordCount*bitsPerWord)
	for i, w := range b.words {
		base := (wordCount - 1 - i) * bitsPerWord
		for j := 0; j < bitsPerWord; j++ {
			if w>>uint64(bitsPerWord-1-j)&1 == 1 {
				buf[base+j] = '1'
			} else {
				buf[base+j] = '0'
			}
		}
	}
	return string(buf)
}

// String returns a binary representation of the bitset.
// Representation uses MSB-first binary representation.
func (b BitSet) String() string {
	s := strings.TrimLeft(b.PaddedString(), "0")
	if s == "" && len(b.words) > 0 {
		return "0"
	}
	return s
}

func (b *BitSet) grow(length int) {
	words := make([]uint64, length)
	copy(words, b.words)
	b.words = words
}

func wordIndex(i uint64) int {
	return int(i / bitsPerWord)
}

func bitMask(i uint64) uint64 {
	return 1 << (i % bitsPerWord)
}

func countSetBits(x uint64) int {
	// Count the number of set bits using Kernighan's algorithm
	var count int
	for x != 0 {
		x &= x - 1
		count++
	}
	return count
}
