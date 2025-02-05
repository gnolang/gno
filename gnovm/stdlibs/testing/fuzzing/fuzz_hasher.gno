package fuzzing

import (
	"bytes"
	"strconv"
)

type (
	// InternalHash represents a value that primarily changes coverage based on hashing.
	InternalHash uint64

	// HashNumber represents the final value that coverage changes to
	// (based on the unique value from hashNumberCounter).
	HashNumber uint64

	// HashNumberCounter is a counter used to assign unique hashNumbers.
	HashNumberCounter struct {
		counter HashNumber
	}
)

type (
	// Coverage: array of coveredLine
	Coverage []CoveredLine

	// CoveredLine: single covered line
	CoveredLine struct {
		CoName string
		CoLine int
	}
)

// ----------------------------------------
// FuzzHasher: main struct
// ----------------------------------------
type FuzzHasher struct {
	HashNumber2Internal     *[MaxCapacity]InternalHash  // Table: hashNumber -> internalHash
	internal2Cov            map[InternalHash]Coverage   // Table: internalHash -> Coverage
	internalHash2hashNumber map[InternalHash]HashNumber // Table: internalHash -> hashNumber
	HashNumberCounter       *HashNumberCounter          // Counter for unique hashNumbers

	// coverageCache map[string]hashNumber  // coverage -> hashNumber
}

// ----------------------------------------
// registerCoverage2HashNumber
// ----------------------------------------
// 1) computes internalHash from coverage
// 2) inserts coverage into internal2Cov
// 3) obtains hashNumber (unique ID)
// 4) updates hashNumber2Internal
// 5) returns the assigned hashNumber
func (fh *FuzzHasher) RegisterCoverage2HashNumber(coverage Coverage) HashNumber {
	ih := GetInternalHash(coverage) // compute internalHash
	fh.internal2Cov[ih] = coverage

	hn := fh.countHashNumber(ih)
	fh.HashNumber2Internal[hn] = ih
	return hn
}

// ----------------------------------------
// GetInternalHash: coverage -> xxHash64
// ----------------------------------------
// replaces SHA256 usage; calls xxh64Sum (assume already implemented).
func GetInternalHash(cv Coverage) InternalHash {
	valBytes := CoverageToBytes(cv)
	// seed=0 is typical, but can be anything
	valUint := xxh64Sum(valBytes, 0)
	return InternalHash(valUint)
}

// ----------------------------------------
// CoverageToBytes: optimized approach
// ----------------------------------------
// uses bytes.Buffer instead of strings.Builder
func CoverageToBytes(cv Coverage) []byte {
	// each line -> coName length + ~few bytes

	var buf bytes.Buffer

	for _, line := range cv {
		// write coName as bytes
		buf.WriteString(line.CoName)
		// write separator
		buf.WriteByte('|')

		// convert coLine to decimal
		// => intToString or strconv.Itoa
		buf.WriteString(strconv.Itoa(line.CoLine))

		// separator
		buf.WriteByte('|')
	}
	return buf.Bytes()
}

func xxh64Sum(input []byte, seed uint64) uint64

// countHashNumber assigns a unique hashNumber to the given internalHash.
// If it exists, returns existing; otherwise create new.
// starting point is 0
func (fh *FuzzHasher) countHashNumber(ih InternalHash) HashNumber {
	if val, exists := fh.internalHash2hashNumber[ih]; exists {
		// if already known, return
		return val
	}
	fh.internalHash2hashNumber[ih] = fh.HashNumberCounter.counter
	current := fh.HashNumberCounter.counter
	fh.HashNumberCounter.counter++
	return current
}

// IsExistingCoverage는 주어진 coverage가 이미 등록되어 있는지 확인하여, 등록되어 있으면 true, 아니면 false를 반환합니다.
func (fh *FuzzHasher) IsExistingCoverage(cv Coverage) bool {
	ih := GetInternalHash(cv)
	_, exists := fh.internal2Cov[ih]
	return exists
}

// hashNumber2Coverage: given a hashNumber -> retrieve Coverage
func (fh *FuzzHasher) HashNumber2Coverage(hn uint) Coverage {
	ih := fh.HashNumber2Internal[hn]
	return fh.internal2Cov[ih]
}

// NewHashNumberCounter creates a new hashNumberCounter with given initial counter value.
func NewHashNumberCounter(counter HashNumber) *HashNumberCounter {
	return &HashNumberCounter{
		counter: counter,
	}
}

// NewFuzzHasher creates and initializes a new fuzzHasher instance.
func NewFuzzHasher() *FuzzHasher {
	return &FuzzHasher{
		HashNumber2Internal:     &[MaxCapacity]InternalHash{},
		internal2Cov:            make(map[InternalHash]Coverage),
		internalHash2hashNumber: make(map[InternalHash]HashNumber),
		HashNumberCounter:       NewHashNumberCounter(0),
	}
}
