package testing

import (
	"math/rand"
)

// singleRand is the singleton instance of rand.Rand.
// isCalled indicates whether the singleton has been initialized.
var (
	singleRand *rand.Rand
	isCalled   bool
)

func UnixNano() int64 {
	return unixNano()
}

// CustomSource is a custom source for rand.Rand that uses a seed.
type CustomSource struct {
	seed uint64
}

func (cs *CustomSource) Uint64() uint64 {
	cs.seed ^= cs.seed >> 12
	cs.seed ^= cs.seed << 25
	cs.seed ^= cs.seed >> 27
	return cs.seed * 2685821657736338717
}

func NewCustomSource(seed int64) *CustomSource {
	return &CustomSource{
		seed: uint64(seed),
	}
}

// used a single tone pattern.
// GetGlobalRand: Assume to be a single high-routine environment
func GetSingleRand() *rand.Rand {
	if !isCalled {
		isCalled = true
		seed := unixNano()
		source := NewCustomSource(seed)
		singleRand = rand.New(source)

	}
	return singleRand
}

// nrand function generates a number approximating a normal distribution[1].
// It uses the Central Limit Theorem[2] by summing multiple uniformly distributed random numbers
// to approximate a normal distribution.
//
// y = Sum(k=1, K) (x_k - K/2) / sqrt(K/12)
//
// Here, K is some integer ans x_k are uniformly distributed numbers,
// even for K as small as 10, the approximation is quite good.
// [1] https://en.wikipedia.org/wiki/Normal_distribution
// [2] https://en.wikipedia.org/wiki/Central_limit_theorem
func nrand() float64 {
	r := GetSingleRand()
	danger := r.NormFloat64()
	scaled := danger / 3
	if scaled < -1 {
		return -1
	} else if scaled > 1 {
		return 1
	}
	return scaled
}

// This function leverages the GetSingleRand function to generate a random number in a specified range.
// Don't use it in a multi-threading environment
// Exclude end
// Note: max should be greater than min.
func RandRange(start, end int64) int64 {
	if start >= end {
		panic("start >= end ")
	}
	r := GetSingleRand()
	randNum := r.Int64()
	hashedNum := (randNum) % (int64(end - start))
	result := int64(start) + hashedNum

	return result
}

// This function leverages the GetSingleRand function to generate a random number in a specified range.
// Don't use it in a multi-threading environment
// Note: max should be greater than min.
func UintRandRange(start, end uint64) uint64 {
	if start >= end {
		panic("start >= end ")
	}
	r := GetSingleRand()
	randNum := r.Uint64()
	hashedNum := (randNum) % (uint64(end - start))
	result := uint64(start) + hashedNum

	return result
}

// This function leverages the GetSingleRand function to generate a random number in a specified range.
// Don't use it in a multi-threading environment
// Note: max should be greater than min.
func RandInt64() int64 {
	r := GetSingleRand()
	randNum := r.Int64()
	return randNum
}

// This function leverages the GetSingleRand function to generate a random number in a specified range.
// Don't use it in a multi-threading environment
// Note: max should be greater than min.
func RandUint64() uint64 {
	r := GetSingleRand()
	randNum := r.Uint64()
	return randNum
}

// This function leverages the GetSingleRand function to generate a random number in a specified range.
// Don't use it in a multi-threading environment
// Note: max should be greater than min.
func RandUint32() uint32 {
	r := GetSingleRand()
	randNum := r.Uint32()
	return randNum
}

// This function leverages the GetSingleRand function to generate a normal distributionbool in a specified range.
// Don't use it in a multi-threading environment
func GenerateRandomBool(bias float64) bool {
	// Modify to use fuzz's random function for generating boolean with bias
	if bias < 0 || bias > 1 {
		panic("bias should be in the range [0, 1]")
	}

	res := (nrand() + 1) / 2
	return res > bias
}

// This function leverages the GetSingleRand function to generate a random bool in a specified range.
// Don't use it in a multi-threading environment
func UniformRandomBool(probability float64) bool {
	if probability < 0.0 || probability > 1.0 {
		panic("Probability must be between 0.0 and 1.0")
	}
	r := GetSingleRand()
	return r.Float64() < probability
}
