package testing

import (
	"crypto/sha256"
	"strings"
)

type coveredLine struct {
	coName string
	coLine int
}

type Coverage []coveredLine

type (
	internalHash uint64
	HashNumber   uint64
)

type hashMachine struct {
	Internal2Cov            map[internalHash]Coverage
	HashNumber2Internal     *[MaxCapacity]internalHash
	internalHash2hashNumber map[internalHash]HashNumber
	hashNumberCounter       *hashNumberCounter
}

func (hm *hashMachine) RegisterCoverage2HashNumber(coverage Coverage) HashNumber {
	internalHash := getInternalHash(coverage)
	hm.Internal2Cov[internalHash] = coverage
	hashNumber := hm.CountHashNumber(internalHash)
	hm.HashNumber2Internal[hashNumber] = internalHash
	return HashNumber(hashNumber)
}

func (hm *hashMachine) HashNumber2Coverage(hn uint) Coverage {
	internal := hm.HashNumber2Internal[hn]
	cov := hm.Internal2Cov[internal]
	return cov
}

type hashNumberCounter struct {
	counter HashNumber
}

func newHashNumberCounter(counter HashNumber) *hashNumberCounter {
	return &hashNumberCounter{
		counter: counter,
	}
}

func newHashMachine() *hashMachine {
	return &hashMachine{
		HashNumber2Internal:     &[MaxCapacity]internalHash{},
		hashNumberCounter:       newHashNumberCounter(0),
		Internal2Cov:            make(map[internalHash]Coverage),
		internalHash2hashNumber: make(map[internalHash]HashNumber),
	}
}

func (hm *hashMachine) CountHashNumber(ih internalHash) HashNumber {
	if value, exists := hm.internalHash2hashNumber[ih]; exists {
		// If the key exists, return the value
		return value
	}
	hm.internalHash2hashNumber[ih] = hm.hashNumberCounter.counter
	current := hm.hashNumberCounter.counter

	hm.hashNumberCounter.counter++
	return current
}

func coverageToBytes(coverage Coverage) []byte {
	var builder strings.Builder
	for _, line := range coverage {
		builder.WriteString(line.coName)
		builder.WriteString("|")
		builder.WriteString(intToString(line.coLine))
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
func getInternalHash(input Coverage) internalHash {
	valBytes := coverageToBytes(input)
	valArray := sha256.Sum256(valBytes)
	return internalHash(bytesToUint64(valArray))
}

// bytesToUint64 converts the first 8 bytes of a SHA256 hash to uint64
func bytesToUint64(b [32]byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}

func removeDuplicates(coverage Coverage) Coverage {
	uniqueMap := make(map[coveredLine]bool)
	result := Coverage{}

	for _, line := range coverage {
		if !uniqueMap[line] {
			uniqueMap[line] = true
			result = append(result, line)
		}
	}

	return result
}

func findDifferences(cov1, cov2 Coverage) Coverage {
	diff := Coverage{}
	for _, line1 := range cov1 {
		found := false
		for _, line2 := range cov2 {
			if line1.coName == line2.coName && line1.coLine == line2.coLine {
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
