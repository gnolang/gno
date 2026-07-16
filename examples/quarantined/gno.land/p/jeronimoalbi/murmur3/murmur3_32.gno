package murmur3

import (
	"hash"
	"math/bits"
)

const (
	c1          = 0xcc9e2d51
	c2          = 0x1b873593
	r1          = 15
	r2          = 13
	blockSize32 = 4
	hashSize32  = 4
)

type digest32 struct {
	seed           uint32
	h1             uint32
	tail           [4]byte
	tailBytesCount int
	writtenLen     int
}

// New32 returns a new hash.Hash32 computing the MurmurHash3 hash.
// It can generate 2 to the power of 32 unique 32-bit values ranging
// from 0 to 4_294_967_296.
// Hash is not collision safe, so it's recommended to handle potential
// collisions when dealing with more than 65_536 values which is when
// probability raises.
func New32() hash.Hash32 {
	return &digest32{}
}

// NewWithSeed32 returns a new hash.Hash32 computing the MurmurHash3
// hash with the given seed.
// It can generate 2 to the power of 32 unique 32-bit values ranging
// from 0 to 4_294_967_296.
// Hash is not collision safe, so it's recommended to handle potential
// collisions when dealing with more than 65_536 values which is when
// provability raises.
func NewWithSeed32(seed uint32) hash.Hash32 {
	return &digest32{seed: seed, h1: seed}
}

// Sum32 returns the MurmurHash3 hash of the specified data.
func Sum32(data []byte) uint32 {
	return Sum32WithSeed(data, 0)
}

// Sum32WithSeed returns the MurmurHash3 hash of the specified data
// using the given seed.
func Sum32WithSeed(data []byte, seed uint32) uint32 {
	h := NewWithSeed32(seed)
	h.Write(data)
	return h.Sum32()
}

func (digest32) BlockSize() int {
	return blockSize32
}

func (digest32) Size() int {
	return hashSize32
}

func (d *digest32) Reset() {
	d.h1 = d.seed
	d.tailBytesCount = 0
	d.writtenLen = 0
}

func (d *digest32) Write(bz []byte) (int, error) {
	// Keep track of all written bytes
	total := len(bz)
	d.writtenLen += total

	// If there are leftover bytes in the tail,
	// try to fill a block with the new data.
	if d.tailBytesCount > 0 {
		need := blockSize32 - d.tailBytesCount
		if len(bz) < need {
			// Append more data to the block
			copy(d.tail[d.tailBytesCount:], bz)
			d.tailBytesCount += len(bz)

			// Tail block is still not filled
			return total, nil
		}

		// Append remaining block data and process current block
		copy(d.tail[d.tailBytesCount:], bz[:need])
		d.processBlock(d.tail[:])

		// Update data to exclude the appended processed data
		bz = bz[need:]
		d.tailBytesCount = 0
	}

	// Process full 4-byte blocks
	for len(bz) >= blockSize32 {
		d.processBlock(bz[:blockSize32])
		bz = bz[blockSize32:]
	}

	// Save remaining bytes in the tail. Remaining data can be left
	// out of the last block when data size is less than 4 bytes.
	if len(bz) > 0 {
		d.tailBytesCount = copy(d.tail[:], bz)
	}

	return total, nil
}

func (d digest32) Sum32() uint32 {
	h1 := d.h1

	// Process tail bytes when they exist. Blocks have 4 bytes, so depending
	// on the size of the data some bytes might not be processed after the last
	// write if the last block has less than 4 bytes.
	var k1 uint32
	switch d.tailBytesCount {
	case 3:
		k1 ^= uint32(d.tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(d.tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(d.tail[0])
		k1 *= c1
		k1 = bits.RotateLeft32(k1, r1)
		k1 *= c2
		h1 ^= k1
	}

	// Finalization mix
	h1 ^= uint32(d.writtenLen)

	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> r2
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}

func (d digest32) Sum(bz []byte) []byte {
	// Append hash in little-endian order
	s := d.Sum32()
	return append(bz, byte(s), byte(s>>8), byte(s>>16), byte(s>>24))
}

// processBlock reads a 4-byte little-endian block
// and mixes it into the hash state.
func (d *digest32) processBlock(bz []byte) {
	k1 := uint32(bz[0]) | uint32(bz[1])<<8 | uint32(bz[2])<<16 | uint32(bz[3])<<24
	k1 *= c1
	k1 = bits.RotateLeft32(k1, r1)
	k1 *= c2

	d.h1 ^= k1
	d.h1 = bits.RotateLeft32(d.h1, r2)
	d.h1 = d.h1*5 + 0xe6546b64 // MurmurHash3 "m" = 5, "n" = 0xe6546b64
}
