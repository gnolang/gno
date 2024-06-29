package bitmap

import (
	"gno.land/p/demo/ufmt"
)

// A simple implementation of Bitmap

// Bitmap represents a bitmap using a slice of bytes
type Bitmap struct {
	data []byte
}

// NewBitmap creates a new Bitmap with a specific size (in bits)
func New(size uint64) *Bitmap {
	byteSize := (size + 7) / 8 // Calculate the number of bytes needed
	return &Bitmap{
		data: make([]byte, byteSize),
	}
}

// MustSet sets the bit at the given index (0-based) and returns an error if out of bounds
func (b *Bitmap) Set(index uint64) error {
	if index >= uint64(len(b.data))*8 {
		return ufmt.Errorf("Index out of bounds")
	}

	byteIndex := index / 8
	bitIndex := index % 8

	b.data[byteIndex] |= 1 << uint(bitIndex) // Set the corresponding bit using bitwise OR

	return nil
}

// Set sets the bit at the given index (0-based)
func (b *Bitmap) MustSet(index uint64) {
	if index >= uint64(len(b.data))*8 {
		panic("Index out of bounds")
	}

	byteIndex := index / 8
	bitIndex := index % 8

	b.data[byteIndex] |= 1 << uint(bitIndex) // Set the corresponding bit using bitwise OR
}

// Get checks if the bit at the given index is set
func (b *Bitmap) Get(index uint64) bool {
	if index >= uint64(len(b.data))*8 {
		panic("Index out of bounds")
	}

	byteIndex := index / 8
	bitIndex := index % 8

	return b.data[byteIndex]&(1<<uint(bitIndex)) > 0 // Check if the corresponding bit is set using bitwise AND
}
