package testing

import (
	"crypto/sha256"
	"strings"
)

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

func (hm *HashMachine) HashNumber2Coverage(hn uint) Coverage {
	internal := hm.HashNumber2Internal[hn]
	cov := hm.Internal2Cov_dict[internal]
	return cov
}

type hashNumberCounter struct {
	counter HashNumber
}

func New_HashNumberCounter(counter HashNumber) *hashNumberCounter {
	return &hashNumberCounter{
		counter: counter,
	}
}

func New_HashMachine() *HashMachine {
	return &HashMachine{
		HashNumber2Internal:     &[MAX_CAPACITY]internalHash{},
		hashNumber_counter:      New_HashNumberCounter(0),
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
		builder.WriteString(intToString(line.co_line))
		builder.WriteString("|")
	}
	return []byte(builder.String())
}

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

// calculates hash numbers without removing redundancy of line iterations in coverage
func GetInternalHash(input Coverage) internalHash {
	valBytes := CoverageToBytes(input)
	valArray := sha256.Sum256(valBytes)
	return internalHash(BytesToUint64(valArray))
}

// BytesToUint64 converts the first 8 bytes of a SHA256 hash to uint64
func BytesToUint64(b [32]byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}

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
