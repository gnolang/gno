package testing

import (
	"crypto/sha256"
	"strings"
)

// fuzzHasher assigns numbers to coverage data during AFL fuzzing operations.
// It receives Coverage, converts it internally to internalHash, and assigns each hashNumber to each hash.
// It also manages internal tables that allow conversion between Coverage, internalHash, and hashNumber.
type fuzzHasher struct {
	hashNumber2Internal     *[MaxCapacity]internalHash  // Table to retrieve internalHash from hashNumber
	internal2Cov            map[internalHash]Coverage   // Table to retrieve Coverage from internalHash
	internalHash2hashNumber map[internalHash]hashNumber // Table to retrieve hashNumber from internalHash
	hashNumberCounter       *hashNumberCounter          // Counter for assigning unique hashNumbers
}

type (
	// internalHash represents a value that primarily changes coverage based on SHA256 hashing.
	internalHash uint64 // The value that coverage changes based on SHA256 hash.

	// hashNumber represents the final value that coverage changes to, based on the unique value from hashNumberCounter.
	hashNumber uint64 // The final value coverage changes to, based on the unique value from hashNumberCounter.

	// hashNumberCounter is a counter used to assign unique hashNumbers.
	hashNumberCounter struct {
		counter hashNumber
	}
)

type (
	// Coverage represents coverage data. It is assumed that a new Coverage type will be created and exported later.
	// The Coverage type may change depending on how coverage is implemented in the future.
	// Currently, it is created assuming coverage is similar to Python's coverage type.
	Coverage []coveredLine

	// coveredLine represents a single covered line, similar to Python debugger's coverage type.
	coveredLine struct {
		coName string
		coLine int
	}
)

// registerCoverage2HashNumber registers the given Coverage in the fuzzHasher and returns its assigned hashNumber.
// Coverage is assigned a unique number starting from 0 based on the order it is input into the fuzzHasher.
func (fh *fuzzHasher) registerCoverage2HashNumber(coverage Coverage) hashNumber {
	internalHash := getInternalHash(coverage)
	fh.internal2Cov[internalHash] = coverage
	hashnumber := fh.countHashNumber(internalHash)
	fh.hashNumber2Internal[hashnumber] = internalHash
	return hashNumber(hashnumber)
}

// getInternalHash calculates the internalHash for the given Coverage.
// It converts the Coverage to bytes, computes its SHA256 hash, and converts the first 8 bytes to a uint64.
func getInternalHash(input Coverage) internalHash {
	valBytes := coverageToBytes(input)
	valArray := sha256.Sum256(valBytes)
	return internalHash(bytesToUint64(valArray))
}

// coverageToBytes converts Coverage data to a byte slice by concatenating coName and coLine of each coveredLine.
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

// newHashNumberCounter creates a new hashNumberCounter with the given initial counter value.
func newHashNumberCounter(counter hashNumber) *hashNumberCounter {
	return &hashNumberCounter{
		counter: counter,
	}
}

// newFuzzHasher creates and initializes a new fuzzHasher instance.
func newFuzzHasher() *fuzzHasher {
	return &fuzzHasher{
		hashNumber2Internal:     &[MaxCapacity]internalHash{},
		hashNumberCounter:       newHashNumberCounter(0),
		internal2Cov:            make(map[internalHash]Coverage),
		internalHash2hashNumber: make(map[internalHash]hashNumber),
	}
}

// intToString converts an integer to its string representation.
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

// bytesToUint64 converts the first 8 bytes of a SHA256 hash to a uint64.
func bytesToUint64(b [32]byte) uint64 {
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
}

// countHashNumber assigns a unique hashNumber to the given internalHash.
// If the internalHash already exists, it returns the existing hashNumber.
// Otherwise, it assigns a new hashNumber using the hashNumberCounter.
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

// hashNumber2Coverage converts a hashNumber to its corresponding Coverage using the tables in fuzzHasher.
func (fh *fuzzHasher) hashNumber2Coverage(hn uint) Coverage {
	internal := fh.hashNumber2Internal[hn]
	cov := fh.internal2Cov[internal]
	return cov
}
