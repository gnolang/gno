package testing

import (
	"math"
	"math/rand"
)

// Internal state for the random number generator.
var x uint64 = 42

func ExUnixNano() int64 {
	return unixNano()
}

var (
	singleRand *rand.Rand
	isCalled   bool
)

// CustomSource는 Source 인터페이스 구현
type CustomSource struct {
	seed uint64
}

// Uint64는 pseudo-random 64-bit value를 리턴해야 함
func (cs *CustomSource) Uint64() uint64 {
	// 예: 간단한 XORShift-like 알고리즘
	cs.seed ^= cs.seed >> 12
	cs.seed ^= cs.seed << 25
	cs.seed ^= cs.seed >> 27
	return cs.seed * 2685821657736338717
}

// NewCustomSource는 커스텀 소스를 초기화하며 시드를 설정
func NewCustomSource(seed int64) *CustomSource {
	return &CustomSource{
		seed: uint64(seed),
	}
}

// 싱글톤 패턴 사용했습니다.
// GetGlobalRand: 단일 고루틴 환경이라 가정
func GetSingleRand() *rand.Rand {
	if !isCalled {
		isCalled = true
		seed := unixNano() // 동적 시드 설정
		source := NewCustomSource(seed)
		singleRand = rand.New(source)

	}
	return singleRand
}

// UniformRand generates a uniformly distributed random number.
// It uses the linear congrential generator method to produce the random number.
// and the result is in the range from 0 to m-1. here, m is 32768.
// To produce random number in [0, m-1], repeat this method as many times as needed.
// [1] https://en.wikipedia.org/wiki/Linear_congruential_generator
func UniformRand() uint64 {
	var a uint64 = 950213
	var c uint64 = 12345
	var m uint64 = 32768
	x = x*a + c
	return (x >> 16) % m
}

// _srand function sets the seed for the random number generator.
// This function provides an initial starting point for the sequence of random numbers.

// 고쳤습니다.
func _srand() {
	r := GetSingleRand()
	x = uint64(r.Uint64())
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

// 수정했습니다.
// 요구에 부합하도록 -1~1의 값 리턴하게 했습니다
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

// randRange generates a random integer between min and max (inclusive).
// TODO: 랜덤 질문. 이거 왜 인클루시브 처리했는지?
// 일단 나는 익스클루시브 했음음
// This function leverages the UniformRand function to generate a random number in a specified range.
// Note: max should be greater than min.
// 랜덤을 랜덤으로 수정했습니다.
func RandRange(start, end int64) int64 {
	if start >= end {
		panic("start >= end ")
	}
	r := GetSingleRand()
	rand_num := r.Int64()
	hashed_num := (rand_num) % (int64(end - start))
	result := int64(start) + hashed_num

	return result
}

func UintRandRange(start, end uint64) uint64 {
	if start >= end {
		panic("start >= end ")
	}
	r := GetSingleRand()
	rand_num := r.Uint64()
	hashed_num := (rand_num) % (uint64(end - start))
	result := uint64(start) + hashed_num

	return result
}

func RandInt64() int64 {
	r := GetSingleRand()
	rand_num := r.Int64()
	return rand_num
}

func RandUint64() uint64 {
	r := GetSingleRand()
	rand_num := r.Uint64()
	return rand_num
}

func RandUint32() uint32 {
	r := GetSingleRand()
	rand_num := r.Uint32()
	return rand_num
}

// 여기에 맞게 nrand고쳤습니다
func GenerateRandomBool(bias float64) bool {
	// Modify to use fuzz's random function for generating boolean with bias
	if bias < 0 || bias > 1 {
		panic("bias should be in the range [0, 1]")
	}
	// Convert fuzz's normalized range random float [-1, 1] to [0, 1]
	res := (nrand() + 1) / 2
	return res > bias
}

// 균등분포 추가했습니다.
func UniformRandomBool(probability float64) bool {
	if probability < 0.0 || probability > 1.0 {
		panic("Probability must be between 0.0 and 1.0")
	}
	r := GetSingleRand()
	return r.Float64() < probability
}
