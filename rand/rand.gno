package rand

import (
	"crypto/chacha20/rand"
)

// Realm giving access to a shared, insecure, predictable integer pseudo-random
// number generator (for casual games or dapps that don't want to initialize
// their own PRNG with their own random seed). Using this for non-frivolous
// purposes is a terrible idea.
//
// Note: underlying package does not support floats for now.

var prng *rand.RNG

func init() {
	prng = rand.NewFromSeed([]byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"))
}

// --- funcs ----------------------------------------------------

// Intn returns a uniform random int in [0,n). It panics if n <= 0.
func Intn(n int) int { return prng.Intn(n) }

// // Uint64n returns a uniform random uint64 in [0,n). It panics if n == 0.
// func Uint64n(n uint64) uint64 { return prng.Uint64n(n) }

// // Bytes is a helper function that allocates and returns n bytes of random data.
// func Bytes(n int) []byte { return prng.Bytes(n) }

// // Perm returns a random permutation of the integers [0,n). It panics if n < 0.
// func Perm(n int) []int { return prng.Perm(n) }

// // Shuffle randomly permutes n elements by repeatedly calling swap in the range
// // [0,n). It panics if n < 0.
// func Shuffle(n int, swap func(i, j int)) { prng.Shuffle(n, swap) }
