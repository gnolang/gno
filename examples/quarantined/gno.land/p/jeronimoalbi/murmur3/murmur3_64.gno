package murmur3

import "hash"

const hashSize64 = 8

type digest64 struct {
	d1 digest32
	d2 digest32
}

// New64 returns a new hash.Hash64 computing a 64-bit MurmurHash3
// hash using two 32-bit hashes with different seeds.
// It can generate 2 to the power of 64 unique 64-bit values.
// Hash is not collision safe, so it's recommended to handle potential
// collisions when dealing with more than 4_294_967_296 values which is
// when probability raises.
func New64() hash.Hash64 {
	return &digest64{
		d1: digest32{},
		d2: digest32{seed: 1, h1: 1},
	}
}

// NewWithSeed64 returns a new hash.Hash64 computing a 64-bit
// MurmurHash3 hash. The first 32-bit hash uses seed1 and the
// second uses seed2. The two seeds should ideally be different.
// It can generate 2 to the power of 64 unique 64-bit values.
// Hash is not collision safe, so it's recommended to handle potential
// collisions when dealing with more than 4_294_967_296 values which is
// when provability raises.
func NewWithSeed64(seed1, seed2 uint32) hash.Hash64 {
	return &digest64{
		d1: digest32{seed: seed1, h1: seed1},
		d2: digest32{seed: seed2, h1: seed2},
	}
}

// Sum64 returns the 64-bit MurmurHash3 hash of the specified data.
func Sum64(data []byte) uint64 {
	return Sum64WithSeed(data, 0, 1)
}

// Sum64WithSeed returns the 64-bit MurmurHash3 hash of the specified
// data using the given seeds.
func Sum64WithSeed(data []byte, seed1, seed2 uint32) uint64 {
	h := NewWithSeed64(seed1, seed2)
	h.Write(data)
	return h.Sum64()
}

func (digest64) BlockSize() int {
	// Use a 32-bit hash block size because writes are
	// done to two hash32 hashes using the same data.
	return blockSize32
}

func (digest64) Size() int {
	return hashSize64
}

func (d *digest64) Reset() {
	d.d1.Reset()
	d.d2.Reset()
}

func (d *digest64) Write(bz []byte) (int, error) {
	d.d1.Write(bz)
	return d.d2.Write(bz)
}

func (d digest64) Sum64() uint64 {
	return uint64(d.d1.Sum32())<<32 | uint64(d.d2.Sum32())
}

func (d digest64) Sum(bz []byte) []byte {
	s := d.Sum64()
	return append(bz,
		byte(s), byte(s>>8), byte(s>>16), byte(s>>24),
		byte(s>>32), byte(s>>40), byte(s>>48), byte(s>>56),
	)
}
