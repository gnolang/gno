package main

import (
	"crypto/sha256"
	"strings"
)

type CoveredLine struct {
	co_name string
	co_line int
}

type Coverage []CoveredLine

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
func GetPathHash(input Coverage) uint64 {
	input = RemoveDuplicates(input)
	valBytes := CoverageToBytes(input)
	valArray := sha256.Sum256(valBytes)
	return BytesToUint64(valArray)
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

func findDifferences(cov1, cov2 Coverage) Coverage {
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
