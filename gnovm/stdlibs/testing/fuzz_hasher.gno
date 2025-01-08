package testing

import (
	"crypto/sha256"
	"strings"
)

type fuzzHasher struct {
	internal2Cov            map[internalHash]Coverage
	hashNumber2Internal     *[MaxCapacity]internalHash
	internalHash2hashNumber map[internalHash]hashNumber
	hashNumberCounter       *hashNumberCounter
}

type (
	// It is a type made assuming that a new Coverage type will be created and exported later
	Coverage    []coveredLine
	coveredLine struct {
		coName string
		coLine int
	}
)

type (
	internalHash      uint64
	hashNumber        uint64
	hashNumberCounter struct {
		counter hashNumber
	}
)

func (fh *fuzzHasher) registerCoverage2HashNumber(coverage Coverage) hashNumber {
	internalHash := getInternalHash(coverage)
	fh.internal2Cov[internalHash] = coverage
	hashnumber := fh.countHashNumber(internalHash)
	fh.hashNumber2Internal[hashnumber] = internalHash
	return hashNumber(hashnumber)
}

// calculates hash numbers without removing redundancy of line iterations in coverage
func getInternalHash(input Coverage) internalHash {
	valBytes := coverageToBytes(input)
	valArray := sha256.Sum256(valBytes)
	return internalHash(bytesToUint64(valArray))
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

func newHashNumberCounter(counter hashNumber) *hashNumberCounter {
	return &hashNumberCounter{
		counter: counter,
	}
}

func newFuzzHasher() *fuzzHasher {
	return &fuzzHasher{
		hashNumber2Internal:     &[MaxCapacity]internalHash{},
		hashNumberCounter:       newHashNumberCounter(0),
		internal2Cov:            make(map[internalHash]Coverage),
		internalHash2hashNumber: make(map[internalHash]hashNumber),
	}
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

// bytesToUint64 converts the first 8 bytes of a SHA256 hash to uint64
func bytesToUint64(b [32]byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}

func (fh *fuzzHasher) countHashNumber(ih internalHash) hashNumber {
	if value, exists := fh.internalHash2hashNumber[ih]; exists {
		// If the key exists, return the value
		return value
	}
	fh.internalHash2hashNumber[ih] = fh.hashNumberCounter.counter
	current := fh.hashNumberCounter.counter

	fh.hashNumberCounter.counter++
	return current
}

func (fh *fuzzHasher) hashNumber2Coverage(hn uint) Coverage {
	internal := fh.hashNumber2Internal[hn]
	cov := fh.internal2Cov[internal]
	return cov
}
