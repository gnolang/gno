package testing

import (
	"crypto/sha256"
	"strings"
)

// func main() {
// 	c1 := Coverage{
// 		{"a", 1},
// 		{"a", 2},
// 		{"a", 3},
// 	}
// 	c2 := Coverage{
// 		{"a", 1},
// 		{"a", 1},
// 		{"a", 1},
// 		{"a", 2},
// 		{"a", 3},
// 	}
// 	c3 := Coverage{
// 		{"a", 1},
// 		{"a", 3},
// 	}
// 	c4 := Coverage{
// 		{"a", 1},
// 		{"a", 2},
// 		{"a", 3},
// 		{"a", 3},
// 	}
// 	c5 := Coverage{
// 		{"a", 4},
// 		{"a", 2},
// 		{"a", 3},
// 		{"a", 3},
// 	}

// 	hm := New_HashMachine()

// 	h1 := hm.RegisterCoverage2HashNumber(c1)
// 	h2 := hm.RegisterCoverage2HashNumber(c2)
// 	h3 := hm.RegisterCoverage2HashNumber(c3)
// 	h4 := hm.RegisterCoverage2HashNumber(c4)
// 	h5 := hm.RegisterCoverage2HashNumber(c5)
// 	println(h1, h2, h3, h4, h5)
// 	println(hm.HashNumber2Internal[0])
// 	println(hm.HashNumber2Internal[100])
// 	println(hm.Internal2Cov_dict)
// }

type CoveredLine struct {
	co_name string
	co_line int
}

type Coverage []CoveredLine

type (
	internalHash uint64
	HashNumber   uint64
)

type HashMachine struct {
	Internal2Cov_dict       map[internalHash]Coverage
	HashNumber2Internal     *[MAX_CAPACITY]internalHash
	internalHash2hashNumber map[internalHash]HashNumber
	hashNumber_counter      *hashNumberCounter
}

func (hm *HashMachine) RegisterCoverage2HashNumber(coverage Coverage) HashNumber {
	internal_hash := GetInternalHash(coverage)
	hm.Internal2Cov_dict[internal_hash] = coverage
	hashNumber := hm.Count_HashNumber(internal_hash)
	hm.HashNumber2Internal[hashNumber] = internal_hash
	return HashNumber(hashNumber)
}

type hashNumberCounter struct {
	counter HashNumber
}

func New_HashNumberCounter() *hashNumberCounter {
	return &hashNumberCounter{
		counter: 0,
	}
}

func New_HashMachine() *HashMachine {
	return &HashMachine{
		HashNumber2Internal:     &[MAX_CAPACITY]internalHash{},
		hashNumber_counter:      New_HashNumberCounter(),
		Internal2Cov_dict:       make(map[internalHash]Coverage),
		internalHash2hashNumber: make(map[internalHash]HashNumber),
	}
}

func (hm *HashMachine) Count_HashNumber(ih internalHash) HashNumber {
	if value, exists := hm.internalHash2hashNumber[ih]; exists {
		// If the key exists, return the value
		return value
	}
	hm.internalHash2hashNumber[ih] = hm.hashNumber_counter.counter
	current := hm.hashNumber_counter.counter

	hm.hashNumber_counter.counter++
	return current
}

func CoverageToBytes(coverage Coverage) []byte {
	var builder strings.Builder
	for _, line := range coverage {
		builder.WriteString(line.co_name)
		builder.WriteString("|")
		builder.WriteString(intToString(line.co_line)) // 숫자 변환 함수 사용
		builder.WriteString("|")
	}
	return []byte(builder.String())
}

// int를 string으로 변환하는 함수
func intToString(n int) string {
	if n == 0 {
		return "0"
	}

	isNegative := false
	if n < 0 {
		isNegative = true
		n = -n
	}

	var digits []byte
	for n > 0 {
		digit := n % 10
		digits = append([]byte{'0' + byte(digit)}, digits...)
		n /= 10
	}

	if isNegative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// GetPathHash calculates the hash of a Coverage type
// 중복 제거도 추가함.
// 타입 맞게 수정 추가함. 그래도 기반 기능은 동일.
func GetInternalHash(input Coverage) internalHash {
	// 다시 생각해보니 좀 차이가 있긴 한데...
	// 차이가 있긴 하네
	// 중복 제거는 안하는 걸로 (평탄 확률이 아닐 시, 중복 놔두는게 더 정교)
	// TODO: 다만 중복 제거는 이후 오염 체킹 등에 쓰기!!!
	// input = RemoveDuplicates(input)
	valBytes := CoverageToBytes(input)
	valArray := sha256.Sum256(valBytes)
	return internalHash(BytesToUint64(valArray))
}

// BytesToUint64 converts the first 8 bytes of a SHA256 hash to uint64
func BytesToUint64(b [32]byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}

// 중복 제거 함수
func RemoveDuplicates(coverage Coverage) Coverage {
	uniqueMap := make(map[CoveredLine]bool)
	result := Coverage{}

	for _, line := range coverage {
		if !uniqueMap[line] {
			uniqueMap[line] = true
			result = append(result, line)
		}
	}

	return result
}

// func main() {
// 	coverage1 := Coverage{
// 		{"closure", 13},
// 		{"closure", 14},
// 		{"Reverse1", 37},
// 		{"Reverse1", 38},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 42},
// 		{"closure", 15},
// 		{"Reverse1", 37},
// 		{"Reverse1", 38},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 42},
// 		{"closure", 16},
// 		{"closure", 18},
// 		{"closure", 19},
// 		{"closure", 21},
// 		{"closure", 22},
// 	}

// 	coverage2 := Coverage{
// 		{"closure", 13},
// 		{"closure", 14},
// 		{"Reverse1", 37},
// 		{"Reverse1", 38},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 42},
// 		{"closure", 15},
// 		{"Reverse1", 37},
// 		{"Reverse1", 38},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 42},
// 		{"closure", 16},
// 		{"closure", 17},
// 	}

// 	eq_to_coverage1 := Coverage{
// 		// covere1과 같음. 그러나 중복 라인 추가.
// 		// 이 경우 해시값이 같아야 함.
// 		{"closure", 13},
// 		{"closure", 13},
// 		{"closure", 13},
// 		{"closure", 13},
// 		{"closure", 13},
// 		{"closure", 14},
// 		{"Reverse1", 37},
// 		{"Reverse1", 38},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 42},
// 		{"closure", 15},
// 		{"Reverse1", 37},
// 		{"Reverse1", 38},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 42},
// 		{"closure", 16},
// 		{"closure", 18},
// 		{"closure", 19},
// 		{"closure", 21},
// 		{"closure", 22},
// 	}

// 	similar_to_coverage1 := Coverage{
// 		// coverage1과 유사하지만, 결국 그 해시 값은 차이가 나야 함.
// 		{"closure", 13},
// 		{"closure", 14},
// 		{"Reverse2", 37},
// 		{"Reverse1", 38},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 42},
// 		{"closure", 15},
// 		{"Reverse1", 37},
// 		{"Reverse1", 38},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 40},
// 		{"Reverse1", 41},
// 		{"Reverse1", 39},
// 		{"Reverse1", 42},
// 		{"closure", 16},
// 		{"closure", 18},
// 		{"closure", 19},
// 		{"closure", 21},
// 		{"closure", 22},
// 	}

// 	hash1 := GetPathHash(coverage1)
// 	hash2 := GetPathHash(coverage2)
// 	e1_to_hash1 := GetPathHash(eq_to_coverage1)
// 	sim_to_hash1 := GetPathHash(similar_to_coverage1)

// 	println("Hash for Coverage1: %x\n", hash1)

// 	println("Hash for Coverage2: %x\n", hash2)
// 	println("Hash for e1_to_cov1: %x\n", e1_to_hash1)
// 	println("Hash for sim_to_cov1: %x\n", sim_to_hash1)
// }

func FindDifferences(cov1, cov2 Coverage) Coverage {
	diff := Coverage{}
	for _, line1 := range cov1 {
		found := false
		for _, line2 := range cov2 {
			if line1.co_name == line2.co_name && line1.co_line == line2.co_line {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, line1)
		}
	}
	return diff
}
